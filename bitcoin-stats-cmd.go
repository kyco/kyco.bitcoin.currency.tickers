package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var err error

// Luno Ticker
type LunoTicker struct {
	Tickers []struct {
		Timestamp           int64  `json:"timestamp"`
		Bid                 string `json:"bid"`
		Ask                 string `json:"ask"`
		LastTrade           string `json:"last_trade"`
		Rolling24HourVolume string `json:"rolling_24_hour_volume"`
		Pair                string `json:"pair"`
	} `json:"tickers"`
}
type Bitstamp struct {
	High      string `json:"high"`
	Last      string `json:"last"`
	Timestamp string `json:"timestamp"`
	Bid       string `json:"bid"`
	Vwap      string `json:"vwap"`
	Volume    string `json:"volume"`
	Low       string `json:"low"`
	Ask       string `json:"ask"`
	Open      string `json:"open"`
}

// Setup global home and config folders
// File location
var fileLocation string = "/.config/bitcoin-stats/"

// Get home folder
var home string = os.Getenv("HOME") + fileLocation

// Initialises various bitcoin price tickers
func bitcoin_prices() {
	for {

		// Start luno ticker
		luno_ticker()
		// Start bitstamp ticker
		bitstamp_ticker()

		time.Sleep(11 * time.Minute)
	}

}

// Grabs a snapshot of the current luno exchange
func luno_ticker() {
	// Make API call to luno
	resp := api_call("https://api.mybitx.com/api/1/tickers")

	// Callers should close resp.Body
	// when done reading from it
	// Defer the closing of the body
	defer resp.Body.Close()

	// Fill the record with the data from the JSON
	var record LunoTicker

	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		log.Println(err)
	}

	// Write to DB
	// Loop through the slice
	for i := range record.Tickers {
		// Format timestamp as string
		timestampString := strconv.FormatInt(record.Tickers[i].Timestamp, 10)
		insert_into_sqlite("Luno", timestampString, record.Tickers[i].Ask, record.Tickers[i].Bid, "0", record.Tickers[i].Pair[3:])
	}
}

// Grabs a snapshot of the current bitstamp exchange
func bitstamp_ticker() {
	// Make API call to bitstamp
	resp := api_call("https://www.bitstamp.net/api/v2/ticker_hour/btcusd/")

	// Callers should close resp.Body
	// when done reading from it
	// Defer the closing of the body
	defer resp.Body.Close()

	// Fill the record with the data from the JSON
	var record Bitstamp

	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		log.Println(err)
	}
	// Insert into SQlite
	insert_into_sqlite("Bitstamp", record.Timestamp, record.Ask, record.Bid, record.Volume, "USD")
}

// performs an API call to a URL and returns a JSON body response
func api_call(urlRequest string) *http.Response {

	url := fmt.Sprintf(urlRequest)

	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return nil
	}

	// For control over HTTP client headers,
	// redirect policy, and other settings,
	// create a Client
	// A Client is an HTTP client
	client := &http.Client{}

	// Send the request via a client
	// Do sends an HTTP request and
	// returns an HTTP response
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return nil
	}

	// // Callers should close resp.Body
	// // when done reading from it
	// // Defer the closing of the body
	// defer resp.Body.Close()

	// Return the body
	return resp
}

// Check for and print/panic on errors
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Open SQlite Connection
func sqlite_open() *sql.DB {
	db, err := sql.Open("sqlite3", home+"/data.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// Sets up the sqlite databases and connections
func setup_sqlite_db() {
	// os.Remove(home + "/data.db")
	// Check if the sqlite database already exists, if it does not, continue
	// else, don't care
	if _, err := os.Stat(home + "/data.db"); os.IsNotExist(err) {
		sqliteDB := sqlite_open()

		sqlStmt := `create table exchanges (id integer not null primary key, exchange text, timestamp real, ask real, bid real, volume real default 0, currencyCode text);`
		_, err = sqliteDB.Exec(sqlStmt)
		if err != nil {
			log.Printf("%q: %s\n", err, sqlStmt)
			return
		}
	}
}

// Insert function into sqlite
func insert_into_sqlite(exchange string, timestamp string, ask string, bid string, volume string, currencyCode string) {

	// Write to DB
	sqliteDB := sqlite_open()
	// Insert the database record
	sqlStmt := `insert into exchanges (exchange, timestamp, ask, bid, volume, currencyCode) values ('` + exchange + `',` + timestamp + `, ` + ask + `,` + bid + `, ` + volume + `, '` + currencyCode + `');`
	_, err = sqliteDB.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
	// Close the sqlite connection
	sqliteDB.Close()
}

func main() {
	// Logging
	// open a file
	f, err := os.OpenFile("/tmp/bitcoin-stats.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Printf("error opening file: %v", err)
	}

	// don't forget to close it
	defer f.Close()

	// assign it to the standard logger
	log.SetOutput(f)

	// Initialise config file and settings
	// config_init()

	// Setup Sqlite DB
	setup_sqlite_db()

	// Start bitcoin ticker
	go bitcoin_prices()

	select {}
}
