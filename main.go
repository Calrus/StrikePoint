package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strikelogic/calculator"
	"strikelogic/strategist"
)

func main() {
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
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		trades, err := strategist.FindTrades(req.Ticker, req.CurrentPrice, req.TargetPrice, req.Date)
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

	fmt.Println("Server starting on :8081...")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}
