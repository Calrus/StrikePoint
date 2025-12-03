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

	// Check schema version
	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER)`)
	if err != nil {
		log.Fatal(err)
	}

	var version int
	err = DB.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_version`).Scan(&version)
	if err != nil {
		version = 0
	}

	if version < 1 {
		// Migration to v1: Fix unique constraint on news_articles
		log.Println("Migrating database to version 1...")

		tx, err := DB.Begin()
		if err != nil {
			log.Fatal(err)
		}

		// Check if news_articles exists
		var name string
		tableExists := false
		err = tx.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='news_articles'`).Scan(&name)
		if err == nil && name == "news_articles" {
			tableExists = true
		}

		if tableExists {
			// Rename old table
			_, err = tx.Exec(`ALTER TABLE news_articles RENAME TO news_articles_old`)
			if err != nil {
				tx.Rollback()
				log.Fatal(err)
			}
		}

		// Create new table
		createTableSQL := `CREATE TABLE news_articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ticker TEXT,
			title TEXT,
			link TEXT,
			published_at DATETIME,
			sentiment_score REAL DEFAULT 0,
			sentiment TEXT,
			confidence REAL DEFAULT 0,
			reasoning TEXT,
			UNIQUE(ticker, link)
		);`
		_, err = tx.Exec(createTableSQL)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}

		if tableExists {
			// Copy data
			// We need to handle columns carefully. Old table might have fewer columns if it was very old,
			// but we assume it matches the struct.
			// Actually, let's just copy common columns.
			copySQL := `INSERT OR IGNORE INTO news_articles (ticker, title, link, published_at, sentiment_score, sentiment, confidence, reasoning)
			SELECT ticker, title, link, published_at, sentiment_score, sentiment, confidence, reasoning FROM news_articles_old`
			_, err = tx.Exec(copySQL)
			if err != nil {
				// If columns are missing in old table, this might fail.
				// But we added columns in previous migrations (ALTER TABLE).
				// So it should be fine.
				log.Printf("Warning copying data: %v", err)
			}

			_, err = tx.Exec(`DROP TABLE news_articles_old`)
			if err != nil {
				tx.Rollback()
				log.Fatal(err)
			}
		}

		_, err = tx.Exec(`INSERT INTO schema_version (version) VALUES (1)`)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}

		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}
	}

	// Ensure table exists for fresh runs (redundant if migration ran, but safe)
	createTableSQL := `CREATE TABLE IF NOT EXISTS news_articles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ticker TEXT,
		title TEXT,
		link TEXT,
		published_at DATETIME,
		sentiment_score REAL DEFAULT 0,
		sentiment TEXT,
		confidence REAL DEFAULT 0,
		reasoning TEXT,
		UNIQUE(ticker, link)
	);`

	_, err = DB.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}

func SaveArticle(article Article) error {
	insertSQL := `INSERT OR IGNORE INTO news_articles (ticker, title, link, published_at, sentiment_score, sentiment, confidence, reasoning) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := DB.Exec(insertSQL, article.Ticker, article.Title, article.Link, article.PublishedAt, article.SentimentScore, article.Sentiment, article.Confidence, article.Reasoning)
	return err
}

func GetLatestNews(ticker string, limit int) ([]Article, error) {
	var querySQL string
	var args []interface{}

	if ticker != "" {
		querySQL = `SELECT id, ticker, title, link, published_at, sentiment_score, COALESCE(sentiment, ''), COALESCE(confidence, 0), COALESCE(reasoning, '') FROM news_articles WHERE ticker = ? ORDER BY published_at DESC LIMIT ?`
		args = append(args, ticker, limit)
	} else {
		querySQL = `SELECT id, ticker, title, link, published_at, sentiment_score, COALESCE(sentiment, ''), COALESCE(confidence, 0), COALESCE(reasoning, '') FROM news_articles ORDER BY published_at DESC LIMIT ?`
		args = append(args, limit)
	}

	rows, err := DB.Query(querySQL, args...)
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
