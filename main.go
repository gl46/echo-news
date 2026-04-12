package main

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	defaultDBPath = "./echo-news.db"
	defaultPort   = "8080"
)

var fetchURL = map[string]string{
	"bbc": "https://feeds.bbci.co.uk/news/rss.xml",
}

const articleSchema = `
CREATE TABLE IF NOT EXISTS articles (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	source TEXT NOT NULL,
	guid TEXT NOT NULL UNIQUE,
	title TEXT NOT NULL,
	description TEXT,
	link TEXT NOT NULL,
	published_at TEXT NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title         string `xml:"title"`
	Description   string `xml:"description"`
	Link          string `xml:"link"`
	LastBuildDate string `xml:"lastBuildDate"`
	Language      string `xml:"language"`
	TTL           string `xml:"ttl"`
	Items         []Item `xml:"item"`
}

type Item struct {
	Title       string         `xml:"title"`
	Description string         `xml:"description"`
	Link        string         `xml:"link"`
	GUID        string         `xml:"guid"`
	PubDate     string         `xml:"pubDate"`
	Thumbnail   MediaThumbnail `xml:"thumbnail"`
}

type MediaThumbnail struct {
	URL    string `xml:"url,attr"`
	Width  string `xml:"width,attr"`
	Height string `xml:"height,attr"`
}

type ArticleResponse struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Summary        string `json:"summary"`
	URL            string `json:"url"`
	PublishedAt    string `json:"published_at"`
	Source         string `json:"source"`
	HasTranslation bool   `json:"has_translation"`
}

type App struct {
	db *sql.DB
}

func fetchRSS(url string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "EchoNews/0.1 (+https://github.com/gl46/echo-news)")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func parseRSS(data string) (*RSS, error) {
	var rss RSS
	if err := xml.Unmarshal([]byte(data), &rss); err != nil {
		return nil, err
	}

	return &rss, nil
}

func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func initSchema(db *sql.DB) error {
	_, err := db.Exec(articleSchema)
	return err
}

func syncArticles(db *sql.DB) error {
	xmlData, err := fetchRSS(fetchURL["bbc"])
	if err != nil {
		return err
	}

	rss, err := parseRSS(xmlData)
	if err != nil {
		return err
	}

	const upsertArticleSQL = `
	INSERT INTO articles (source, guid, title, description, link, published_at)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(guid) DO UPDATE SET
		title = excluded.title,
		description = excluded.description,
		link = excluded.link,
		published_at = excluded.published_at,
		updated_at = CURRENT_TIMESTAMP;
	`

	for _, item := range rss.Channel.Items {
		guid := articleID(item)
		publishedAt := normalizePublishedAt(item.PubDate)

		if _, err := db.Exec(
			upsertArticleSQL,
			"BBC",
			guid,
			strings.TrimSpace(item.Title),
			strings.TrimSpace(item.Description),
			strings.TrimSpace(item.Link),
			publishedAt,
		); err != nil {
			return err
		}
	}

	return nil
}

func listArticles(db *sql.DB) ([]ArticleResponse, error) {
	const query = `
	SELECT guid, title, description, link, published_at, source
	FROM articles
	ORDER BY published_at DESC, id DESC;
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := make([]ArticleResponse, 0)
	for rows.Next() {
		var article ArticleResponse
		if err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Summary,
			&article.URL,
			&article.PublishedAt,
			&article.Source,
		); err != nil {
			return nil, err
		}

		article.HasTranslation = false
		if article.Summary == "" {
			article.Summary = article.Title
		}

		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

func (a *App) articleHandler(w http.ResponseWriter, r *http.Request) {
	articles, err := listArticles(a.db)
	if err != nil {
		http.Error(w, "failed to load articles", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, articles)
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func articleID(item Item) string {
	if guid := strings.TrimSpace(item.GUID); guid != "" {
		return guid
	}

	if link := strings.TrimSpace(item.Link); link != "" {
		return link
	}

	return strings.TrimSpace(item.Title)
}

func normalizePublishedAt(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Now().UTC().Format(time.RFC3339)
	}

	layouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC850,
		time.ANSIC,
	}

	for _, layout := range layouts {
		parsedTime, err := time.Parse(layout, value)
		if err == nil {
			return parsedTime.UTC().Format(time.RFC3339)
		}
	}

	return value
}

func main() {
	dbPath := defaultDBPath
	if envPath := strings.TrimSpace(os.Getenv("DATABASE_URL")); envPath != "" {
		dbPath = envPath
	}

	port := defaultPort
	if envPort := strings.TrimSpace(os.Getenv("PORT")); envPort != "" {
		port = envPort
	}

	db, err := openDB(dbPath)
	if err != nil {
		fmt.Println("failed to open database:", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := initSchema(db); err != nil {
		fmt.Println("failed to initialize schema:", err)
		os.Exit(1)
	}

	if err := syncArticles(db); err != nil {
		fmt.Println("warning: failed to sync RSS feed:", err)
	}

	app := &App{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/articles", app.articleHandler)

	fmt.Printf("server listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Println("server error:", err)
		os.Exit(1)
	}
}
