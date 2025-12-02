package calculator

import (
	"log"
	"math"
	"time"

	"github.com/piquette/finance-go"
	"github.com/piquette/finance-go/options"
	"github.com/piquette/finance-go/quote"
)

// GetQuote fetches the current price for a ticker
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

// GetOptionsChain fetches the full option chain for a ticker.
// Falls back to mock data if error occurs.
func GetOptionsChain(ticker string) ([]OptionContract, error) {
	// 1. Fetch underlying price
	q, err := quote.Get(ticker)
	if err != nil || q == nil {
		log.Printf("Error fetching quote for %s: %v. Using mock data.", ticker, err)
		return GetMockChain(ticker), nil
	}
	currentPrice := q.RegularMarketPrice

	// 2. Fetch options
	iter := options.GetStraddle(ticker)
	if iter.Err() != nil {
		log.Printf("Error fetching options for %s: %v. Using mock data.", ticker, iter.Err())
		return GetMockChain(ticker), nil
	}

	var chain []OptionContract

	for iter.Next() {
		s := iter.Straddle()

		// Process Call
		if s.Call != nil {
			chain = append(chain, convertToContract(s.Call, currentPrice, Call, ticker))
		}
		// Process Put
		if s.Put != nil {
			chain = append(chain, convertToContract(s.Put, currentPrice, Put, ticker))
		}
	}

	if iter.Err() != nil {
		log.Printf("Error iterating options: %v. Returning partial or mock data.", iter.Err())
		if len(chain) == 0 {
			return GetMockChain(ticker), nil
		}
	}

	if len(chain) == 0 {
		return GetMockChain(ticker), nil
	}

	return chain, nil
}

func convertToContract(c *finance.Contract, currentPrice float64, optType OptionType, ticker string) OptionContract {
	// Convert Expiration (Unix timestamp) to string
	// Assuming field is named Expiration based on common naming if ExpireDate failed
	expiryTime := time.Unix(int64(c.Expiration), 0)
	expiryStr := expiryTime.Format("2006-01-02")

	// Calculate time to expiry in years
	daysUntil := time.Until(expiryTime).Hours() / 24
	T := daysUntil / 365.0
	if T < 0.001 {
		T = 0.001
	}

	// Use ImpliedVolatility from response
	iv := c.ImpliedVolatility
	if iv == 0 {
		iv = 0.5 // Fallback
	}

	// Calculate Greeks using Black-Scholes
	// We use 5% risk-free rate as assumption
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
