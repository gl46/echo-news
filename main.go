package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
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

CREATE TABLE IF NOT EXISTS translations (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	article_id INTEGER NOT NULL UNIQUE,
	target_lang TEXT NOT NULL DEFAULT 'zh-CN',
	translation TEXT NOT NULL,
	model TEXT,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY(article_id) REFERENCES articles(id)
);
`

const defaultTargetLang = "zh-CN"

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

type TranslateResponse struct {
	ID          string `json:"id"`
	Translation string `json:"translation"`
}

type articleForTranslation struct {
	ID      int64
	Title   string
	Summary string
}

type App struct {
	db         *sql.DB
	translator *Translator
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
	SELECT
		a.id,
		a.title,
		a.description,
		a.link,
		a.published_at,
		a.source,
		EXISTS (
			SELECT 1 FROM translations t
			WHERE t.article_id = a.id AND t.target_lang = ?
		)
	FROM articles a
	ORDER BY a.published_at DESC, a.id DESC;
	`

	rows, err := db.Query(query, defaultTargetLang)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := make([]ArticleResponse, 0)
	for rows.Next() {
		var article ArticleResponse
		var articleID int64
		var summary sql.NullString
		var hasTranslation int
		if err := rows.Scan(
			&articleID,
			&article.Title,
			&summary,
			&article.URL,
			&article.PublishedAt,
			&article.Source,
			&hasTranslation,
		); err != nil {
			return nil, err
		}

		article.ID = strconv.FormatInt(articleID, 10)
		article.Summary = strings.TrimSpace(summary.String)
		article.HasTranslation = hasTranslation != 0
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

func (a *App) translateHandler(w http.ResponseWriter, r *http.Request) {
	articleID, err := parseArticleID(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid article id", http.StatusBadRequest)
		return
	}

	cached, ok, err := cachedTranslation(a.db, articleID)
	if err != nil {
		http.Error(w, "failed to load translation", http.StatusInternalServerError)
		return
	}
	if ok {
		writeJSON(w, http.StatusOK, TranslateResponse{
			ID:          strconv.FormatInt(articleID, 10),
			Translation: cached,
		})
		return
	}

	article, err := loadArticleForTranslation(a.db, articleID)
	if err == sql.ErrNoRows {
		http.Error(w, "article not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to load article", http.StatusInternalServerError)
		return
	}

	if a.translator == nil {
		http.Error(w, "translation service is not configured", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	translation, err := a.translator.TranslateArticle(ctx, article.Title, article.Summary)
	if err != nil {
		http.Error(w, "translation failed", http.StatusBadGateway)
		return
	}

	if err := saveTranslation(a.db, article.ID, defaultTargetLang, translation, a.translator.Model()); err != nil {
		http.Error(w, "failed to save translation", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, TranslateResponse{
		ID:          strconv.FormatInt(articleID, 10),
		Translation: translation,
	})
}

func parseArticleID(value string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid article id %q", value)
	}

	return id, nil
}

func cachedTranslation(db *sql.DB, articleID int64) (string, bool, error) {
	var translation string
	err := db.QueryRow(
		`SELECT translation FROM translations WHERE article_id = ? AND target_lang = ?`,
		articleID,
		defaultTargetLang,
	).Scan(&translation)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}

	return translation, true, nil
}

func loadArticleForTranslation(db *sql.DB, articleID int64) (articleForTranslation, error) {
	var article articleForTranslation
	var summary sql.NullString
	err := db.QueryRow(
		`SELECT id, title, description FROM articles WHERE id = ?`,
		articleID,
	).Scan(&article.ID, &article.Title, &summary)
	if err != nil {
		return articleForTranslation{}, err
	}

	article.Summary = strings.TrimSpace(summary.String)
	if article.Summary == "" {
		article.Summary = article.Title
	}

	return article, nil
}

func saveTranslation(db *sql.DB, articleID int64, targetLang, translation, model string) error {
	const query = `
	INSERT INTO translations (article_id, target_lang, translation, model)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(article_id) DO UPDATE SET
		target_lang = excluded.target_lang,
		translation = excluded.translation,
		model = excluded.model,
		updated_at = CURRENT_TIMESTAMP;
	`

	_, err := db.Exec(query, articleID, targetLang, translation, model)
	return err
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
	loadEnvFile()

	dbPath := defaultDBPath
	if envPath := strings.TrimSpace(os.Getenv("DATABASE_URL")); envPath != "" {
		dbPath = envPath
	}

	port := defaultPort
	if envPort := strings.TrimSpace(os.Getenv("PORT")); envPort != "" {
		port = envPort
	}

	staticDir := strings.TrimSpace(os.Getenv("STATIC_DIR"))
	if staticDir == "" {
		staticDir = "./echo-news-fronted/dist"
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

	app := &App{
		db:         db,
		translator: NewTranslatorFromEnv(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/articles", app.articleHandler)
	mux.HandleFunc("POST /api/translate/{id}", app.translateHandler)

	// Serve frontend static files if the directory exists
	if info, err := os.Stat(staticDir); err == nil && info.IsDir() {
		fs := http.FileServer(http.Dir(staticDir))
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			// Try serving the file directly first
			path := staticDir + r.URL.Path
			if _, err := os.Stat(path); err == nil {
				fs.ServeHTTP(w, r)
				return
			}
			// Fall back to index.html for SPA routing
			http.ServeFile(w, r, staticDir+"/index.html")
		})
		fmt.Printf("serving static files from %s\n", staticDir)
	}

	fmt.Printf("server listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Println("server error:", err)
		os.Exit(1)
	}
}

func loadEnvFile() {
	_ = godotenv.Load()
}
