### Currency Ticker
Written in Golang. This is a project which combines the data from various Bitcoin exchanges around the world.

### How To Run?
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