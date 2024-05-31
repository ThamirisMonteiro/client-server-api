package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type ExchangeRate struct {
	USDBRL struct {
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	http.HandleFunc("/cotacao", handler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	select {
	case <-time.After(time.Millisecond * 200):
		req, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
		if err != nil {
			panic(err)
		}
		defer req.Body.Close()

		var exchangeRate ExchangeRate
		if err := json.NewDecoder(req.Body).Decode(&exchangeRate); err != nil {
			panic(err)
		}

		_, err = w.Write([]byte(exchangeRate.USDBRL.Bid))
		if err != nil {
			return
		}
		w.WriteHeader(http.StatusOK)

	case <-ctx.Done():
		http.Error(w, "Request canceled or timed out", http.StatusRequestTimeout)
	}
}
