package calculator

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"
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

// Global Yahoo Session Variables
var (
	yahooClient *http.Client
	yahooCrumb  string
	clientOnce  sync.Once
)

const UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"

// InitYahooSession initializes the HTTP client with a cookie jar and fetches a valid crumb.
func InitYahooSession() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Printf("Failed to create cookie jar: %v", err)
		return
	}

	yahooClient = &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}

	// Step 1: Get Cookie
	// Make a GET request to https://fc.yahoo.com (or https://finance.yahoo.com).
	req, err := http.NewRequest("GET", "https://fc.yahoo.com", nil)
	if err == nil {
		req.Header.Set("User-Agent", UserAgent)
		resp, err := yahooClient.Do(req)
		if err == nil {
			// Ensure the set-cookie header is processed by the Jar
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}

	// Step 2: Get Crumb
	// Make a GET request to https://query1.finance.yahoo.com/v1/test/getcrumb
	req, err = http.NewRequest("GET", "https://query1.finance.yahoo.com/v1/test/getcrumb", nil)
	if err == nil {
		req.Header.Set("User-Agent", UserAgent)
		resp, err := yahooClient.Do(req)
		if err == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			yahooCrumb = strings.TrimSpace(string(body))
			log.Printf("Yahoo Session Initialized. Crumb: %s", yahooCrumb)
		} else {
			if err != nil {
				log.Printf("Failed to get crumb: %v", err)
			} else {
				log.Printf("Failed to get crumb. Status: %s", resp.Status)
			}
		}
	}
}

func ensureSession() {
	clientOnce.Do(InitYahooSession)
}

// GetQuote fetches the current price for a ticker using the shared Yahoo session.
func GetQuote(ticker string) (float64, error) {
	// Reusing fetchYahooOptions to get the quote as it includes it in the response.
	// This avoids maintaining a separate Quote struct/request logic for now,
	// and ensures we use the authenticated client.
	metaChain, err := fetchYahooOptions(ticker, 0)
	if err != nil || len(metaChain.OptionChain.Result) == 0 {
		log.Printf("Error fetching quote for %s: %v. Using mock data.", ticker, err)
		return getMockQuote(ticker)
	}

	return metaChain.OptionChain.Result[0].Quote.RegularMarketPrice, nil
}

func getMockQuote(ticker string) (float64, error) {
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

	// Step 1: Fetch Meta-Data (The Menu)
	// Make an initial request to https://query2.finance.yahoo.com/v7/finance/options/{ticker}.
	// Do not look at the calls or puts yet.
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
	expirationDates := result.ExpirationDates

	// Step 2: The Timestamp Matcher
	// Iterate through expirationDates. Find the timestamp that is closest to the user's targetDate.
	var matchTimestamp int64

	if !targetDate.IsZero() {
		minDiff := math.MaxFloat64
		for _, ts := range expirationDates {
			expiryTime := time.Unix(ts, 0)
			// Compare dates (truncate time by using math.Abs of hours diff)
			diff := math.Abs(expiryTime.Sub(targetDate).Hours())

			if diff < minDiff {
				minDiff = diff
				matchTimestamp = ts
			}
		}

		if matchTimestamp == 0 && len(expirationDates) > 0 {
			matchTimestamp = expirationDates[0]
		}
	} else {
		// Default to first available if no target is specified
		if len(expirationDates) > 0 {
			matchTimestamp = expirationDates[0]
		}
	}

	if matchTimestamp == 0 {
		log.Println("No expiration dates found in metadata.")
		return GetMockChain(ticker), nil
	}

	// Step 3: The Targeted Fetch (The Order)
	// Make a SECOND request to the API: https://query2.finance.yahoo.com/v7/finance/options/{ticker}?date={MATCHED_TIMESTAMP}.
	finalChainData, err := fetchYahooOptions(ticker, matchTimestamp)
	if err != nil {
		log.Printf("Error fetching targeted chain: %v", err)
		return nil, err
	}

	if len(finalChainData.OptionChain.Result) == 0 || len(finalChainData.OptionChain.Result[0].Options) == 0 {
		return nil, fmt.Errorf("no options data found for timestamp %d", matchTimestamp)
	}

	data := finalChainData.OptionChain.Result[0]
	optData := data.Options[0] // contracts are here
	contractsCount := len(optData.Calls) + len(optData.Puts)

	// Step 4: Debug Log
	log.Printf("Target: %s | Found Timestamp: %d | Contracts Fetched: %d", targetDateStr, matchTimestamp, contractsCount)

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
	ensureSession()

	// Updated to query2.finance.yahoo.com
	url := fmt.Sprintf("https://query2.finance.yahoo.com/v7/finance/options/%s", ticker)

	// Prepare Query Params
	params := []string{}

	if date > 0 {
		params = append(params, fmt.Sprintf("date=%d", date))
	}

	// Append crumb
	if yahooCrumb != "" {
		params = append(params, fmt.Sprintf("crumb=%s", yahooCrumb))
	}

	if len(params) > 0 {
		url = fmt.Sprintf("%s?%s", url, strings.Join(params, "&"))
	}

	// Use shared yahooClient
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return YahooOptionsResponse{}, err
	}

	req.Header.Set("User-Agent", UserAgent)

	resp, err := yahooClient.Do(req)
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
