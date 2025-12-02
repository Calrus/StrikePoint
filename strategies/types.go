package strategies

import (
	"strikelogic/calculator"
)

type Action string

const (
	Buy  Action = "Buy"
	Sell Action = "Sell"
)

type TradeLeg struct {
	Action     Action                    `json:"action"`
	Quantity   int                       `json:"quantity"`
	IsStock    bool                      `json:"isStock"`              // True if this leg is the underlying stock
	StockPrice float64                   `json:"stockPrice,omitempty"` // Price of the stock for stock legs
	Option     calculator.OptionContract `json:"option"`
}

type Trade struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Sentiment   string     `json:"sentiment"` // "Bullish", "Bearish", "Neutral"
	Legs        []TradeLeg `json:"legs"`

	// Calculated fields
	MaxProfit  float64   `json:"maxProfit"`
	MaxRisk    float64   `json:"maxRisk"` // Positive number representing max loss
	BreakEvens []float64 `json:"breakEvens"`

	// Greeks (Portfolio)
	Delta float64 `json:"delta"`
	Gamma float64 `json:"gamma"`
	Theta float64 `json:"theta"`
	Vega  float64 `json:"vega"`

	// Cost
	NetDebit float64 `json:"netDebit"` // Positive for debit, negative for credit
}

// StrategyRecipe contains the rules for constructing a trade
type StrategyRecipe struct {
	Name        string
	Description string
	// Builder is a function that attempts to construct the trade from the chain
	Builder func(chain []calculator.OptionContract, targetPrice float64) *Trade
}
