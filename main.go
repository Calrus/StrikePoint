package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strikelogic/calculator"
	"strikelogic/news_engine"
	"strikelogic/newsfeed"
	"strikelogic/storage"
	"strikelogic/strategies"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	storage.InitDB()

	// Background news fetcher
	go func() {
		tickers := []string{"TSLA", "NVDA", "SPY"}
		for {
			for _, ticker := range tickers {
				log.Printf("Fetching news for %s...", ticker)
				newsfeed.FetchStockNews(ticker)
			}
			time.Sleep(15 * time.Minute)
		}
	}()

	http.HandleFunc("/chain", func(w http.ResponseWriter, r *http.Request) {
		// Enable CORS for frontend development convenience
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		ticker := r.URL.Query().Get("ticker")
		if ticker == "" {
			ticker = "SPY" // Default ticker
		}

		chain := calculator.GetChain(ticker)

		if err := json.NewEncoder(w).Encode(chain); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/api/calculate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Ticker       string  `json:"ticker"`
			CurrentPrice float64 `json:"currentPrice"`
			TargetPrice  float64 `json:"targetPrice"`
			Date         string  `json:"date"`
			Sentiment    string  `json:"sentiment"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Fetch Option Chain
		chain, err := calculator.GetOptionsChain(req.Ticker, req.Date)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to fetch options chain: %v", err), http.StatusInternalServerError)
			return
		}

		// Generate Strategies
		// Pass sentiment from request
		trades, err := strategies.GenerateAllStrategies(chain, req.Date, req.Sentiment, req.TargetPrice)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(trades); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/api/quote", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		ticker := r.URL.Query().Get("ticker")
		if ticker == "" {
			http.Error(w, "Ticker required", http.StatusBadRequest)
			return
		}

		price, err := calculator.GetQuote(ticker)
		if err != nil {
			// Fallback
			if ticker == "AAPL" {
				price = 150.0
			} else {
				price = 100.0
			}
		}

		json.NewEncoder(w).Encode(map[string]float64{"price": price})
	})

	http.HandleFunc("/api/news/signals", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		headlines, err := news_engine.FetchHeadlines()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var signals []news_engine.Signal
		for _, headline := range headlines {
			sigs, err := news_engine.AnalyzeSentiment(headline)
			if err != nil {
				log.Printf("Error analyzing sentiment for '%s': %v", headline, err)
				continue
			}
			signals = append(signals, sigs...)
		}

		json.NewEncoder(w).Encode(signals)
	})

	http.HandleFunc("/api/news/track", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		ticker := r.URL.Query().Get("ticker")
		if ticker == "" {
			http.Error(w, "Ticker required", http.StatusBadRequest)
			return
		}

		go func() {
			log.Printf("Manually fetching news for %s...", ticker)
			newsfeed.FetchStockNews(ticker)
		}()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "fetching", "message": fmt.Sprintf("Started fetching news for %s", ticker)})
	})

	http.HandleFunc("/api/news", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		ticker := r.URL.Query().Get("ticker")
		// if ticker is empty, we fetch all

		limitStr := r.URL.Query().Get("limit")
		limit := 20
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
		}

		articles, err := storage.GetLatestNews(ticker, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(articles)
	})

	http.HandleFunc("/api/simulate/matrix", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Input struct matching JS Trade object structure somewhat
		var req struct {
			Strategies []strategies.Trade `json:"strategies"` // Or single trade? JS sends 'trade' object
			Strategy   strategies.Trade   `json:"strategy"`
			Price      float64            `json:"currentPrice"`
			Vol        float64            `json:"volatility"`
		}

		// JS Code: fetchMatrixData(trade, currentPrice, vol)
		// We'll likely send { strategy: trade, currentPrice: ..., volatility: ... }

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Map strategies.Trade to calculator.StrategyInput
		var calcInput calculator.StrategyInput
		calcInput.InitialDebit = req.Strategy.NetDebit

		for _, leg := range req.Strategy.Legs {
			expiry, _ := time.Parse("2006-01-02", leg.Option.Expiry)

			optType := calculator.Call
			if leg.Option.Type == "Put" || leg.Option.Type == calculator.Put {
				optType = calculator.Put
			}

			l := calculator.LegInput{
				Strike:   leg.Option.Strike,
				Type:     optType,
				Action:   string(leg.Action),
				Quantity: float64(leg.Quantity),
				Expiry:   expiry,
				IV:       leg.Option.Vol,
			}
			calcInput.Legs = append(calcInput.Legs, l)
		}

		// If NetDebit is stored as per-share, total debit is NetDebit * 1 (standard 1 lot context) * 100?
		// Wait, loop calculated leg value.
		// calcInput.InitialDebit: logic in calculator uses this to subtract from final value.
		// If leg quantity is 1, and InitialDebit is say $1.50 (per share).
		// CombinedOptionValue will be (Price - K) * 100 * Qty. (e.g. $5 * 100 * 1 = $500).
		// If we subtract $1.50, we get $498.50. This is wrong. It should be subtract $150.
		// So InitialDebit should be scaled by 100 if it's per-share price.
		// However, in strategies.go, NetDebit = Ask - Bid (per share).
		// So we should multiply by 100 here.
		calcInput.InitialDebit = calcInput.InitialDebit * 100

		matrix, err := calculator.CalculateProfitMatrix(calcInput, req.Price, req.Vol)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(matrix.Grid)
	})

	fmt.Println("Server starting on :8081...")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}
