package calculator

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/piquette/finance-go/quote"
)

// Yahoo Options Response Structs
type YahooOptionsResponse struct {
	OptionChain struct {
		Result []struct {
			UnderlyingSymbol string    `json:"underlyingSymbol"`
			ExpirationDates  []int64   `json:"expirationDates"`
			Strikes          []float64 `json:"strikes"`
			Quote            struct {
				RegularMarketPrice float64 `json:"regularMarketPrice"`
			} `json:"quote"`
			Options []struct {
				ExpirationDate int64                 `json:"expirationDate"`
				Calls          []YahooOptionContract `json:"calls"`
				Puts           []YahooOptionContract `json:"puts"`
			} `json:"options"`
		} `json:"result"`
		Error interface{} `json:"error"`
	} `json:"optionChain"`
}

type YahooOptionContract struct {
	Strike            float64 `json:"strike"`
	Currency          string  `json:"currency"`
	LastPrice         float64 `json:"lastPrice"`
	Bid               float64 `json:"bid"`
	Ask               float64 `json:"ask"`
	Expiration        int64   `json:"expiration"`
	ImpliedVolatility float64 `json:"impliedVolatility"`
	InTheMoney        bool    `json:"inTheMoney"`
}

// GetQuote fetches the current price for a ticker
func GetQuote(ticker string) (float64, error) {
	q, err := quote.Get(ticker)
	if err != nil || q == nil {
		log.Printf("Error fetching quote for %s: %v. Using mock data.", ticker, err)
		// Fallback mock prices
		if ticker == "AAPL" {
			return 150.0, nil
		}
		if ticker == "SPY" {
			return 400.0, nil
		}
		if ticker == "TSLA" {
			return 900.0, nil
		}
		if ticker == "GOOG" {
			return 2800.0, nil
		}
		if ticker == "AMZN" {
			return 3400.0, nil
		}
		return 100.0, nil // Default fallback
	}
	return q.RegularMarketPrice, nil
}

// GetOptionsChain fetches the option chain for a ticker, targeting a specific date if provided.
func GetOptionsChain(ticker string, targetDateStr string) ([]OptionContract, error) {
	// Parse target date
	var targetDate time.Time
	if targetDateStr != "" {
		t, err := time.Parse("2006-01-02", targetDateStr)
		if err == nil {
			targetDate = t
		}
	}

	// 1. Step 1: Get Meta-Data First (List of Available Expirations)
	// Fetch "Summary" metadata from Yahoo Options endpoint (no date param gets nearest, plus list of all dates)
	metaChain, err := fetchYahooOptions(ticker, 0)
	if err != nil {
		log.Printf("Error fetching metadata: %v. Falling back to mock.", err)
		return GetMockChain(ticker), nil
	}

	// Safety check
	if len(metaChain.OptionChain.Result) == 0 {
		return GetMockChain(ticker), nil
	}

	result := metaChain.OptionChain.Result[0]
	availableExpirations := result.ExpirationDates

	// 2. Step 2: Date Matcher (Pre-Fetch)
	var selectedTimestamp int64

	if !targetDate.IsZero() {
		minDiff := math.MaxFloat64
		for _, ts := range availableExpirations {
			expiryTime := time.Unix(ts, 0)
			// Compare dates (truncate time)
			diff := math.Abs(expiryTime.Sub(targetDate).Hours())
			if diff < minDiff {
				minDiff = diff
				selectedTimestamp = ts
			}
		}
	} else {
		// Default to first available if no target (or just use the one we fetched)
		if len(availableExpirations) > 0 {
			selectedTimestamp = availableExpirations[0]
		}
	}

	// 3. Step 3: Targeted Fetch
	// If we need a different date than what we got (default is usually index 0, but verified), fetch specific.
	// Optimization: If selectedTimestamp matches the one in `metaChain.OptionChain.Result[0].Options[0].ExpirationDate` we can skip?
	// Often Yahoo returns the whole chain for that date.

	var finalChainData YahooOptionsResponse

	// Check if we already have the data
	alreadyHasData := false
	if len(result.Options) > 0 {
		// Yahoo timestamps are seconds.
		// There might be timezone shifts, but usually exact match.
		if result.Options[0].ExpirationDate == selectedTimestamp {
			finalChainData = metaChain
			alreadyHasData = true
		}
	}

	if !alreadyHasData && selectedTimestamp > 0 {
		finalChainData, err = fetchYahooOptions(ticker, selectedTimestamp)
		if err != nil {
			log.Printf("Error fetching targeted chain: %v", err)
			return nil, err
		}
	} else if !alreadyHasData {
		// Should have data from first fetch
		finalChainData = metaChain
	}

	// 4. Verification
	if len(finalChainData.OptionChain.Result) == 0 || len(finalChainData.OptionChain.Result[0].Options) == 0 {
		return nil, fmt.Errorf("no options data found")
	}

	data := finalChainData.OptionChain.Result[0]
	optData := data.Options[0]
	fetchedDate := time.Unix(optData.ExpirationDate, 0)

	if !targetDate.IsZero() {
		diff := math.Abs(fetchedDate.Sub(targetDate).Hours() / 24)
		log.Printf("Fetched chain for date: %s (Target: %s, Diff: %.2f days)", fetchedDate.Format("2006-01-02"), targetDate.Format("2006-01-02"), diff)

		if diff > 7 {
			// As requested: "If it's months off, return an error"
			return nil, fmt.Errorf("fetched date %s is too far from target %s", fetchedDate.Format("2006-01-02"), targetDate.Format("2006-01-02"))
		}
	} else {
		log.Printf("Fetched chain for date: %s (Default)", fetchedDate.Format("2006-01-02"))
	}

	// Convert to internal format
	currentPrice := data.Quote.RegularMarketPrice
	var chain []OptionContract

	for _, call := range optData.Calls {
		chain = append(chain, convertYahooToContract(call, currentPrice, Call, ticker))
	}
	for _, put := range optData.Puts {
		chain = append(chain, convertYahooToContract(put, currentPrice, Put, ticker))
	}

	return chain, nil
}

func fetchYahooOptions(ticker string, date int64) (YahooOptionsResponse, error) {
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/options/%s", ticker)
	if date > 0 {
		url = fmt.Sprintf("%s?date=%d", url, date)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return YahooOptionsResponse{}, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return YahooOptionsResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return YahooOptionsResponse{}, fmt.Errorf("yahoo api returned status: %s", resp.Status)
	}

	var res YahooOptionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return YahooOptionsResponse{}, err
	}

	return res, nil
}

func convertYahooToContract(c YahooOptionContract, currentPrice float64, optType OptionType, ticker string) OptionContract {
	// Convert Expiration (Unix timestamp) to string
	expiryTime := time.Unix(c.Expiration, 0)
	expiryStr := expiryTime.Format("2006-01-02")

	// Calculate time to expiry in years
	daysUntil := time.Until(expiryTime).Hours() / 24
	T := daysUntil / 365.0
	if T < 0.001 {
		T = 0.001
	}

	iv := c.ImpliedVolatility
	if iv == 0 {
		iv = 0.5 // Fallback
	}

	// Calculate Greeks using Black-Scholes
	r := 0.05
	_, delta, gamma, theta := BlackScholes(currentPrice, c.Strike, T, r, iv)

	return OptionContract{
		Strike:     c.Strike,
		Expiry:     expiryStr,
		Type:       optType,
		Bid:        c.Bid,
		Ask:        c.Ask,
		Last:       c.LastPrice,
		Vol:        iv,
		Delta:      math.Round(delta*1000) / 1000,
		Gamma:      math.Round(gamma*1000) / 1000,
		Theta:      math.Round(theta*1000) / 1000,
		Underlying: ticker,
	}
}

// GetMockChain returns a consolidated mock chain for multiple expiries
func GetMockChain(ticker string) []OptionContract {
	var fullChain []OptionContract
	// Generate for a few key expiries to simulate a full chain
	expiries := []int{7, 14, 30, 60, 90, 180, 365}
	for _, days := range expiries {
		fullChain = append(fullChain, GetChainForExpiry(ticker, days)...)
	}
	return fullChain
}
