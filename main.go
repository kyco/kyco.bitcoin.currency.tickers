package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Beldur/kraken-go-api-client"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/mux"
	"github.com/jyap808/go-poloniex"
	_ "github.com/mattn/go-sqlite3"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
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

	data, err := queryExchangeSQLite(params["exchange"], params["currencyCode"])
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
	url = "http://" + req.Header.Get("X-Forwarded-Server") + "/" + params["exchange"] + "/"

	// Grab exchange data
	data, err := queryExchangeCurrencyCodesSQLite(params["exchange"], url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Infof("Called: %s\n", params["exchange"])

	// Return exchange data
	json.NewEncoder(w).Encode(data)
}

// Get list of exchanges
func showExchanges(w http.ResponseWriter, req *http.Request) {

	var (
		url = ""
	)

	// Get the URL
	url = "http://" + req.Header.Get("X-Forwarded-Server") + "/"

	// Grab exchange data
	data, err := queryListOfExchanges(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Infof("Called Root Page")

	// Return exchange data
	json.NewEncoder(w).Encode(data)
}

// Initialises various bitcoin price tickers
func bitcoinPrices() {

	// Tick on the minute
	// t := minuteTicker()

	for {

		// wait for the tick
		// <-t.C

		// Start luno ticker
		lunoTicker()
		log.Notice("Ran Luno Ticker")

		// Start bitstamp ticker
		bitstampTicker()
		log.Notice("Ran Bitstamp Ticker")

		// Start kraken ticker
		krakenTicker()
		log.Notice("Ran Kraken Ticker")

		// Start bitfinex ticker
		bitfinexTicker()
		log.Notice("Ran Bitfinex Ticker")

		// Start bitsquare ticker
		bitsquareTicker()
		log.Notice("Ran Bitsquare Ticker")

		// Start btcc ticker
		btccTicker()
		log.Notice("Ran BTCChina Ticker")

		// Start okcoin ticker
		okcoinTicker()
		log.Notice("Ran OKCoin Ticker")

		// Start poloniex ticker
		poloniexTicker()
		log.Notice("Ran Poloniex Ticker")

		time.Sleep(10 * time.Minute)

	}

}

// Grabs a snapshot of the current luno exchange
func lunoTicker() {
	// Make API call to luno
	resp := apiCall(config.Luno.URL)
	// If an empty response was returned
	if resp == nil {
		return
	}

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
			insertIntoSQLite("Luno", timestampString, record.Tickers[i].Ask, record.Tickers[i].Bid, "1", record.Tickers[i].Pair[3:])
		}
	}
}

// Grabs a snapshot of the current bitstamp exchange
func bitstampTicker() {
	// Make API call to bitstamp
	resp := apiCall(config.Bitstamp.URL)

	// If an empty response was returned
	if resp == nil {
		return
	}

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
		insertIntoSQLite("Bitstamp", record.Timestamp, record.Ask, record.Bid, record.Volume, "USD")
	}
}

// Gets ticker data from kraken
func krakenTicker() {
	// If the API keys are not present, just return
	if len(config.Kraken.APIKey) == 0 && len(config.Kraken.APISecret) == 0 {
		return
	}

	api := krakenapi.New(config.Kraken.APIKey, config.Kraken.APISecret)

	// There are also some strongly typed methods available
	ticker, err := api.Ticker(krakenapi.XXBTZEUR, krakenapi.XXBTZUSD, krakenapi.XXBTZGBP, krakenapi.DASHXBT, krakenapi.XETCXXBT, krakenapi.XLTCXXBT)
	if err != nil {
		log.Error(err.Error())
	} else {

		v := reflect.ValueOf(ticker).Elem()
		typeOfT := v.Type()
		for j := 0; j < v.NumField(); j++ {

			f := v.Field(j)
			inter := f.Interface().(krakenapi.PairTickerInfo)

			// Check if the ask value is empty
			if len(inter.Ask) > 0 {

				// Insert into SQlite
				insertIntoSQLite("Kraken", strconv.FormatInt(int64(time.Now().Unix()), 10), inter.Ask[0], inter.Bid[0], inter.Volume[0], formatCurrencyString(typeOfT.Field(j).Name, "Kraken"))
			}
		}
	}
}

// Grabs a snapshot of the current bitfinex exchange
func bitfinexTicker() {
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
		resp := apiCall(config.Bitfinex.URL + tickerSplit[i])

		// If an empty response was returned
		if resp == nil {
			continue
		}

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
			insertIntoSQLite("Bitfinex", record.Timestamp, record.Ask, record.Bid, record.Volume, formatCurrencyString(tickerSplit[i], "Bitfinex"))
		}
	}
}

// Grabs a snapshot of the current bitsquare exchange
func bitsquareTicker() {
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
		resp := apiCall(config.Bitsquare.URL + tickerSplit[i])

		// If an empty response was returned
		if resp == nil {
			continue
		}

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
			insertIntoSQLite("Bitsquare", ts, record[0].Sell, record[0].Buy, record[0].VolumeRight, formatCurrencyString(tickerSplit[i], "Bitsquare"))
		}
	}
}

// Grabs a snapshot of the current BTCC exchange
func btccTicker() {
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
		resp := apiCall(config.BTCC.URL + tickerSplit[i])

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
			insertIntoSQLite("BTCChina", strconv.FormatInt((record.Ticker.Timestamp/1000), 10), strconv.FormatFloat(record.Ticker.AskPrice, 'f', 2, 64), strconv.FormatFloat(record.Ticker.BidPrice, 'f', 2, 64), strconv.FormatFloat(record.Ticker.Volume, 'f', 2, 64), formatCurrencyString(tickerSplit[i], "btcc"))
		}
	}
}

// Grabs a snapshot of the current OKCoin exchange
func okcoinTicker() {
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
		resp := apiCall(config.OKCoin.URL + tickerSplit[i])

		// If an empty response was returned
		if resp == nil {
			continue
		}

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
			insertIntoSQLite("OKCoin", record.Date, record.Ticker.Sell, record.Ticker.Buy, record.Ticker.Vol, formatCurrencyString(tickerSplit[i], "okcoin"))
		}
	}
}

// Grabs a snapshot of the current Poloniex exchange
func poloniexTicker() {

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
			insertIntoSQLite("Poloniex", ts, strconv.FormatFloat(ticker.LowestAsk, 'f', 8, 64), strconv.FormatFloat(ticker.HighestBid, 'f', 8, 64), strconv.FormatFloat(ticker.BaseVolume, 'f', 8, 64), key)
		}
	}
}

// performs an API call to a URL and returns a JSON body response
func apiCall(urlRequest string) *http.Response {

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
func formatCurrencyString(currencyCode string, exchange string) string {
	// Replace BTC
	replaceBTC := strings.Replace(strings.ToUpper(currencyCode), "BTC", "", -1)

	// Perform extra replacements with Kraken
	if exchange == "Kraken" {
		replaceBTC = strings.Replace(replaceBTC, "XBTC", "", -1)
		replaceBTC = strings.Replace(replaceBTC, "XXBTZ", "", -1)
		replaceBTC = strings.Replace(replaceBTC, "XXBT", "", -1)
		replaceBTC = strings.Replace(replaceBTC, "X", "", -1)
		replaceBTC = strings.Replace(replaceBTC, "DASHBT", "DASH", -1)
	}
	replaceBTC = strings.Replace(replaceBTC, "_", "", -1)
	return replaceBTC
}

// Check for and print/panic on errors
func check(e error) {
	if e != nil {
		log.Error(e.Error())
	}
}

// Open SQlite Connection
func sqliteOpen() *sql.DB {
	db, err := sql.Open("sqlite3", home+"/data.db")
	if err != nil {
		log.Error(err.Error())
	}
	return db
}

// Sets up the sqlite databases and connections
func setupSQLiteDB() {
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
		sqliteDB := sqliteOpen()

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
func insertIntoSQLite(exchange string, timestamp string, ask string, bid string, volume string, currencyCode string) {

	// If the exchange name is not there, ignore, otherwise run
	if len(exchange) > 0 && len(currencyCode) > 0 {

		// Clean strings, if the string doesn't contain anything, default
		cleanStrings(&timestamp, &ask, &bid, &volume)

		// Write to DB
		sqliteDB := sqliteOpen()
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
func queryExchangeSQLite(exchange string, currencyCode string) (resp *APIStruct, err error) {

	// If the exchange name is not there, ignore, otherwise run
	if len(exchange) > 0 && len(currencyCode) > 0 {

		// Write to DB
		sqliteDB := sqliteOpen()

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
func queryExchangeCurrencyCodesSQLite(exchange string, url string) (resp []*string, err error) {

	// If the exchange name is not there, ignore, otherwise run
	if len(exchange) > 0 {

		// Write to DB
		sqliteDB := sqliteOpen()

		// Query for data
		response, err := sqliteDB.Query(`select DISTINCT currencyCode from exchanges where exchange = ?;`, exchange)
		// Check if there are errors
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}

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

		// return response
		return resp, nil
	}
	log.Warning("Nothing was queried!")
	return nil, errors.New("Exchange doesn't exists")
}

// SELECT function sqlite
func queryListOfExchanges(url string) (resp []*string, err error) {

	// Write to DB
	sqliteDB := sqliteOpen()

	// Query for data
	response, err := sqliteDB.Query(`select DISTINCT exchange from exchanges;`)
	// Check if there are errors
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

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

	// return response
	return resp, nil
}

// clean strings before inserting, to provide default values
func cleanStrings(timestamp *string, ask *string, bid *string, volume *string) {
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
func configLog() {

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
func configInit() {

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
		configInit()

		// Print out what the new config is
		log.Info("Log file %+v", config)

		// Re-configure logging
		configLog()

		log.Info("Config file changed")
	})
}

func main() {

	// Initialise config file and settings
	configInit()

	// Configure logging
	configLog()

	// don't forget to close the log file
	defer logFile.Close()

	// Setup Sqlite DB
	setupSQLiteDB()

	// Start bitcoin ticker
	go bitcoinPrices()

	// Notify log that we are up and running
	log.Info("started kyco.bitcoin.currency.tickers")

	// Setup API
	router := mux.NewRouter()

	// Setup Route
	router.HandleFunc("/{exchange}/{currencyCode}", get_exchange_rate).Methods("GET")
	router.HandleFunc("/{exchange}", show_exchange_methods).Methods("GET")
	router.HandleFunc("/", showExchanges).Methods("GET")

	// Create listen and serve
	http.ListenAndServe(":"+config.Port, router)
}
