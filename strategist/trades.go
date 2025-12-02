package strategist

import (
	"fmt"
	"math"
	"strikelogic/calculator"
	"time"
)

type RiskProfile string

const (
	Low    RiskProfile = "Low (The Landlord)"
	Medium RiskProfile = "Medium (The Strategist)"
	Degen  RiskProfile = "Degen (The Moonshot)"
)

type TradeLeg struct {
	Action string                    `json:"action"` // "Buy" or "Sell"
	Option calculator.OptionContract `json:"option"`
}

type Trade struct {
	RiskProfile RiskProfile `json:"riskProfile"`
	Description string      `json:"description"`
	Legs        []TradeLeg  `json:"legs"`
	NetDebit    float64     `json:"netDebit"`
	MaxProfit   string      `json:"maxProfit"` // String for flexibility (e.g. "Unlimited")
	ROI         string      `json:"roi"`       // Estimated ROI
}

// FindTrades generates 3 distinct trade ideas
func FindTrades(ticker string, currentPrice, targetPrice float64, targetDateStr string) ([]Trade, error) {
	var trades []Trade

	// Fetch full option chain (Real data with fallback)
	fullChain, err := calculator.GetOptionsChain(ticker)
	if err != nil {
		return nil, err
	}

	// 1. Low Risk: PMCC (The Landlord)
	// Buy Deep ITM Call (~0.85 Delta) 6 months out
	// Sell OTM Call (~0.30 Delta) 1 month out
	longChain := filterChainByDays(fullChain, 180)
	shortChain := filterChainByDays(fullChain, 30)

	longLeg := findClosestDelta(longChain, 0.85)
	shortLeg := findClosestDelta(shortChain, 0.30)

	if longLeg != nil && shortLeg != nil {
		netDebit := longLeg.Ask - shortLeg.Bid
		trades = append(trades, Trade{
			RiskProfile: Low,
			Description: "Poor Man's Covered Call (PMCC). Buy deep ITM LEAPS and sell monthly calls against it for income.",
			Legs: []TradeLeg{
				{Action: "Buy", Option: *longLeg},
				{Action: "Sell", Option: *shortLeg},
			},
			NetDebit:  math.Round(netDebit*100) / 100,
			MaxProfit: "Variable (Income Generation)",
			ROI:       "~15-25% annualized",
		})
	}

	// 2. Medium Risk: PMCC (The Strategist)
	// Buy ITM Call (~0.70 Delta) 3 months out
	// Sell OTM Call (~0.40 Delta) 2 weeks out
	longChainMed := filterChainByDays(fullChain, 90)
	shortChainMed := filterChainByDays(fullChain, 14)

	longLegMed := findClosestDelta(longChainMed, 0.70)
	shortLegMed := findClosestDelta(shortChainMed, 0.40)

	if longLegMed != nil && shortLegMed != nil {
		netDebit := longLegMed.Ask - shortLegMed.Bid
		trades = append(trades, Trade{
			RiskProfile: Medium,
			Description: "Aggressive PMCC. Higher delta short call for more premium, but capped upside.",
			Legs: []TradeLeg{
				{Action: "Buy", Option: *longLegMed},
				{Action: "Sell", Option: *shortLegMed},
			},
			NetDebit:  math.Round(netDebit*100) / 100,
			MaxProfit: "Capped at Short Strike",
			ROI:       "~30-50% annualized",
		})
	}

	// 3. Degen: Naked Call (The Moonshot)
	// Find highest ROI if price hits targetPrice by targetDate
	targetDate, err := time.Parse("2006-01-02", targetDateStr)
	if err != nil {
		targetDate = time.Now().AddDate(0, 1, 0)
	}

	daysToTarget := targetDate.Sub(time.Now()).Hours() / 24
	if daysToTarget < 1 {
		daysToTarget = 1
	}

	// Look at options expiring shortly after target date
	daysOut := int(daysToTarget) + 7
	degenChain := filterChainByDays(fullChain, daysOut)

	var bestOption *calculator.OptionContract
	maxROI := -1.0

	for i := range degenChain {
		opt := &degenChain[i]
		cost := opt.Ask
		if cost == 0 {
			continue
		}

		// Calculate theoretical price at target
		timeRemaining := float64(daysOut-int(daysToTarget)) / 365.0
		if timeRemaining < 0 {
			timeRemaining = 0
		}

		theoPrice, _, _, _ := calculator.BlackScholes(targetPrice, opt.Strike, timeRemaining, 0.05, opt.Vol)

		profit := theoPrice - cost
		roi := profit / cost

		if roi > maxROI {
			maxROI = roi
			bestOption = opt
		}
	}

	if bestOption != nil {
		trades = append(trades, Trade{
			RiskProfile: Degen,
			Description: "Naked Call. Highest theoretical ROI if target hit.",
			Legs: []TradeLeg{
				{Action: "Buy", Option: *bestOption},
			},
			NetDebit:  bestOption.Ask,
			MaxProfit: "Unlimited",
			ROI:       fmt.Sprintf("%.0f%%", maxROI*100),
		})
	}

	return trades, nil
}

func filterChainByDays(fullChain []calculator.OptionContract, targetDays int) []calculator.OptionContract {
	var filtered []calculator.OptionContract
	targetDate := time.Now().AddDate(0, 0, targetDays)

	var bestExpiry string
	minDiff := 10000.0 // days

	// Find closest expiry
	expiries := make(map[string]bool)
	for _, opt := range fullChain {
		expiries[opt.Expiry] = true
	}

	for expiry := range expiries {
		t, _ := time.Parse("2006-01-02", expiry)
		diff := math.Abs(t.Sub(targetDate).Hours() / 24)
		if diff < minDiff {
			minDiff = diff
			bestExpiry = expiry
		}
	}

	// Filter by best expiry and Call type
	for _, opt := range fullChain {
		if opt.Expiry == bestExpiry && opt.Type == calculator.Call {
			filtered = append(filtered, opt)
		}
	}
	return filtered
}

func findClosestDelta(chain []calculator.OptionContract, targetDelta float64) *calculator.OptionContract {
	var best *calculator.OptionContract
	minDiff := 1.0

	for i := range chain {
		diff := math.Abs(chain[i].Delta - targetDelta)
		if diff < minDiff {
			minDiff = diff
			best = &chain[i]
		}
	}
	return best
}
