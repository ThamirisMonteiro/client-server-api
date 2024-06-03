package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
	"net/http"
	"os"
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

		createDBFileIfNotExists()

		db := connectToDB()
		defer db.Close()

		createTableIfNotExists(db)

		err = insertIntoDB(db, exchangeRate)
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

func createTableIfNotExists(db *sql.DB) {
	if !tableExists(db, "exchange_rate") {
		createTableSQL := `
			CREATE TABLE exchange_rate (
				id VARCHAR(255),
				codein VARCHAR(255),
				name VARCHAR(255),
				high VARCHAR(255),
				low VARCHAR(255),
				var_bid VARCHAR(255),
				pct_change VARCHAR(255),
				bid VARCHAR(255),
				ask VARCHAR(255),
				timestamp VARCHAR(255),
				createDate VARCHAR(255)
			);
		`

		stmt, err := db.Prepare(createTableSQL)
		if err != nil {
			panic(err)
		}
		defer stmt.Close()

		_, err = stmt.Exec()
		if err != nil {
			panic(err)
		}
	} else {
		println("Table already exists.")
	}
}

func connectToDB() *sql.DB {
	db, err := sql.Open("sqlite", "client_server_api.db")
	if err != nil {
		panic(err)
	}
	return db
}

func createDBFileIfNotExists() {
	if _, err := os.Stat("client_server_api.db"); os.IsNotExist(err) {
		file, err := os.Create("client_server_api.db")
		if err != nil {
			panic(err)
		}
		file.Close()
	}
}

func insertIntoDB(db *sql.DB, r ExchangeRate) error {
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

func tableExists(db *sql.DB, tableName string) bool {
	var tableExists bool
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"
	err := db.QueryRow(query, tableName).Scan(&tableName)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}
	tableExists = err != sql.ErrNoRows
	return tableExists
}
