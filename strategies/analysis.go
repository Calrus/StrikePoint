package strategies

import (
	"math"
	"strikelogic/calculator"
)

// CalculateMetrics computes MaxProfit, MaxRisk, BreakEvens, and Greeks for the trade
func (t *Trade) CalculateMetrics(currentPrice float64) {
	// 1. Calculate Net Debit/Credit
	t.NetDebit = 0
	for _, leg := range t.Legs {
		cost := leg.Option.Ask // Assuming we buy at Ask
		if leg.Action == Sell {
			cost = leg.Option.Bid // Sell at Bid
		}

		// If it's a stock leg
		if leg.IsStock {
			cost = leg.StockPrice
		}

		amount := cost * float64(leg.Quantity)
		if !leg.IsStock {
			amount *= 100 // Options multiplier
		}

		if leg.Action == Buy {
			t.NetDebit += amount
		} else {
			t.NetDebit -= amount
		}
	}

	// 2. Calculate Greeks (Portfolio Greeks)
	t.Delta = 0
	t.Gamma = 0
	t.Theta = 0
	t.Vega = 0

	for _, leg := range t.Legs {
		if leg.IsStock {
			// Stock has Delta 1, others 0
			q := float64(leg.Quantity)
			if leg.Action == Sell {
				q = -q
			}
			t.Delta += q
			continue
		}

		q := float64(leg.Quantity) * 100
		if leg.Action == Sell {
			q = -q
		}

		t.Delta += leg.Option.Delta * q
		t.Gamma += leg.Option.Gamma * q
		t.Theta += leg.Option.Theta * q
		t.Vega += leg.Option.Vega * q
	}

	// 3. Analyze P&L Profile to find MaxProfit, MaxRisk, BreakEvens
	// We scan a range of prices at expiry.
	// Range: 0 to 2 * currentPrice (or higher if needed)
	// Resolution: 1000 points?

	minPrice := 0.0
	maxPrice := currentPrice * 3.0
	steps := 1000
	stepSize := (maxPrice - minPrice) / float64(steps)

	maxProfit := -math.MaxFloat64
	minProfit := math.MaxFloat64 // This tracks the lowest profit (max loss)

	var pnlPoints []float64
	prices := []float64{}

	for i := 0; i <= steps; i++ {
		price := minPrice + float64(i)*stepSize
		pnl := t.CalculatePnLAtExpiry(price)

		if pnl > maxProfit {
			maxProfit = pnl
		}
		if pnl < minProfit {
			minProfit = pnl
		}

		pnlPoints = append(pnlPoints, pnl)
		prices = append(prices, price)
	}

	t.MaxProfit = maxProfit
	t.MaxRisk = -minProfit // Risk is usually expressed as a positive number (amount at risk)
	if t.MaxRisk < 0 {
		t.MaxRisk = 0 // If minProfit is positive, there is no risk (arbitrage?)
	}

	// Find BreakEvens (Zero crossings)
	var breakEvens []float64
	for i := 0; i < len(pnlPoints)-1; i++ {
		p1 := pnlPoints[i]
		p2 := pnlPoints[i+1]

		if (p1 < 0 && p2 >= 0) || (p1 >= 0 && p2 < 0) {
			// Linear interpolation for more precision
			// 0 = p1 + (p2 - p1) * (x - x1) / (x2 - x1)
			// x = x1 - p1 * (x2 - x1) / (p2 - p1)
			x1 := prices[i]
			x2 := prices[i+1]
			be := x1 - p1*(x2-x1)/(p2-p1)
			breakEvens = append(breakEvens, math.Round(be*100)/100)
		}
	}
	t.BreakEvens = breakEvens
}

// CalculatePnLAtExpiry calculates the P&L of the trade if the underlying is at `price` at expiry
func (t *Trade) CalculatePnLAtExpiry(price float64) float64 {
	pnl := 0.0

	// Start with initial debit/credit
	// If NetDebit is positive, we paid money. P&L starts at -NetDebit.
	// If NetDebit is negative (credit), we received money. P&L starts at -NetDebit (positive).
	pnl -= t.NetDebit

	for _, leg := range t.Legs {
		value := 0.0
		if leg.IsStock {
			value = price * float64(leg.Quantity)
		} else {
			// Option value at expiry
			optValue := 0.0
			if leg.Option.Type == calculator.Call {
				if price > leg.Option.Strike {
					optValue = price - leg.Option.Strike
				}
			} else {
				if price < leg.Option.Strike {
					optValue = leg.Option.Strike - price
				}
			}
			value = optValue * 100 * float64(leg.Quantity)
		}

		if leg.Action == Buy {
			pnl += value
		} else {
			pnl -= value
		}
	}
	return pnl
}
