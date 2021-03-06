# Currency Ticker
Written in Golang. This is a project which combines the data from various Bitcoin exchanges around the world.

## How To Run?

### Automatically
Just run 

```
chmod +x install.sh; ./install.sh
```

And it will take care of the compiling and installation process for you. You just need to have [go](https://golang.org) installed as well as [glide](https://glide.sh/).

### Manually
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

## Config File
In both installation types, a config file is required. You'll need to create that manually until I've written an automated way to deal with that.

Create the folder ```kyco.bitcoin.currency.tickers``` in ```~/.config/``` and copy the contents from ```init/config.toml``` into that folder.
```
mkdir -p ~/.config/kyco.bitcoin.currency.tickers/
cp init/config.toml ~/.config/kyco.bitcoin.currency.tickers/
```

You'll also need to grab an API key from Kraken if you want to use their exchange.

## How it works.
It queries the APIs of various exchanges (more will be added as time goes by) and pops them into a sqlite database.

## What Works
At this point in time each exchange is queried every 10 minutes and the results are saved into a sqlite database

 - Querying Bitstamp (USD only)
 - Querying Luno (NGN, ZAR, MYR, IDR)
 - Querying Kraken (EUR, USD, GBP)
 - Querying Bitfinex (Defined by Config File)
 - Querying Bitsquare (Defined by Config File)
 - Querying BTCChina (Defined by Config File)
 - Querying OKCoin (Defined by Config File)
 - Querying Poloniex (All Supported Tickers)

## Service File
A service file for linux exists in the folder ```init```. Copy this to ```/usr/lib/systemd/user/```. Change the user in the service file to match the user and group of your choice on your machine. Then run:

```
systemctl enable /usr/lib/systemd/user/kbct.service
systemctl daemon-reload
service bitcoin-stats start | status | stop
```

Unless you modify the location of the binary in the service file, you must copy the compiled binary to ```/usr/bin/```.

## Future / TODO
 - More Database options (MySQL / Postgres)
 - More Exchanges
 - Functions which are more dynamic
 - Better logging
 - Automated installation
 