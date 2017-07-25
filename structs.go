package main

// Config type
type Config struct {
	LogFile        string
	SqliteLocation string
	Kraken         KrakenConfig
	Luno           LunoConfig
	Bitstamp       BitstampConfig
	Bitfinex       BitfinexConfig
	Bitsquare      BitsquareConfig
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

type BitfinexConfig struct {
	URL     string
	Tickers string
}

type BitsquareConfig struct {
	URL     string
	Tickers string
}

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

type Bitfinex struct {
	Ask       string `json:"ask"`
	Bid       string `json:"bid"`
	High      string `json:"high"`
	LastPrice string `json:"last_price"`
	Low       string `json:"low"`
	Mid       string `json:"mid"`
	Timestamp string `json:"timestamp"`
	Volume    string `json:"volume"`
}

type Bitsquare struct {
	Buy         string `json:"buy"`
	High        string `json:"high"`
	Last        string `json:"last"`
	Low         string `json:"low"`
	Sell        string `json:"sell"`
	VolumeLeft  string `json:"volume_left"`
	VolumeRight string `json:"volume_right"`
}
