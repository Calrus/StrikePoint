package newsfeed

import (
	"fmt"
	"log"
	"strikelogic/news_engine"
	"strikelogic/storage"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

func FetchStockNews(ticker string) {
	fp := gofeed.NewParser()
	url := fmt.Sprintf("https://news.google.com/rss/search?q=%s+stock+news&hl=en-US&gl=US&ceid=US:en", ticker)
	feed, err := fp.ParseURL(url)
	if err != nil {
		log.Printf("Error fetching feed for %s: %v\n", ticker, err)
		return
	}

	for _, item := range feed.Items {
		title := item.Title
		// Clean title: remove " - SourceName"
		if idx := strings.LastIndex(title, " - "); idx != -1 {
			title = title[:idx]
		}

		publishedAt := time.Now()
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		}

		// Analyze sentiment
		signals, err := news_engine.AnalyzeSentiment(title)
		if err != nil || len(signals) == 0 {
			log.Printf("Error analyzing sentiment for '%s': %v", title, err)
			// Save with default/neutral sentiment for the requested ticker
			signals = []news_engine.Signal{{
				Ticker:     ticker,
				Sentiment:  "NEUTRAL",
				Confidence: 0,
				Reasoning:  "Analysis failed",
			}}
		}

		for _, signal := range signals {
			// Use the ticker identified by LLM, or fallback to the search ticker
			articleTicker := signal.Ticker
			if articleTicker == "" {
				articleTicker = ticker
			}

			article := storage.Article{
				Ticker:         articleTicker,
				Title:          title,
				Link:           item.Link,
				PublishedAt:    publishedAt,
				Sentiment:      signal.Sentiment,
				Confidence:     signal.Confidence,
				Reasoning:      signal.Reasoning,
				SentimentScore: 0,
			}

			if article.Sentiment == "BULLISH" {
				article.SentimentScore = article.Confidence
			} else if article.Sentiment == "BEARISH" {
				article.SentimentScore = -article.Confidence
			}

			storage.SaveArticle(article)
		}
	}
}
