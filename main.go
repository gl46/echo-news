package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// RSS 结构体，后续准备迁移到数据库，先用map做替代
var fetch_url = map[string]string{
	"bbc": "https://feeds.bbci.co.uk/news/rss.xml",
}

// RSS 结构体
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

// 返回前端的文章结构体
type ArticleResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	PubDate     string `json:"pub_date"`
}

// 获取RSS原始文件
func fetch_rss(url string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
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

// 解析RSS数据
func parse_rss(data string) (*RSS, error) {
	var rss RSS
	err := xml.Unmarshal([]byte(data), &rss)
	if err != nil {
		return nil, err
	}
	return &rss, nil
}

// HTTP handler函数，返回前端需要的json

func article_handler(w http.ResponseWriter, r *http.Request) {
	// 获取RSS数据
	xmlData, err := fetch_rss(fetch_url["bbc"])
	if err != nil {
		http.Error(w, "failed to fetch RSS feed", http.StatusInternalServerError)
		return
	}
	rss, err := parse_rss(xmlData)
	if err != nil {
		http.Error(w, "failed to parse RSS feed", http.StatusInternalServerError)
		return
	}
	// 将RSS数据转换为json
	articles := make([]ArticleResponse, 0, len(rss.Channel.Items))
	for _, item := range rss.Channel.Items {
		articles = append(articles, ArticleResponse{
			ID:          item.GUID,
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
			PubDate:     item.PubDate,
		})
	}
	// 设置响应头
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]any{"items": articles})
}

// debug用的main函数，后续会删除

/* func main() {
	xmlData, err := fetch_rss(fetch_url["bbc"])
	if err != nil {
		fmt.Printf("Error fetching RSS feed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Fetched RSS feed:\n%s\n", xmlData)
	rss, err := parse_rss(xmlData)
	if err != nil {
		fmt.Printf("Error parsing RSS feed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Parsed RSS feed:\n%+v\n", rss)
	fmt.Println("channel title:", rss.Channel.Title)
    fmt.Println("article count:", len(rss.Channel.Items))

    if len(rss.Channel.Items) > 0 {
    	fmt.Println("first title:", rss.Channel.Items[0].Title)
		fmt.Println("first link:", rss.Channel.Items[0].Link)
    }

} */

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/articles", article_handler)

	fmt.Println("server listening on :8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println("server error:", err)
		os.Exit(1)
	}
}
