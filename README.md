### Currency Ticker
Written in Golang. This is a project which combines the data from various Bitcoin exchanges around the world.

### How To Run?

#### Dependencies
To install the required dependencies to compile your own binaries.
```
go get -u github.com/Beldur/kraken-go-api-client
go get -u github.com/fsnotify/fsnotify
go get -u github.com/mattn/go-sqlite3
go get -u github.com/spf13/viper
go get -u github.com/op/go-logging
```

or

```
glide install
```

Build

```
go build
./kyco.bitcoin.currency.tickers
```

### How it works.
It queries the APIs of various exchanges (more will be added as time goes by) and pops them into a sqlite database.

### What Works
At this point in time each exchange is queried every 10 minutes and the results are saved into a sqlite database

 - Querying Bitstamp (USD only)
 - Querying Luno (NGN, ZAR, MYR, IDR)
 - Querying Kraken (EUR, USD, GBP)
 - Querying Bitfinex (Defined by Config File)
 - Querying Bitsquare (Defined by Config File)
 - Querying BTCChina (Defined by Config File)
 - Querying OKCoin (Defined by Config File)

### Service File
A service file for linux exists in the folder ```init```. Copy this to ```/usr/lib/systemd/user/```. Then run:

```
systemctl enable /usr/lib/systemd/user/bitcoin-stats.service
systemctl daemon-reload
service bitcoin-stats start | status | stop
```

Unless you modify the location of the binary in the service file, you must copy the compiled binary to ```/usr/bin/```.

### Future / TODO
 - More Database options (MySQL / Postgres)
 - More Exchanges
 - Functions which are more dynamic
 - Better logging
 - Automated installation
 