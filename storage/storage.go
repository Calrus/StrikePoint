package storage

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

type Article struct {
	ID             int       `json:"id"`
	Ticker         string    `json:"ticker"`
	Title          string    `json:"title"`
	Link           string    `json:"link"`
	PublishedAt    time.Time `json:"published_at"`
	SentimentScore float64   `json:"sentiment_score"`
	Sentiment      string    `json:"sentiment"`
	Confidence     float64   `json:"confidence"`
	Reasoning      string    `json:"reasoning"`
}

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./strikelogic.db")
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS news_articles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ticker TEXT,
		title TEXT,
		link TEXT UNIQUE,
		published_at DATETIME,
		sentiment_score REAL DEFAULT 0,
		sentiment TEXT,
		confidence REAL DEFAULT 0,
		reasoning TEXT
	);`

	_, err = DB.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Migrations for existing tables
	migrations := []string{
		"ALTER TABLE news_articles ADD COLUMN sentiment TEXT;",
		"ALTER TABLE news_articles ADD COLUMN confidence REAL DEFAULT 0;",
		"ALTER TABLE news_articles ADD COLUMN reasoning TEXT;",
	}

	for _, query := range migrations {
		DB.Exec(query) // Ignore errors if columns already exist
	}
}

func SaveArticle(article Article) error {
	insertSQL := `INSERT OR IGNORE INTO news_articles (ticker, title, link, published_at, sentiment_score, sentiment, confidence, reasoning) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := DB.Exec(insertSQL, article.Ticker, article.Title, article.Link, article.PublishedAt, article.SentimentScore, article.Sentiment, article.Confidence, article.Reasoning)
	return err
}

func GetLatestNews(ticker string, limit int) ([]Article, error) {
	querySQL := `SELECT id, ticker, title, link, published_at, sentiment_score, COALESCE(sentiment, ''), COALESCE(confidence, 0), COALESCE(reasoning, '') FROM news_articles WHERE ticker = ? ORDER BY published_at DESC LIMIT ?`
	rows, err := DB.Query(querySQL, ticker, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var a Article
		// Handle NULLs for new fields if necessary, but Scan should handle empty strings/zeros if DB has defaults or we use sql.NullString.
		// For simplicity, we assume the driver handles mapping to string/float64 zero values if NULL, or we ensure we don't insert NULLs.
		// Actually sqlite3 driver might error on NULL -> string.
		// Let's use sql.NullString if we were strict, but let's try direct scan first.
		// To be safe, we can use COALESCE in query or just ensure we write empty strings.
		err = rows.Scan(&a.ID, &a.Ticker, &a.Title, &a.Link, &a.PublishedAt, &a.SentimentScore, &a.Sentiment, &a.Confidence, &a.Reasoning)
		if err != nil {
			// If scan fails due to NULLs (for old records), we might need to handle it.
			// Let's update the query to handle NULLs.
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, nil
}
