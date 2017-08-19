package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Beldur/kraken-go-api-client"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/mux"
	"github.com/jyap808/go-poloniex"
	_ "github.com/mattn/go-sqlite3"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var err error

/*
	STRUCTS CONFIGURED IN structs.go
*/

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
var fileLocation string = "/.config/kyco.bitcoin.currency.tickers/"

// Get home folder
var home string = os.Getenv("HOME") + fileLocation

// Get Exchange rate based on an API call
func get_exchange_rate(w http.ResponseWriter, req *http.Request) {

	var (
		params = mux.Vars(req)
	)

	data, err := query_exchange_sqlite(params["exchange"], params["currencyCode"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Infof("Called: %s -> %s\n", params["exchange"], params["currencyCode"])

	json.NewEncoder(w).Encode(data)
}

// Get Exchange data based on an API call
func show_exchange_methods(w http.ResponseWriter, req *http.Request) {

	var (
		params = mux.Vars(req)
		url    = ""
	)

	// Get the URL
	url = req.Host + "/" + params["exchange"] + "/"

	// Grab exchange data
	data, err := query_exchange_currency_codes_sqlite(params["exchange"], url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Infof("Called: %s -> %s\n", params["exchange"])

	// Return exchange data
	json.NewEncoder(w).Encode(data)
}

// Get list of exchanges
func show_exchanges(w http.ResponseWriter, req *http.Request) {

	var (
		url = ""
	)

	// Get the URL
	url = req.Host + "/"

	// Grab exchange data
	data, err := query_list_of_exchanges(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Infof("Called Root Page")

	// Return exchange data
	json.NewEncoder(w).Encode(data)
}

// Initialises various bitcoin price tickers
func bitcoin_prices() {

	// Tick on the minute
	// t := minuteTicker()

	for {

		// wait for the tick
		// <-t.C

		// Start luno ticker
		luno_ticker()
		log.Notice("Ran Luno Ticker")

		// Start bitstamp ticker
		bitstamp_ticker()
		log.Notice("Ran Bitstamp Ticker")

		// Start kraken ticker
		kraken_ticker()
		log.Notice("Ran Kraken Ticker")

		// Start bitfinex ticker
		bitfinex_ticker()
		log.Notice("Ran Bitfinex Ticker")

		// Start bitsquare ticker
		bitsquare_ticker()
		log.Notice("Ran Bitsquare Ticker")

		// Start btcc ticker
		btcc_ticker()
		log.Notice("Ran BTCChina Ticker")

		// Start okcoin ticker
		okcoin_ticker()
		log.Notice("Ran OKCoin Ticker")

		// Start poloniex ticker
		poloniex_ticker()
		log.Notice("Ran Poloniex Ticker")

		time.Sleep(10 * time.Minute)

	}

}

// Grabs a snapshot of the current luno exchange
func luno_ticker() {
	// Make API call to luno
	resp := api_call(config.Luno.URL)

	// Callers should close resp.Body
	// when done reading from it
	// Defer the closing of the body
	defer resp.Body.Close()

	// Fill the record with the data from the JSON
	var record LunoTicker

	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		log.Error(err.Error())
	} else {
		// Write to DB
		// Loop through the slice
		for i := range record.Tickers {
			// Format timestamp as string
			timestampString := strconv.FormatInt(time.Now().Unix(), 10)
			insert_into_sqlite("Luno", timestampString, record.Tickers[i].Ask, record.Tickers[i].Bid, "1", record.Tickers[i].Pair[3:])
		}
	}
}

// Grabs a snapshot of the current bitstamp exchange
func bitstamp_ticker() {
	// Make API call to bitstamp
	resp := api_call(config.Bitstamp.URL)

	// Callers should close resp.Body
	// when done reading from it
	// Defer the closing of the body
	defer resp.Body.Close()

	// Fill the record with the data from the JSON
	var record Bitstamp

	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		log.Error(err.Error())
	} else {
		// Insert into SQlite
		insert_into_sqlite("Bitstamp", record.Timestamp, record.Ask, record.Bid, record.Volume, "USD")
	}
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
		log.Error(err.Error())
	} else {
		// Insert into SQlite
		insert_into_sqlite("Kraken", strconv.FormatInt(int64(time.Now().Unix()), 10), tickerEUR.XXBTZEUR.Ask[0], tickerEUR.XXBTZEUR.Bid[0], tickerEUR.XXBTZEUR.Volume[0], "EUR")
	}

	// There are also some strongly typed methods available
	tickerUSD, err := api.Ticker(krakenapi.XXBTZUSD)
	if err != nil {
		log.Error(err.Error())
	} else {
		// Insert into SQlite
		insert_into_sqlite("Kraken", strconv.FormatInt(int64(time.Now().Unix()), 10), tickerUSD.XXBTZUSD.Ask[0], tickerUSD.XXBTZUSD.Bid[0], tickerUSD.XXBTZUSD.Volume[0], "USD")
	}

	// There are also some strongly typed methods available
	tickerGBP, err := api.Ticker(krakenapi.XXBTZGBP)
	if err != nil {
		log.Error(err.Error())
	} else {
		// Insert into SQlite
		insert_into_sqlite("Kraken", strconv.FormatInt(int64(time.Now().Unix()), 10), tickerGBP.XXBTZGBP.Ask[0], tickerGBP.XXBTZGBP.Bid[0], tickerGBP.XXBTZGBP.Volume[0], "GBP")
	}
}

// Grabs a snapshot of the current bitfinex exchange
func bitfinex_ticker() {
	// In this case, we will loop through all
	// the tickers set in the config file
	tickerSplit := strings.Split(config.Bitfinex.Tickers, ",")

	for i := range tickerSplit {

		// Check if there is any data in the string
		// if not, skip this loop
		if len(tickerSplit[i]) < 2 {
			continue
		}

		// Make API call to bitfinex
		resp := api_call(config.Bitfinex.URL + tickerSplit[i])

		// Callers should close resp.Body
		// when done reading from it
		// Defer the closing of the body
		defer resp.Body.Close()

		// Fill the record with the data from the JSON
		var record Bitfinex

		// Use json.Decode for reading streams of JSON data
		if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
			log.Error(err.Error())
		} else {
			// Insert into SQlite
			insert_into_sqlite("Bitfinex", record.Timestamp, record.Ask, record.Bid, record.Volume, format_currency_string(tickerSplit[i]))
		}
	}
}

// Grabs a snapshot of the current bitsquare exchange
func bitsquare_ticker() {
	// In this case, we will loop through all
	// the tickers set in the config file
	tickerSplit := strings.Split(config.Bitsquare.Tickers, ",")

	for i := range tickerSplit {

		// Check if there is any data in the string
		// if not, skip this loop
		if len(tickerSplit[i]) < 2 {
			continue
		}

		// Make API call to bitsquare
		resp := api_call(config.Bitsquare.URL + tickerSplit[i])

		// Callers should close resp.Body
		// when done reading from it
		// Defer the closing of the body
		defer resp.Body.Close()

		// Fill the record with the data from the JSON
		var record []Bitsquare

		// Use json.Decode for reading streams of JSON data
		if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
			log.Error(err.Error())
		} else {
			// Create a timestamp now
			ts := strconv.FormatInt(int64(time.Now().Unix()), 10)
			// Insert into SQlite
			insert_into_sqlite("Bitsquare", ts, record[0].Sell, record[0].Buy, record[0].VolumeRight, format_currency_string(tickerSplit[i]))
		}
	}
}

// Grabs a snapshot of the current BTCC exchange
func btcc_ticker() {
	// In this case, we will loop through all
	// the tickers set in the config file
	tickerSplit := strings.Split(config.BTCC.Tickers, ",")

	for i := range tickerSplit {

		// Check if there is any data in the string
		// if not, skip this loop
		if len(tickerSplit[i]) < 2 {
			continue
		}

		// Make API call to btcc
		resp := api_call(config.BTCC.URL + tickerSplit[i])

		// Callers should close resp.Body
		// when done reading from it
		// Defer the closing of the body
		defer resp.Body.Close()

		// Fill the record with the data from the JSON
		var record BTCC

		// Use json.Decode for reading streams of JSON data
		if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
			log.Error(err.Error())
		} else {
			// Insert into SQlite
			insert_into_sqlite("BTCChina", strconv.FormatInt((record.Ticker.Timestamp/1000), 10), strconv.FormatFloat(record.Ticker.AskPrice, 'f', 2, 64), strconv.FormatFloat(record.Ticker.BidPrice, 'f', 2, 64), strconv.FormatFloat(record.Ticker.Volume, 'f', 2, 64), format_currency_string(tickerSplit[i]))
		}
	}
}

// Grabs a snapshot of the current OKCoin exchange
func okcoin_ticker() {
	// In this case, we will loop through all
	// the tickers set in the config file
	tickerSplit := strings.Split(config.OKCoin.Tickers, ",")

	for i := range tickerSplit {

		// Check if there is any data in the string
		// if not, skip this loop
		if len(tickerSplit[i]) < 2 {
			continue
		}

		// Make API call to OKCoin
		resp := api_call(config.OKCoin.URL + tickerSplit[i])

		// Callers should close resp.Body
		// when done reading from it
		// Defer the closing of the body
		defer resp.Body.Close()

		// Fill the record with the data from the JSON
		var record OKCoin

		// Use json.Decode for reading streams of JSON data
		if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
			log.Error(err.Error())
		} else {
			// Insert into SQlite
			insert_into_sqlite("OKCoin", record.Date, record.Ticker.Sell, record.Ticker.Buy, record.Ticker.Vol, format_currency_string(tickerSplit[i]))
		}
	}
}

// Grabs a snapshot of the current Poloniex exchange
func poloniex_ticker() {

	// Init Poloniex client
	polClient := poloniex.New(config.Poloniex.APIKey, config.Poloniex.APISecret)

	// Get ticker data
	tickers, err := polClient.GetTickers()

	// Check if we had an error, if we did, log it
	if err != nil {
		log.Error(err.Error())
	} else {
		// Create a timestamp now
		ts := strconv.FormatInt(int64(time.Now().Unix()), 10)
		for key, ticker := range tickers {
			// Insert into SQlite
			insert_into_sqlite("Poloniex", ts, strconv.FormatFloat(ticker.LowestAsk, 'f', 8, 64), strconv.FormatFloat(ticker.HighestBid, 'f', 8, 64), strconv.FormatFloat(ticker.BaseVolume, 'f', 8, 64), key)
		}
	}
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

// formats the currency code into something more standard
func format_currency_string(currencyCode string) string {
	return strings.ToUpper(strings.Replace(strings.Replace(strings.Replace(currencyCode, "btc", "", -1), "_", "", -1), "BTC", "", -1))
}

// Check for and print/panic on errors
func check(e error) {
	if e != nil {
		log.Error(e.Error())
	}
}

// Open SQlite Connection
func sqlite_open() *sql.DB {
	db, err := sql.Open("sqlite3", home+"/data.db")
	if err != nil {
		log.Error(err.Error())
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

		defer sqliteDB.Close() // Don't forget to close

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

	// If the exchange name is not there, ignore, otherwise run
	if len(exchange) > 0 && len(currencyCode) > 0 {

		// Clean strings, if the string doesn't contain anything, default
		clean_strings(&timestamp, &ask, &bid, &volume)

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
}

// SELECT function ifromnto sqlite
func query_exchange_sqlite(exchange string, currencyCode string) (resp *APIStruct, err error) {

	// If the exchange name is not there, ignore, otherwise run
	if len(exchange) > 0 && len(currencyCode) > 0 {

		// Write to DB
		sqliteDB := sqlite_open()

		// Query for data
		response := sqliteDB.QueryRow(`select exchange, ask, bid, ROUND((ask + bid) / 2, 8) as price,
				volume as volume, datetime(timestamp, 'unixepoch') as timestamp, currencyCode
				from exchanges
				where currencyCode = ? and exchange = ? order by ID desc LIMIT 1;`, currencyCode, exchange)

		tmp := &APIStruct{}
		// Scan data into response
		err := response.Scan(&tmp.Exchange, &tmp.Ask, &tmp.Bid, &tmp.Average, &tmp.Volume, &tmp.DateUpdated, &tmp.CurrencyCode)
		if err != nil {
			log.Warning("%q\n", err)
			return nil, errors.New("No values found")
		}
		// Close the sqlite connection
		sqliteDB.Close()

		// return response
		return tmp, nil
	}
	log.Warning("Nothing was queried!")
	return nil, errors.New("Exchange or currency code empty")
}

// SELECT function ifromnto sqlite
func query_exchange_currency_codes_sqlite(exchange string, url string) (resp []*string, err error) {

	// If the exchange name is not there, ignore, otherwise run
	if len(exchange) > 0 {

		// Write to DB
		sqliteDB := sqlite_open()

		// Query for data
		response, err := sqliteDB.Query(`select DISTINCT currencyCode from exchanges where exchange = ?;`, exchange)
		// Check if there are errors
		if err != nil {
			log.Error(err.Error())
			return nil, err
		} else {
			// Scan the values into a string slice
			for response.Next() {
				var tmp string
				response.Scan(&tmp)
				tmp = url + tmp
				resp = append(resp, &tmp)
			}

			// If anything was returned
			if len(resp) == 0 {
				return nil, errors.New("Exchange doesn't exist")
			}

		}

		// return response
		return resp, nil
	}
	log.Warning("Nothing was queried!")
	return nil, errors.New("Exchange doesn't exists")
}

// SELECT function sqlite
func query_list_of_exchanges(url string) (resp []*string, err error) {

	// Write to DB
	sqliteDB := sqlite_open()

	// Query for data
	response, err := sqliteDB.Query(`select DISTINCT exchange from exchanges;`)
	// Check if there are errors
	if err != nil {
		log.Error(err.Error())
		return nil, err
	} else {
		// Scan the values into a string slice
		for response.Next() {
			var tmp string
			response.Scan(&tmp)
			tmp = url + tmp
			resp = append(resp, &tmp)
		}

		// If anything was returned
		if len(resp) == 0 {
			return nil, errors.New("No exchanges exist")
		}

	}

	// return response
	return resp, nil
}

// clean strings before inserting, to provide default values
func clean_strings(timestamp *string, ask *string, bid *string, volume *string) {
	if len(*timestamp) == 0 {
		*timestamp = strconv.FormatInt(int64(time.Now().Unix()), 10)
	}
	if len(*ask) == 0 {
		*ask = "0"
	}
	if len(*bid) == 0 {
		*bid = "0"
	}
	if len(*volume) == 0 {
		*volume = "0"
	}
}

// Waits for the minute to tick over
func minuteTicker() *time.Ticker {
	// Return new ticker that triggers on the minute
	return time.NewTicker(time.Second * time.Duration(int(60*10)-time.Now().Second()))
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
		port := viper.GetString("config.port")
		krakenurl := viper.GetString("exchanges.kraken.url")
		krakenAPIKey := viper.GetString("exchanges.kraken.apiKey")
		krakenAPISecret := viper.GetString("exchanges.kraken.apiSecret")
		lunourl := viper.GetString("exchanges.luno.url")
		bitstampurl := viper.GetString("exchanges.bitstamp.url")
		bitfinexurl := viper.GetString("exchanges.bitfinex.url")
		bitfinextickers := viper.GetString("exchanges.bitfinex.tickers")
		bitsquareurl := viper.GetString("exchanges.bitsquare.url")
		bitsquaretickers := viper.GetString("exchanges.bitsquare.tickers")
		btccurl := viper.GetString("exchanges.btcc.url")
		btcctickers := viper.GetString("exchanges.btcc.tickers")
		okcoinurl := viper.GetString("exchanges.okcoin.url")
		okcointickers := viper.GetString("exchanges.okcoin.tickers")
		poloniexAPIKey := viper.GetString("exchanges.poloniex.apiKey")
		poloniexAPISecret := viper.GetString("exchanges.poloniex.apiSecret")

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

		// Bitfinex
		bitfinex := BitfinexConfig{
			URL:     bitfinexurl,
			Tickers: bitfinextickers,
		}

		// Bitsquare
		bitsquare := BitsquareConfig{
			URL:     bitsquareurl,
			Tickers: bitsquaretickers,
		}

		// BTCChina
		btcc := BtccConfig{
			URL:     btccurl,
			Tickers: btcctickers,
		}

		// OKCoin
		okcoin := OKCoinConfig{
			URL:     okcoinurl,
			Tickers: okcointickers,
		}

		// Poloniex
		poloniex := PoloniexConfig{
			APIKey:    poloniexAPIKey,
			APISecret: poloniexAPISecret,
		}

		// Main Config
		config = Config{
			LogFile:        logFile,
			SqliteLocation: sqliteLocation,
			Port:           port,
			Kraken:         kraken,
			Luno:           luno,
			Bitstamp:       bitstamp,
			Bitfinex:       bitfinex,
			Bitsquare:      bitsquare,
			BTCC:           btcc,
			OKCoin:         okcoin,
			Poloniex:       poloniex,
		}
	}

	// Monitor the config file for changes and reload
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {

		// Re-configure config
		config_init()

		// Print out what the new config is
		log.Info("Log file %+v", config)

		// Re-configure logging
		config_log()

		log.Info("Config file changed")
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
	// go bitcoin_prices()

	// Notify log that we are up and running
	log.Info("started kyco.bitcoin.currency.tickers")

	// Setup API
	router := mux.NewRouter()

	// Setup Route
	router.HandleFunc("/{exchange}/{currencyCode}", get_exchange_rate).Methods("GET")
	router.HandleFunc("/{exchange}", show_exchange_methods).Methods("GET")
	router.HandleFunc("/", show_exchanges).Methods("GET")

	// Create listen and serve
	http.ListenAndServe(":"+config.Port, router)
}
