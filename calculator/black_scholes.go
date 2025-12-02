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
	Vega       float64    `json:"vega"`
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

// CalculateOptionPrice calculates Price and Greeks for Call or Put
// S: Current Price, K: Strike Price, T: Time to Expiry (years), r: Risk-Free Rate, sigma: Implied Volatility
func CalculateOptionPrice(optType OptionType, S, K, T, r, sigma float64) (price, delta, gamma, theta, vega float64) {
	d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / (sigma * math.Sqrt(T))
	d2 := d1 - sigma*math.Sqrt(T)

	// Common Greeks
	gamma = pdf(d1) / (S * sigma * math.Sqrt(T))
	vega = S * pdf(d1) * math.Sqrt(T) / 100 // Divide by 100 for percentage change

	if optType == Call {
		price = S*cdf(d1) - K*math.Exp(-r*T)*cdf(d2)
		delta = cdf(d1)

		term1 := -(S * pdf(d1) * sigma) / (2 * math.Sqrt(T))
		term2 := r * K * math.Exp(-r*T) * cdf(d2)
		theta = (term1 - term2) / 365 // Daily Theta
	} else {
		price = K*math.Exp(-r*T)*cdf(-d2) - S*cdf(-d1)
		delta = cdf(d1) - 1

		term1 := -(S * pdf(d1) * sigma) / (2 * math.Sqrt(T))
		term2 := r * K * math.Exp(-r*T) * cdf(-d2)
		theta = (term1 + term2) / 365 // Daily Theta
	}

	return price, delta, gamma, theta, vega
}

// CalculateIV solves for Implied Volatility using Newton-Raphson
func CalculateIV(marketPrice float64, optType OptionType, S, K, T, r float64) float64 {
	sigma := 0.5 // Initial guess
	tolerance := 1e-5
	maxIterations := 100

	for i := 0; i < maxIterations; i++ {
		price, _, _, _, vega := CalculateOptionPrice(optType, S, K, T, r, sigma)
		diff := marketPrice - price

		if math.Abs(diff) < tolerance {
			return sigma
		}

		// Avoid division by zero
		if vega == 0 {
			break
		}

		// Vega in CalculateOptionPrice is scaled by 100, so we need to adjust or use raw vega
		// Let's recalculate raw vega here for precision in NR
		d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / (sigma * math.Sqrt(T))
		rawVega := S * pdf(d1) * math.Sqrt(T)

		sigma = sigma + diff/rawVega
	}

	return sigma
}

// CalculatePoP calculates Probability of Profit
// Formula: PoP = 1 - NormCDF( (log(BreakEven / StockPrice) - (Volatility^2 / 2) * Time) / (Volatility * sqrt(Time)) )
func CalculatePoP(breakEven, S, T, sigma float64) float64 {
	if T <= 0 || sigma <= 0 {
		return 0
	}
	d := (math.Log(breakEven/S) - (0.5*sigma*sigma)*T) / (sigma * math.Sqrt(T))
	return 1 - cdf(d)
}

// CalculateReturnOnRiskDebit calculates RoR for Debit Spreads
func CalculateReturnOnRiskDebit(maxProfit, netDebit float64) float64 {
	if netDebit <= 0 {
		return 0 // Avoid division by zero or negative debit weirdness
	}
	return (maxProfit / netDebit) * 100
}

// CalculateReturnOnRiskCredit calculates RoR for Credit Spreads
func CalculateReturnOnRiskCredit(netCredit, maxRisk float64) float64 {
	if maxRisk <= 0 {
		return 0
	}
	return (netCredit / maxRisk) * 100
}

// BlackScholes wrapper for backward compatibility if needed, defaulting to Call
func BlackScholes(S, K, T, r, sigma float64) (price, delta, gamma, theta float64) {
	p, d, g, t, _ := CalculateOptionPrice(Call, S, K, T, r, sigma)
	return p, d, g, t
}

// GetChain generates a mock option chain for a given ticker (default ~30 days)
func GetChain(ticker string) []OptionContract {
	return GetChainForExpiry(ticker, 30)
}

// GetChainForExpiry generates a mock option chain for a given ticker and days to expiry
func GetChainForExpiry(ticker string, daysOut int) []OptionContract {
	currentPrice := 100.0 // Base price for mock
	if ticker == "GOOG" {
		currentPrice = 2800.0
	} else if ticker == "AMZN" {
		currentPrice = 3400.0
	} else if ticker == "TSLA" {
		currentPrice = 900.0
	} else if ticker == "SPY" {
		currentPrice = 450.0
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

		// Call
		cPrice, cDelta, cGamma, cTheta, cVega := CalculateOptionPrice(Call, currentPrice, k, T, r, iv)

		// Spread logic
		spread := cPrice * 0.02
		cBid := cPrice - spread/2
		cAsk := cPrice + spread/2
		if cBid < 0 {
			cBid = 0
		}

		callContract := OptionContract{
			Strike:     math.Round(k*100) / 100,
			Expiry:     time.Now().AddDate(0, 0, daysOut).Format("2006-01-02"),
			Type:       Call,
			Bid:        math.Round(cBid*100) / 100,
			Ask:        math.Round(cAsk*100) / 100,
			Last:       math.Round(cPrice*100) / 100,
			Vol:        math.Round(iv*100) / 100,
			Delta:      math.Round(cDelta*1000) / 1000,
			Gamma:      math.Round(cGamma*1000) / 1000,
			Theta:      math.Round(cTheta*1000) / 1000,
			Vega:       math.Round(cVega*1000) / 1000,
			Underlying: ticker,
		}
		chain = append(chain, callContract)

		// Put
		pPrice, pDelta, pGamma, pTheta, pVega := CalculateOptionPrice(Put, currentPrice, k, T, r, iv)

		pBid := pPrice - spread/2
		pAsk := pPrice + spread/2
		if pBid < 0 {
			pBid = 0
		}

		putContract := OptionContract{
			Strike:     math.Round(k*100) / 100,
			Expiry:     time.Now().AddDate(0, 0, daysOut).Format("2006-01-02"),
			Type:       Put,
			Bid:        math.Round(pBid*100) / 100,
			Ask:        math.Round(pAsk*100) / 100,
			Last:       math.Round(pPrice*100) / 100,
			Vol:        math.Round(iv*100) / 100,
			Delta:      math.Round(pDelta*1000) / 1000,
			Gamma:      math.Round(pGamma*1000) / 1000,
			Theta:      math.Round(pTheta*1000) / 1000,
			Vega:       math.Round(pVega*1000) / 1000,
			Underlying: ticker,
		}
		chain = append(chain, putContract)
	}

	return chain
}
