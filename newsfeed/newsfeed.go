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
		signal, err := news_engine.AnalyzeSentiment(title)
		if err != nil {
			log.Printf("Error analyzing sentiment for '%s': %v", title, err)
			// Continue without sentiment or skip? Let's save without sentiment for now, or just default.
			// Actually, let's just log and save with defaults.
			signal = news_engine.Signal{
				Sentiment:  "NEUTRAL",
				Confidence: 0,
				Reasoning:  "Analysis failed",
			}
		}

		article := storage.Article{
			Ticker:      ticker,
			Title:       title,
			Link:        item.Link,
			PublishedAt: publishedAt,
			Sentiment:   signal.Sentiment,
			Confidence:  signal.Confidence,
			Reasoning:   signal.Reasoning,
			// Map sentiment string to score if needed, e.g. BULLISH=1, BEARISH=-1
			SentimentScore: 0, // Placeholder
		}

		if article.Sentiment == "BULLISH" {
			article.SentimentScore = article.Confidence
		} else if article.Sentiment == "BEARISH" {
			article.SentimentScore = -article.Confidence
		}

		storage.SaveArticle(article)
	}
}
