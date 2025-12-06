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
	Strike   float64
	Type     OptionType // "Call" or "Put"
	Action   string     // "Buy" or "Sell"
	Quantity float64
	Expiry   time.Time
	IV       float64 // Implied Volatility of this specific leg
}

// StrategyInput captures the strategy details for the matrix calculation
type StrategyInput struct {
	Legs         []LegInput
	InitialDebit float64 // Positive for Debit, Negative for Credit
}

// CalculateProfitMatrix generates a heatmap of theoretical profit/loss over time and price
func CalculateProfitMatrix(strategy StrategyInput, currentPrice float64, volatility float64) (MatrixResponse, error) {
	// 1. Identify Time Horizon
	// Find the earliest or latest expiration? distinct "Expiration Date" in prompt implies the main expiry.
	// We'll assume the latest expiry in the legs is the target "Expiration Date" for the graph.
	var expiryDate time.Time
	for _, leg := range strategy.Legs {
		if leg.Expiry.After(expiryDate) {
			expiryDate = leg.Expiry
		}
	}

	if expiryDate.IsZero() {
		// Fallback if no legs or weird data, though unlikely.
		expiryDate = time.Now().AddDate(0, 0, 30)
	}

	startDate := time.Now().AddDate(0, 0, 1) // "Tomorrow"

	// Ensure startDate is before expiryDate
	if startDate.After(expiryDate) {
		startDate = time.Now() // Fallback
	}

	totalDays := expiryDate.Sub(startDate).Hours() / 24
	if totalDays <= 0 {
		totalDays = 0.01 // Avoid div by zero
	}

	// Generate 8 Time Slices
	timeSlices := 8
	var dates []time.Time

	// If totalDays is very small, just do 1 slice? But prompt says 15.
	// We'll interpolate.
	stepDays := totalDays / float64(timeSlices-1)
	if timeSlices == 1 {
		dates = append(dates, expiryDate)
	} else {
		for i := 0; i < timeSlices; i++ {
			d := startDate.Add(time.Duration(float64(i)*stepDays*24) * time.Hour)
			// Clamp to expiry
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
	riskFreeRate := 0.05 // Assumption, could be parameterized

	// 3. Calculation Loop
	for _, d := range dates {
		// Time to expiry from 'd' for the option pricing
		// Note: The option expires at leg.Expiry. 'd' is the simulated current date.
		// T = (leg.Expiry - d) / 365

		// Time elapsed from Now to 'd' for Z-Score
		// T_elapsed = (d - Now)
		timeToSimDate := d.Sub(time.Now()).Hours() / 24 / 365.0
		if timeToSimDate < 0.0001 {
			timeToSimDate = 0.0001
		}

		for _, p := range prices {
			combinedOptionValue := 0.0

			for _, leg := range strategy.Legs {
				// Time remaining for this leg from simulated date 'd'
				T_rem := leg.Expiry.Sub(d).Hours() / 24 / 365.0

				var legValue float64
				if T_rem <= 0 {
					// Expired
					val := 0.0
					if leg.Type == Call {
						if p > leg.Strike {
							val = p - leg.Strike
						}
					} else {
						if p < leg.Strike {
							val = leg.Strike - p
						}
					}
					legValue = val
				} else {
					// Use Black-Scholes
					// Use leg.IV if valid (>0), else use global volatility
					sigma := leg.IV
					if sigma <= 0 {
						sigma = volatility
					}

					price, _, _, _, _ := CalculateOptionPrice(leg.Type, p, leg.Strike, T_rem, riskFreeRate, sigma)
					legValue = price
				}

				// Adjust for Side and Quantity
				qty := leg.Quantity
				// Value of position:
				// Long: +1 * Price * Qty * 100
				// Short: -1 * Price * Qty * 100
				sign := 1.0
				if leg.Action == "Sell" {
					sign = -1.0
				}

				combinedOptionValue += sign * legValue * qty * 100
			}

			// NetProfit = CombinedOptionValue - InitialDebit
			// InitialDebit is (+) for Debit, (-) for Credit.
			netProfit := combinedOptionValue - strategy.InitialDebit

			// 4. Probability Layer (Z-Score)
			// Z-Score = (TargetPrice - CurrentPrice) / (CurrentPrice * Volatility * sqrt(Time))
			denom := currentPrice * volatility * math.Sqrt(timeToSimDate)
			zScore := 0.0
			if denom > 0 {
				zScore = (p - currentPrice) / denom
			}

			// Add to grid
			grid = append(grid, MatrixPoint{
				Date:   d.Format("2006-01-02"),
				Price:  math.Round(p*100) / 100,
				Profit: math.Round(netProfit*100) / 100,
				ZScore: math.Round(zScore*1000) / 1000,
			})
		}
	}

	return MatrixResponse{Grid: grid}, nil
}
