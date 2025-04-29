package bitget

import (
	"github.com/skip-mev/connect/v2/oracle/config"
	"time"
)

const (
	Name = "bitget_api"

	URL = "https://api.bitget.com/api/v2/spot/market/tickers?symbol=%s"
)

var DefaultAPIConfig = config.APIConfig{
	Name:             Name,
	Atomic:           false,
	Enabled:          true,
	Timeout:          3000 * time.Millisecond,
	Interval:         100 * time.Millisecond,
	ReconnectTimeout: 2000 * time.Millisecond,
	MaxQueries:       1,
	Endpoints:        []config.Endpoint{{URL: URL}},
}

type (
	// BitgetResponse {
	//	"code": "00000",
	//	"msg": "success",
	//	"requestTime": 1745825016409,
	//	"data": [
	//		{
	//			"open": "0.83729",
	//			"symbol": "INITUSDT",
	//			"high24h": "0.84892",
	//			"low24h": "0.6568",
	//			"lastPr": "0.67876",
	//			"quoteVolume": "24248569.29",
	//			"baseVolume": "33375691.87",
	//			"usdtVolume": "24248569.2829031",
	//			"ts": "1745825015007",
	//			"bidPr": "0.67854",
	//			"askPr": "0.67896",
	//			"bidSz": "143.16",
	//			"askSz": "1026.02",
	//			"openUtc": "0.70213",
	//			"changeUtc24h": "-0.03321",
	//			"change24h": "-0.18934"
	//		}
	//	]
	//}
	BitgetResponse struct {
		Code        string       `json:"code"`
		Msg         string       `json:"msg"`
		RequestTime int64        `json:"requestTime"`
		Data        []BitgetData `json:"data"`
	}

	BitgetData struct {
		Open         string `json:"open"`
		Symbol       string `json:"symbol"`
		High24H      string `json:"high24h"`
		Low24H       string `json:"low24h"`
		LastPr       string `json:"lastPr"`
		QuoteVolume  string `json:"quoteVolume"`
		BaseVolume   string `json:"baseVolume"`
		UsdtVolume   string `json:"usdtVolume"`
		Ts           string `json:"ts"`
		BidPr        string `json:"bidPr"`
		AskPr        string `json:"askPr"`
		BidSz        string `json:"bidSz"`
		AskSz        string `json:"askSz"`
		OpenUtc      string `json:"openUtc"`
		ChangeUtc24H string `json:"changeUtc24h"`
		Change24H    string `json:"change24h"`
	}
)
