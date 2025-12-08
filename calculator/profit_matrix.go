package calculator

import (
	"math"
	"time"
)

// MatrixPoint represents a single data point in the profit heatmap
type MatrixPoint struct {
	Date   string  `json:"date"`
	Price  float64 `json:"price"`
	Profit float64 `json:"profit"`
	ZScore float64 `json:"zScore"`
}

// MatrixResponse holds the grid of profit points
type MatrixResponse struct {
	Grid []MatrixPoint `json:"grid"`
}

// LegInput defines the necessary parameters for a strategy leg to be priced
type LegInput struct {
	Strike     float64
	Type       OptionType // "Call" or "Put"
	Action     string     // "Buy" or "Sell"
	Quantity   float64
	Expiry     time.Time
	IV         float64 // Implied Volatility of this specific leg
	EntryPrice float64 // The price per share paid/received for this leg
}

// StrategyInput captures the strategy details for the matrix calculation
type StrategyInput struct {
	Legs         []LegInput
	InitialDebit float64 // Deprecated in favor of per-leg EntryPrice, but kept for compatibility
}

// CalculateProfitMatrix generates a heatmap of theoretical profit/loss over time and price
func CalculateProfitMatrix(strategy StrategyInput, currentPrice float64, volatility float64) (MatrixResponse, error) {
	// 1. Identify Time Horizon
	var expiryDate time.Time
	for _, leg := range strategy.Legs {
		if leg.Expiry.After(expiryDate) {
			expiryDate = leg.Expiry
		}
	}

	if expiryDate.IsZero() {
		expiryDate = time.Now().AddDate(0, 0, 30) // Fallback
	}

	startDate := time.Now().AddDate(0, 0, 1) // "Tomorrow"

	if startDate.After(expiryDate) {
		startDate = time.Now()
	}

	totalDays := expiryDate.Sub(startDate).Hours() / 24
	if totalDays <= 0 {
		totalDays = 0.01
	}

	// Generate 8 Time Slices
	timeSlices := 8
	var dates []time.Time

	stepDays := totalDays / float64(timeSlices-1)
	if timeSlices == 1 {
		dates = append(dates, expiryDate)
	} else {
		for i := 0; i < timeSlices; i++ {
			d := startDate.Add(time.Duration(float64(i)*stepDays*24) * time.Hour)
			if d.After(expiryDate) {
				d = expiryDate
			}
			dates = append(dates, d)
		}
	}

	// 2. Generate 21 Price Points (-20% to +20%)
	pricePoints := 21
	minPrice := currentPrice * 0.80
	maxPrice := currentPrice * 1.20
	priceStep := (maxPrice - minPrice) / float64(pricePoints-1)

	var prices []float64
	for i := 0; i < pricePoints; i++ {
		prices = append(prices, minPrice+float64(i)*priceStep)
	}

	grid := []MatrixPoint{}
	riskFreeRate := 0.05

	// 3. Calculation Loop
	for _, d := range dates {
		// Time to expiry from 'd' (simulated date)
		timeToSimDate := d.Sub(time.Now()).Hours() / 24 / 365.0
		if timeToSimDate < 0.0001 {
			timeToSimDate = 0.0001
		}

		for _, p := range prices {
			totalPnL := 0.0

			for _, leg := range strategy.Legs {
				// Time remaining for this leg from simulated date 'd'
				T_rem := leg.Expiry.Sub(d).Hours() / 24 / 365.0

				var optionValue float64
				if T_rem <= 0 {
					// Expired Value
					if leg.Type == Call {
						optionValue = math.Max(0, p-leg.Strike)
					} else { // Put
						optionValue = math.Max(0, leg.Strike-p)
					}
				} else {
					// Black-Scholes Value
					sigma := leg.IV
					if sigma <= 0 {
						sigma = volatility
					}
					// CalculateTheoreticalPrice
					price, _, _, _, _ := CalculateOptionPrice(leg.Type, p, leg.Strike, T_rem, riskFreeRate, sigma)
					optionValue = price
				}

				// The PnL Formula (Per Leg)
				// Profit = (ExitValue - EntryValue) * Sign * Multiplier
				// Or specifically as requested:
				// Long Call/Put: Value - Cost
				// Short Call/Put: Cost - Value

				// Standardize:
				// CostBasis = EntryPrice * 100 * Quantity
				// CurrentValue = OptionValue * 100 * Quantity

				// Apply 100x Multiplier
				entryVal := leg.EntryPrice * 100
				exitVal := optionValue * 100

				var legProfit float64

				if leg.Action == "Buy" {
					// Long: Profit = Exit - Entry
					legProfit = (exitVal - entryVal) * leg.Quantity
				} else {
					// Short: Profit = Entry - Exit
					legProfit = (entryVal - exitVal) * leg.Quantity
				}

				totalPnL += legProfit
			}

			// 4. Probability Layer (Z-Score)
			denom := currentPrice * volatility * math.Sqrt(timeToSimDate)
			zScore := 0.0
			if denom > 0 {
				zScore = (p - currentPrice) / denom
			}

			// Add to grid
			grid = append(grid, MatrixPoint{
				Date:   d.Format("2006-01-02"),
				Price:  math.Round(p*100) / 100,
				Profit: math.Round(totalPnL*100) / 100,
				ZScore: math.Round(zScore*1000) / 1000,
			})
		}
	}

	return MatrixResponse{Grid: grid}, nil
}
