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
	Ask                 string `json:"ask"`
	Timestamp           int64  `json:"timestamp"`
	Bid                 string `json:"bid"`
	Rolling24HourVolume string `json:"rolling_24_hour_volume"`
	LastTrade           string `json:"last_trade"`
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
	resp := api_call("https://api.mybitx.com/api/1/ticker?pair=XBTZAR")

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
	// Format timestamp as string
	timestampString := strconv.FormatInt(record.Timestamp, 10)
	insert_into_sqlite("Luno", timestampString, record.Ask, record.Bid, "0", "ZAR")
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

// // Creates a standard config file
// func create_default_config_file() {
// 	// Create a config file in the user's home folder
// 	createFolderErr := os.MkdirAll(home, 0777)
// 	check(createFolderErr)
// 	// Create a new file and write some data to it
// 	d1 := []byte("[testnet]\nnodeHost=\"127.0.0.1:8333\"\nnodePost=\"8333\"\nnodeUsername=\"nothing\"\nnodePassword=\"password\"\nblocks=true\nmempool=true\n\n[mainnet]\nnodeHost=\"127.0.0.1:8333\"\nnodePost=\"8333\"\nnodeUsername=\"nothing\"\nnodePassword=\"password\"\nblocks=true\nmempool=true")
// 	err := ioutil.WriteFile(home+"config.toml", d1, 0666)
// 	check(err)
// }

// // Initialise configuration file
// func config_init() {

// 	// Config File
// 	viper.SetConfigName("config") // no need to include file extension
// 	viper.AddConfigPath(home)     // set the path of your config file

// 	err := viper.ReadInConfig()
// 	if err != nil {
// 		log.Println("Config file not found...")
// 		// Create the config file
// 		create_default_config_file()
// 	}
// 	dev_nodeHost := viper.GetString("testnet.nodeHost")
// 	dev_nodePort := viper.GetString("testnet.nodePort")
// 	dev_nodeUsername := viper.GetString("testnet.nodeUsername")
// 	dev_nodePassword := viper.GetString("testnet.nodePassword")
// 	dev_blocks := viper.GetBool("testnet.blocks")
// 	dev_mempool := viper.GetBool("testnet.mempool")

// 	prod_nodeHost := viper.GetString("mainnet.nodeHost")
// 	prod_nodePort := viper.GetString("mainnet.nodePort")
// 	prod_nodeUsername := viper.GetString("mainnet.nodeUsername")
// 	prod_nodePassword := viper.GetString("mainnet.nodePassword")
// 	prod_blocks := viper.GetBool("mainnet.blocks")
// 	prod_mempool := viper.GetBool("mainnet.mempool")

// 	Testnet = Config{
// 		nodeHost:     dev_nodeHost,
// 		nodePort:     dev_nodePort,
// 		nodeUsername: dev_nodeUsername,
// 		nodePassword: dev_nodePassword,
// 		Blocks:       dev_blocks,
// 		Mempool:      dev_mempool}
// 	Mainnet = Config{
// 		nodeHost:     prod_nodeHost,
// 		nodePort:     prod_nodePort,
// 		nodeUsername: prod_nodeUsername,
// 		nodePassword: prod_nodePassword,
// 		Blocks:       prod_blocks,
// 		Mempool:      prod_mempool}
// }

// Check for and print/panic on errors
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// // Initialise clear
// func init() {
// 	clear = make(map[string]func()) //Initialize it
// 	clear["linux"] = func() {
// 		cmd := exec.Command("clear") //Linux example, its tested
// 		cmd.Stdout = os.Stdout
// 		cmd.Run()
// 	}
// 	clear["windows"] = func() {
// 		cmd := exec.Command("cls") //Windows example it is untested, but I think its working
// 		cmd.Stdout = os.Stdout
// 		cmd.Run()
// 	}
// }

// // Clear
// func CallClear() {
// 	value, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
// 	if ok {                          //if we defined a clear func for that platform:
// 		value() //we execute it
// 	} else { //unsupported platform
// 		panic("Your platform is unsupported! I can't clear terminal screen :(")
// 	}
// }

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
