package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type ExchangeRateResponse struct {
	USDBRL ExchangeRate `json:"USDBRL"`
}

type ExchangeRate struct {
	Code       string `json:"code"`
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
			http.Error(w, "Failed to fetch data from API", http.StatusInternalServerError)
			return
		}
		defer req.Body.Close()

		var exchangeRateResponse ExchangeRateResponse
		if err := json.NewDecoder(req.Body).Decode(&exchangeRateResponse); err != nil {
			http.Error(w, "Failed to parse JSON response: "+err.Error(), http.StatusInternalServerError)
			return
		}

		exchangeRate := exchangeRateResponse.USDBRL

		err = insertIntoDB(exchangeRate)
		if err != nil {
			http.Error(w, "Failed to insert data into database", http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte(exchangeRate.Bid + "\n"))
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	case <-ctx.Done():
		http.Error(w, "Request canceled or timed out", http.StatusRequestTimeout)
	}
}

func insertIntoDB(r ExchangeRate) error {
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/clientserverapi")
	if err != nil {
		return err
	}
	defer db.Close()
	stmt, err := db.Prepare("INSERT INTO exchange_rate (id, codein, name, high, low, var_bid, pct_change, bid, ask, timestamp, createDate) VALUES (?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(uuid.New().String(), r.Codein, r.Name, r.High, r.Low, r.VarBid, r.PctChange, r.Bid, r.Ask, r.Timestamp, r.CreateDate)
	if err != nil {
		return err
	}
	return nil
}
