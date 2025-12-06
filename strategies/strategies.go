package strategies

import (
	"fmt"
	"math"
	"sort"
	"strikelogic/calculator"
	"time"
)

// GenerateAllStrategies iterates through the option chain and builds standard trades
func GenerateAllStrategies(chain []calculator.OptionContract, targetDate string, sentiment string, targetPrice float64) ([]Trade, error) {
	// 1. Strict Filter: Ensure we only work with the date closest to targetDate
	// The incoming chain might contain one or multiple dates depending on how strict the fetch was.
	// We re-apply the strict "Closest Date" logic to be 100% sure we isolate one single expiry.
	filteredChain := filterChainByClosestDate(chain, targetDate)
	if len(filteredChain) == 0 {
		return nil, fmt.Errorf("No contracts found for date %s", targetDate)
	}

	// 2. Get current price (from first option underlying or fetch)
	// We'll assume we can get it from the first option's underlying
	ticker := filteredChain[0].Underlying
	currentPrice, err := calculator.GetQuote(ticker)
	if err != nil {
		// Fallback if quote fails, try to infer from chain (not reliable) or just error
		// For now, let's assume we can get it.
		// If calculator.GetQuote is not available or fails, we might need a fallback.
		// But let's rely on it as per previous files.
		return nil, fmt.Errorf("failed to get quote for %s: %v", ticker, err)
	}

	var trades []Trade

	// Define Recipes
	// Define Recipes
	recipes := []StrategyRecipe{
		{
			Name:        "Long Call",
			Description: "Buy 1 Call (Strike = Target Price)",
			Builder: func(c []calculator.OptionContract, tp float64) *Trade {
				opt := findClosest(c, tp, calculator.Call)
				if opt == nil {
					return nil
				}
				return &Trade{
					Name:        "Long Call",
					Description: fmt.Sprintf("Buy 1 Call at Strike %.2f", opt.Strike),
					Sentiment:   "Bullish",
					Legs: []TradeLeg{
						{Action: Buy, Quantity: 1, Option: *opt},
					},
				}
			},
		},
		{
			Name:        "Long Put",
			Description: "Buy 1 Put (Strike = Target Price)",
			Builder: func(c []calculator.OptionContract, tp float64) *Trade {
				opt := findClosest(c, tp, calculator.Put)
				if opt == nil {
					return nil
				}
				return &Trade{
					Name:        "Long Put",
					Description: fmt.Sprintf("Buy 1 Put at Strike %.2f", opt.Strike),
					Sentiment:   "Bearish",
					Legs: []TradeLeg{
						{Action: Buy, Quantity: 1, Option: *opt},
					},
				}
			},
		},
		{
			Name:        "Covered Call",
			Description: "Buy 100 Shares + Sell 1 OTM Call",
			Builder: func(c []calculator.OptionContract, tp float64) *Trade {
				call := findOTM(c, currentPrice, calculator.Call, 1)
				if call == nil {
					return nil
				}
				return &Trade{
					Name:        "Covered Call",
					Description: fmt.Sprintf("Buy 100 Shares + Sell 1 Call at Strike %.2f", call.Strike),
					Sentiment:   "Bullish",
					Legs: []TradeLeg{
						{Action: Buy, Quantity: 100, IsStock: true, StockPrice: currentPrice},
						{Action: Sell, Quantity: 1, Option: *call},
					},
				}
			},
		},
		{
			Name:        "Cash-Secured Put",
			Description: "Sell 1 OTM Put",
			Builder: func(c []calculator.OptionContract, tp float64) *Trade {
				put := findOTM(c, currentPrice, calculator.Put, 1)
				if put == nil {
					return nil
				}
				return &Trade{
					Name:        "Cash-Secured Put",
					Description: fmt.Sprintf("Sell 1 Put at Strike %.2f", put.Strike),
					Sentiment:   "Bullish",
					Legs: []TradeLeg{
						{Action: Sell, Quantity: 1, Option: *put},
					},
				}
			},
		},
		{
			Name:        "Bull Call Spread",
			Description: "Buy ITM Call + Sell OTM Call",
			Builder: func(c []calculator.OptionContract, tp float64) *Trade {
				buyLeg := findITM(c, currentPrice, calculator.Call, 1)
				sellLeg := findOTM(c, currentPrice, calculator.Call, 1)
				if buyLeg == nil || sellLeg == nil {
					return nil
				}
				return &Trade{
					Name:        "Bull Call Spread",
					Description: fmt.Sprintf("Buy Call %.2f, Sell Call %.2f", buyLeg.Strike, sellLeg.Strike),
					Sentiment:   "Bullish",
					Legs: []TradeLeg{
						{Action: Buy, Quantity: 1, Option: *buyLeg},
						{Action: Sell, Quantity: 1, Option: *sellLeg},
					},
				}
			},
		},
		{
			Name:        "Bull Put Spread",
			Description: "Buy OTM Put + Sell Higher Strike Put",
			Builder: func(c []calculator.OptionContract, tp float64) *Trade {
				sellLeg := findOTM(c, currentPrice, calculator.Put, 1)
				buyLeg := findOTM(c, currentPrice, calculator.Put, 3)
				if sellLeg == nil || buyLeg == nil {
					return nil
				}
				return &Trade{
					Name:        "Bull Put Spread",
					Description: fmt.Sprintf("Sell Put %.2f, Buy Put %.2f", sellLeg.Strike, buyLeg.Strike),
					Sentiment:   "Bullish",
					Legs: []TradeLeg{
						{Action: Sell, Quantity: 1, Option: *sellLeg},
						{Action: Buy, Quantity: 1, Option: *buyLeg},
					},
				}
			},
		},
		{
			Name:        "Bear Call Spread",
			Description: "Sell ITM Call + Buy OTM Call",
			Builder: func(c []calculator.OptionContract, tp float64) *Trade {
				sellLeg := findITM(c, currentPrice, calculator.Call, 1)
				buyLeg := findOTM(c, currentPrice, calculator.Call, 1)
				if sellLeg == nil || buyLeg == nil {
					return nil
				}
				return &Trade{
					Name:        "Bear Call Spread",
					Description: fmt.Sprintf("Sell Call %.2f, Buy Call %.2f", sellLeg.Strike, buyLeg.Strike),
					Sentiment:   "Bearish",
					Legs: []TradeLeg{
						{Action: Sell, Quantity: 1, Option: *sellLeg},
						{Action: Buy, Quantity: 1, Option: *buyLeg},
					},
				}
			},
		},
		{
			Name:        "Bear Put Spread",
			Description: "Buy ITM Put + Sell OTM Put",
			Builder: func(c []calculator.OptionContract, tp float64) *Trade {
				buyLeg := findITM(c, currentPrice, calculator.Put, 1)
				sellLeg := findOTM(c, currentPrice, calculator.Put, 1)
				if buyLeg == nil || sellLeg == nil {
					return nil
				}
				return &Trade{
					Name:        "Bear Put Spread",
					Description: fmt.Sprintf("Buy Put %.2f, Sell Put %.2f", buyLeg.Strike, sellLeg.Strike),
					Sentiment:   "Bearish",
					Legs: []TradeLeg{
						{Action: Buy, Quantity: 1, Option: *buyLeg},
						{Action: Sell, Quantity: 1, Option: *sellLeg},
					},
				}
			},
		},
		{
			Name:        "Straddle",
			Description: "Buy ATM Call + Buy ATM Put (Same Strike)",
			Builder: func(c []calculator.OptionContract, tp float64) *Trade {
				call := findClosest(c, currentPrice, calculator.Call)
				put := findClosest(c, currentPrice, calculator.Put)
				if call == nil || put == nil {
					return nil
				}
				if call.Strike != put.Strike {
					put = findClosest(c, call.Strike, calculator.Put)
				}
				if put == nil || call.Strike != put.Strike {
					return nil
				}

				return &Trade{
					Name:        "Straddle",
					Description: fmt.Sprintf("Buy Call & Put at Strike %.2f", call.Strike),
					Sentiment:   "Neutral",
					Legs: []TradeLeg{
						{Action: Buy, Quantity: 1, Option: *call},
						{Action: Buy, Quantity: 1, Option: *put},
					},
				}
			},
		},
		{
			Name:        "Strangle",
			Description: "Buy OTM Call + Buy OTM Put (Different Strikes)",
			Builder: func(c []calculator.OptionContract, tp float64) *Trade {
				call := findOTM(c, currentPrice, calculator.Call, 1)
				put := findOTM(c, currentPrice, calculator.Put, 1)
				if call == nil || put == nil {
					return nil
				}
				return &Trade{
					Name:        "Strangle",
					Description: fmt.Sprintf("Buy Call %.2f, Buy Put %.2f", call.Strike, put.Strike),
					Sentiment:   "Neutral",
					Legs: []TradeLeg{
						{Action: Buy, Quantity: 1, Option: *call},
						{Action: Buy, Quantity: 1, Option: *put},
					},
				}
			},
		},
		{
			Name:        "Iron Condor",
			Description: "Sell OTM Put + Buy Further OTM Put + Sell OTM Call + Buy Further OTM Call",
			Builder: func(c []calculator.OptionContract, tp float64) *Trade {
				sellPut := findOTM(c, currentPrice, calculator.Put, 1)
				buyPut := findOTM(c, currentPrice, calculator.Put, 3)
				sellCall := findOTM(c, currentPrice, calculator.Call, 1)
				buyCall := findOTM(c, currentPrice, calculator.Call, 3)

				if sellPut == nil || buyPut == nil || sellCall == nil || buyCall == nil {
					return nil
				}

				return &Trade{
					Name:        "Iron Condor",
					Description: fmt.Sprintf("Sell Put %.2f/Call %.2f, Buy Put %.2f/Call %.2f", sellPut.Strike, sellCall.Strike, buyPut.Strike, buyCall.Strike),
					Sentiment:   "Neutral",
					Legs: []TradeLeg{
						{Action: Sell, Quantity: 1, Option: *sellPut},
						{Action: Buy, Quantity: 1, Option: *buyPut},
						{Action: Sell, Quantity: 1, Option: *sellCall},
						{Action: Buy, Quantity: 1, Option: *buyCall},
					},
				}
			},
		},
		{
			Name:        "Iron Butterfly",
			Description: "Sell ATM Put + Sell ATM Call + Buy OTM Put + Buy OTM Call",
			Builder: func(c []calculator.OptionContract, tp float64) *Trade {
				atmCall := findClosest(c, currentPrice, calculator.Call)
				atmPut := findClosest(c, currentPrice, calculator.Put)
				otmCall := findOTM(c, currentPrice, calculator.Call, 2)
				otmPut := findOTM(c, currentPrice, calculator.Put, 2)

				if atmCall == nil || atmPut == nil || otmCall == nil || otmPut == nil {
					return nil
				}
				if atmCall.Strike != atmPut.Strike {
					atmPut = findClosest(c, atmCall.Strike, calculator.Put)
				}
				if atmPut == nil {
					return nil
				}

				return &Trade{
					Name:        "Iron Butterfly",
					Description: fmt.Sprintf("Sell ATM %.2f, Buy OTM Call %.2f/Put %.2f", atmCall.Strike, otmCall.Strike, otmPut.Strike),
					Sentiment:   "Neutral",
					Legs: []TradeLeg{
						{Action: Sell, Quantity: 1, Option: *atmPut},
						{Action: Sell, Quantity: 1, Option: *atmCall},
						{Action: Buy, Quantity: 1, Option: *otmPut},
						{Action: Buy, Quantity: 1, Option: *otmCall},
					},
				}
			},
		},
		{
			Name:        "Call Broken Wing Butterfly",
			Description: "Buy 1 ITM Call + Sell 2 ATM Calls + Buy 1 OTM Call (Strikes are NOT equidistant)",
			Builder: func(c []calculator.OptionContract, tp float64) *Trade {
				itm := findITM(c, currentPrice, calculator.Call, 1)
				atm := findClosest(c, currentPrice, calculator.Call)
				otm := findOTM(c, currentPrice, calculator.Call, 2)

				if itm == nil || atm == nil || otm == nil {
					return nil
				}

				return &Trade{
					Name:        "Call Broken Wing Butterfly",
					Description: fmt.Sprintf("Buy %.2f, Sell 2x %.2f, Buy %.2f", itm.Strike, atm.Strike, otm.Strike),
					Sentiment:   "Bullish",
					Legs: []TradeLeg{
						{Action: Buy, Quantity: 1, Option: *itm},
						{Action: Sell, Quantity: 2, Option: *atm},
						{Action: Buy, Quantity: 1, Option: *otm},
					},
				}
			},
		},
	}

	// Filter by sentiment if provided
	// Normalize sentiment
	targetSentiment := ""
	if sentiment != "" {
		if sentiment == "very_bullish" || sentiment == "bullish" || sentiment == "directional" {
			targetSentiment = "Bullish"
		} else if sentiment == "very_bearish" || sentiment == "bearish" {
			targetSentiment = "Bearish"
		} else if sentiment == "neutral" {
			targetSentiment = "Neutral"
		}
	}

	for _, recipe := range recipes {
		trade := recipe.Builder(filteredChain, targetPrice)
		if trade != nil {
			// Filter logic
			if targetSentiment != "" && trade.Sentiment != targetSentiment {
				continue
			}

			trade.CalculateMetrics(currentPrice)
			trade.ExpirationDate = filteredChain[0].Expiry

			// Calculate Expiry Label
			expiryDate, _ := time.Parse("2006-01-02", trade.ExpirationDate)
			daysToExpiry := math.Ceil(expiryDate.Sub(time.Now()).Hours() / 24)
			trade.ExpiryLabel = fmt.Sprintf("%s (%.0fd)", expiryDate.Format("Jan 02"), daysToExpiry)

			trades = append(trades, *trade)
		}
	}

	return trades, nil
}

// Helper functions

func filterChainByClosestDate(chain []calculator.OptionContract, targetDateStr string) []calculator.OptionContract {
	// 1. Parse target date
	targetDate, err := time.Parse("2006-01-02", targetDateStr)
	if err != nil {
		// Fallback to exact string match check if parsing fails
		return filterChainByDate(chain, targetDateStr)
	}

	// 2. Identify all unique expiration dates in the chain
	expiries := make(map[string]bool)
	for _, opt := range chain {
		expiries[opt.Expiry] = true
	}

	// 3. Find the single expiration date mathematically closest to targetDate
	var bestExpiry string
	minDiff := math.MaxFloat64
	found := false

	for expiryStr := range expiries {
		expiryDate, err := time.Parse("2006-01-02", expiryStr)
		if err != nil {
			continue
		}

		// Calculate absolute difference in hours
		diff := math.Abs(expiryDate.Sub(targetDate).Hours())

		if diff < minDiff {
			minDiff = diff
			bestExpiry = expiryStr
			found = true
		}
	}

	if !found {
		return nil
	}

	// 4. Filter the chain to include ONLY contracts matching the best expiration
	var filtered []calculator.OptionContract
	for _, opt := range chain {
		if opt.Expiry == bestExpiry {
			filtered = append(filtered, opt)
		}
	}
	return filtered
}

func filterChainByDate(chain []calculator.OptionContract, targetDate string) []calculator.OptionContract {
	var filtered []calculator.OptionContract
	for _, opt := range chain {
		if opt.Expiry == targetDate {
			filtered = append(filtered, opt)
		}
	}
	return filtered
}

func findClosest(chain []calculator.OptionContract, targetStrike float64, optType calculator.OptionType) *calculator.OptionContract {
	var best *calculator.OptionContract
	minDiff := math.MaxFloat64

	for i := range chain {
		if chain[i].Type != optType {
			continue
		}
		diff := math.Abs(chain[i].Strike - targetStrike)
		if diff < minDiff {
			minDiff = diff
			best = &chain[i]
		}
	}
	return best
}

func findOTM(chain []calculator.OptionContract, currentPrice float64, optType calculator.OptionType, steps int) *calculator.OptionContract {
	// Sort chain by strike
	sorted := make([]calculator.OptionContract, len(chain))
	copy(sorted, chain)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Strike < sorted[j].Strike
	})

	// Find ATM index
	atmIndex := -1
	minDiff := math.MaxFloat64
	for i, opt := range sorted {
		if opt.Type != optType {
			continue
		}
		diff := math.Abs(opt.Strike - currentPrice)
		if diff < minDiff {
			minDiff = diff
			atmIndex = i
		}
	}

	if atmIndex == -1 {
		return nil
	}

	if optType == calculator.Call {
		// OTM Call: Strike > Current Price. Higher index.
		// Ensure we are actually OTM (Strike > Price)
		// If ATM strike is < Price, we might need more steps.
		// Let's just iterate from ATM upwards
		count := 0
		for i := atmIndex; i < len(sorted); i++ {
			if sorted[i].Type == optType && sorted[i].Strike > currentPrice {
				count++
				if count == steps {
					return &sorted[i]
				}
			}
		}
	} else {
		// OTM Put: Strike < Current Price. Lower index.
		count := 0
		for i := atmIndex; i >= 0; i-- {
			if sorted[i].Type == optType && sorted[i].Strike < currentPrice {
				count++
				if count == steps {
					return &sorted[i]
				}
			}
		}
	}
	return nil
}

func findITM(chain []calculator.OptionContract, currentPrice float64, optType calculator.OptionType, steps int) *calculator.OptionContract {
	// Sort chain by strike
	sorted := make([]calculator.OptionContract, len(chain))
	copy(sorted, chain)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Strike < sorted[j].Strike
	})

	// Find ATM index
	atmIndex := -1
	minDiff := math.MaxFloat64
	for i, opt := range sorted {
		if opt.Type != optType {
			continue
		}
		diff := math.Abs(opt.Strike - currentPrice)
		if diff < minDiff {
			minDiff = diff
			atmIndex = i
		}
	}

	if atmIndex == -1 {
		return nil
	}

	if optType == calculator.Call {
		// ITM Call: Strike < Current Price. Lower index.
		count := 0
		for i := atmIndex; i >= 0; i-- {
			if sorted[i].Type == optType && sorted[i].Strike < currentPrice {
				count++
				if count == steps {
					return &sorted[i]
				}
			}
		}
	} else {
		// ITM Put: Strike > Current Price. Higher index.
		count := 0
		for i := atmIndex; i < len(sorted); i++ {
			if sorted[i].Type == optType && sorted[i].Strike > currentPrice {
				count++
				if count == steps {
					return &sorted[i]
				}
			}
		}
	}
	return nil
}
