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

```
go build bitcoin-stats-cmd
./bitcoin-stats-cmd
```

### How it works.
It queries the APIs of various exchanges (more will be added as time goes by) and pops them into a sqlite database.

### What Works
 - Querying Bitstamp (USD only) every 10 minutes and save the response into a sqlite database
 - Querying Luno (NGN, ZAR, MYR, IDR) every 10 minutes and save the response into a sqlite database

### Future / TODO
 - Systemd service files
 - More Database options (MySQL / Postgres)
 - More Exchanges (Kraken, BTCC, etc)
 - Functions which are more dynamic
 - Better logging