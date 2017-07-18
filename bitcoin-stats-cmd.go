package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Beldur/kraken-go-api-client"
	"github.com/fsnotify/fsnotify"
	_ "github.com/mattn/go-sqlite3"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
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

// Config type
type Config struct {
	LogFile        string
	SqliteLocation string
	Kraken         KrakenConfig
	Luno           LunoConfig
	Bitstamp       BitstampConfig
}

type KrakenConfig struct {
	URL       string
	APIKey    string
	APISecret string
}

type LunoConfig struct {
	URL string
}

type BitstampConfig struct {
	URL string
}

var config Config
var logFile *os.File
var log = logging.MustGetLogger("bitcoin-logger")

// Example format string. Everything except the message has a custom color
// which is dependent on the log level. Many fields have a custom output
// formatting too, eg. the time returns the hour down to the milli second.
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

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
		log.Notice("Started Luno Ticker")

		// Start bitstamp ticker
		bitstamp_ticker()
		log.Notice("Started Bitstamp Ticker")

		// Start kraken ticker
		kraken_ticker()
		log.Notice("Started Kraken Ticker")

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
		log.Error(err)
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
		log.Error(err)
	}
	// Insert into SQlite
	insert_into_sqlite("Bitstamp", record.Timestamp, record.Ask, record.Bid, record.Volume, "USD")
}

// Gets ticker data from kraken
func kraken_ticker() {
	// If the API keys are not present, just return
	if len(config.Kraken.APIKey) == 0 && len(config.Kraken.APISecret) == 0 {
		return
	}

	api := krakenapi.New(config.Kraken.APIKey, config.Kraken.APISecret)

	// There are also some strongly typed methods available
	tickerEUR, err := api.Ticker(krakenapi.XXBTZEUR)
	if err != nil {
		log.Warning(err)
	}

	// Insert into SQlite
	insert_into_sqlite("Kraken", strconv.FormatInt(int64(time.Now().Unix()), 10), tickerEUR.XXBTZEUR.Ask[0], tickerEUR.XXBTZEUR.Bid[0], tickerEUR.XXBTZEUR.Volume[0], "EUR")

	// There are also some strongly typed methods available
	tickerUSD, err := api.Ticker(krakenapi.XXBTZUSD)
	if err != nil {
		log.Warning(err)
	}

	// Insert into SQlite
	insert_into_sqlite("Kraken", strconv.FormatInt(int64(time.Now().Unix()), 10), tickerUSD.XXBTZUSD.Ask[0], tickerUSD.XXBTZUSD.Bid[0], tickerUSD.XXBTZUSD.Volume[0], "USD")

	// There are also some strongly typed methods available
	tickerGBP, err := api.Ticker(krakenapi.XXBTZGBP)
	if err != nil {
		log.Warning(err)
	}

	// Insert into SQlite
	insert_into_sqlite("Kraken", strconv.FormatInt(int64(time.Now().Unix()), 10), tickerGBP.XXBTZGBP.Ask[0], tickerGBP.XXBTZGBP.Bid[0], tickerGBP.XXBTZGBP.Volume[0], "GBP")

}

// performs an API call to a URL and returns a JSON body response
func api_call(urlRequest string) *http.Response {

	url := fmt.Sprintf(urlRequest)

	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error("NewRequest: ", err)
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
		log.Error("Do: ", err)
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
		log.Error(e)
	}
}

// Open SQlite Connection
func sqlite_open() *sql.DB {
	db, err := sql.Open("sqlite3", home+"/data.db")
	if err != nil {
		log.Error(err)
	}
	return db
}

// Sets up the sqlite databases and connections
func setup_sqlite_db() {
	// Setup sqlite connection
	var sqliteConnection string

	// If the config file line is empty, then default
	if len(config.SqliteLocation) > 0 {
		sqliteConnection = config.SqliteLocation
	} else {
		sqliteConnection = home + "/data.db"
	}
	// Check if the sqlite database already exists, if it does not, continue
	// else, don't care
	if _, err := os.Stat(sqliteConnection); os.IsNotExist(err) {
		sqliteDB := sqlite_open()

		sqlStmt := `create table exchanges (id integer not null primary key, exchange text, timestamp real, ask real, bid real, volume real default 0, currencyCode text);`
		_, err = sqliteDB.Exec(sqlStmt)
		if err != nil {
			log.Warning("%q: %s\n", err, sqlStmt)
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
		log.Warning("%q: %s\n", err, sqlStmt)
		return
	}
	// Close the sqlite connection
	sqliteDB.Close()
}

// Configure logging
func config_log() {

	// Check if it is already open
	logFile.Close()

	// Configure logging
	logFile, err := os.OpenFile(config.LogFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Info("error opening file: %v", err)
	}

	// For demo purposes, create two backend for os.Stderr.
	loggingFile := logging.NewLogBackend(logFile, "", 0)

	// For messages written to loggingFile we want to add some additional
	// information to the output, including the used log level and the name of
	// the function.
	loggingFileFormatter := logging.NewBackendFormatter(loggingFile, format)

	// Set the backends to be used.
	logging.SetBackend(loggingFileFormatter)
}

// Configure configs
func config_init() {

	// Config File
	viper.SetConfigName("config") // no need to include file extension
	viper.AddConfigPath(home)     // set the path of your config file

	err := viper.ReadInConfig()
	if err != nil {
		log.Info("Config file not found... Error %s\n", err)
	} else {

		// ========= CONFIG ================================================================
		logFile := viper.GetString("config.logFile")
		sqliteLocation := viper.GetString("config.sqliteLocation")
		krakenurl := viper.GetString("exchanges.kraken.url")
		krakenAPIKey := viper.GetString("exchanges.kraken.apiKey")
		krakenAPISecret := viper.GetString("exchanges.kraken.apiSecret")
		lunourl := viper.GetString("exchanges.luno.url")
		bitstampurl := viper.GetString("exchanges.bitstamp.url")

		// Kraken
		kraken := KrakenConfig{
			URL:       krakenurl,
			APIKey:    krakenAPIKey,
			APISecret: krakenAPISecret,
		}

		// Luno
		luno := LunoConfig{
			URL: lunourl,
		}

		// Bitstamp
		bitstamp := BitstampConfig{
			URL: bitstampurl,
		}

		// Main Config
		config = Config{
			LogFile:        logFile,
			SqliteLocation: sqliteLocation,
			Kraken:         kraken,
			Luno:           luno,
			Bitstamp:       bitstamp,
		}
	}

	// Monitor the config file for changes and reload
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {

		// Re-configure config
		config_init()

		// Re-configure logging
		config_log()

		log.Info("Config file changed:", e.Name)
	})
}

func main() {

	// Initialise config file and settings
	config_init()

	// Configure logging
	config_log()

	// don't forget to close the log file
	defer logFile.Close()

	// Setup Sqlite DB
	setup_sqlite_db()

	// Start bitcoin ticker
	go bitcoin_prices()

	// Notify log that we are up and running
	log.Info("started Bitcoin Stats")

	select {}
}
