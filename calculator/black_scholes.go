package calculator

import (
	"math"
	"math/rand"
	"time"
)

// OptionType defines the type of option (Call or Put)
type OptionType string

const (
	Call OptionType = "Call"
	Put  OptionType = "Put"
)

// OptionContract holds details about an option contract
type OptionContract struct {
	Strike     float64    `json:"strike"`
	Expiry     string     `json:"expiry"` // ISO date string
	Type       OptionType `json:"type"`
	Bid        float64    `json:"bid"`
	Ask        float64    `json:"ask"`
	Last       float64    `json:"last"`
	Vol        float64    `json:"vol"` // Implied Volatility
	Delta      float64    `json:"delta"`
	Gamma      float64    `json:"gamma"`
	Theta      float64    `json:"theta"`
	Underlying string     `json:"underlying"`
}

// cumulativeDistributionFunction for standard normal distribution
func cdf(x float64) float64 {
	return 0.5 * (1 + math.Erf(x/math.Sqrt2))
}

// probabilityDensityFunction for standard normal distribution
func pdf(x float64) float64 {
	return (1 / math.Sqrt(2*math.Pi)) * math.Exp(-0.5*x*x)
}

// BlackScholes calculates the Call Price and Greeks
// S: Current Price, K: Strike Price, T: Time to Expiry (years), r: Risk-Free Rate, sigma: Implied Volatility
func BlackScholes(S, K, T, r, sigma float64) (price, delta, gamma, theta float64) {
	d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / (sigma * math.Sqrt(T))
	d2 := d1 - sigma*math.Sqrt(T)

	// Call Price
	price = S*cdf(d1) - K*math.Exp(-r*T)*cdf(d2)

	// Delta
	delta = cdf(d1)

	// Gamma
	gamma = pdf(d1) / (S * sigma * math.Sqrt(T))

	// Theta (Annual)
	// Note: Usually theta is negative for long calls
	term1 := -(S * pdf(d1) * sigma) / (2 * math.Sqrt(T))
	term2 := r * K * math.Exp(-r*T) * cdf(d2)
	theta = term1 - term2

	return price, delta, gamma, theta
}

// GetChain generates a mock option chain for a given ticker (default ~30 days)
func GetChain(ticker string) []OptionContract {
	return GetChainForExpiry(ticker, 30)
}

// GetChainForExpiry generates a mock option chain for a given ticker and days to expiry
func GetChainForExpiry(ticker string, daysOut int) []OptionContract {
	// Seed random for variability (in a real app, do this once in main)
	// But for this function to be self-contained mock, we'll just use global rand which is seeded by 1 by default or we can seed here.
	// Let's just use default rand for reproducibility or simple mock.

	currentPrice := 100.0 // Base price for mock
	if ticker == "GOOG" {
		currentPrice = 2800.0
	} else if ticker == "AMZN" {
		currentPrice = 3400.0
	} else if ticker == "TSLA" {
		currentPrice = 900.0
	}

	r := 0.05                     // 5% risk free rate
	T := float64(daysOut) / 365.0 // Convert days to years

	var chain []OptionContract

	// Generate strikes around current price
	// +/- 20%
	startStrike := currentPrice * 0.8
	endStrike := currentPrice * 1.2

	// Round startStrike to nearest 5
	startStrike = math.Round(startStrike/5) * 5

	for k := startStrike; k <= endStrike; k += 5 {
		// Randomize IV slightly between 20% and 40%
		iv := 0.20 + rand.Float64()*0.20

		price, delta, gamma, theta := BlackScholes(currentPrice, k, T, r, iv)

		// Spread logic
		spread := price * 0.02 // 2% spread
		bid := price - spread/2
		ask := price + spread/2
		if bid < 0 {
			bid = 0
		}

		contract := OptionContract{
			Strike:     math.Round(k*100) / 100,
			Expiry:     time.Now().AddDate(0, 0, daysOut).Format("2006-01-02"),
			Type:       Call,
			Bid:        math.Round(bid*100) / 100,
			Ask:        math.Round(ask*100) / 100,
			Last:       math.Round(price*100) / 100,
			Vol:        math.Round(iv*100) / 100,
			Delta:      math.Round(delta*1000) / 1000,
			Gamma:      math.Round(gamma*1000) / 1000,
			Theta:      math.Round(theta*1000) / 1000,
			Underlying: ticker,
		}
		chain = append(chain, contract)
	}

	return chain
}
