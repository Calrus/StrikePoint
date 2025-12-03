package news_engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Signal struct {
	Ticker     string  `json:"ticker"`
	Sentiment  string  `json:"sentiment"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
	Headline   string  `json:"headline"`
}

func FetchHeadlines() ([]string, error) {
	// In a real production environment, we would use newsapi.org with an API key.
	// For this demo, we will simulate fetched headlines matching the keywords.
	headlines := []string{
		"Elon Musk tweets: 'Tesla FSD v12 is mind-blowing. Rollout accelerating.'",
		"Jerome Powell signals rates may remain higher for longer as inflation persists.",
		"Trump comments on trade tariffs: 'We need to protect our auto industry.'",
		"Tesla Guidance update: Production targets increased for Q4.",
		"Market Guidance: Tech sector faces headwinds as yields rise.",
	}
	return headlines, nil
}

func AnalyzeSentiment(text string) ([]Signal, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY not set")
	}

	prompt := "You are a hedge fund algo. Analyze this text for market sentiment. Identify all relevant tickers. Return valid JSON (RFC 8259) as an array of objects. Example: [{ \"ticker\": \"TSLA\", \"sentiment\": \"BULLISH\", \"confidence\": 0.9, \"reasoning\": \"...\" }]. If only one ticker, return an array with one object."

	requestBody, _ := json.Marshal(map[string]interface{}{
		"model": "x-ai/grok-4.1-fast:free",
		"messages": []map[string]string{
			{"role": "system", "content": prompt},
			{"role": "user", "content": text},
		},
	})

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "http://localhost:8081")
	req.Header.Set("X-Title", "StrikePoint")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	content := result.Choices[0].Message.Content
	// Clean up content if it contains markdown code blocks
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var signals []Signal
	// Try unmarshalling as array first
	if err := json.Unmarshal([]byte(content), &signals); err != nil {
		// If that fails, try as single object
		var singleSignal Signal
		if err2 := json.Unmarshal([]byte(content), &singleSignal); err2 != nil {
			return nil, fmt.Errorf("failed to parse JSON: %v, content: %s", err, content)
		}
		signals = []Signal{singleSignal}
	}

	for i := range signals {
		signals[i].Headline = text
	}

	return signals, nil
}
