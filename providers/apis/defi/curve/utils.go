package curve

import (
	"time"

	"github.com/skip-mev/connect/v2/oracle/config"
)

// API specification:
//
// - https://docs.curve.fi/curve-api/curve-prices/
// - https://docs.curve.fi/curve-api/curve-api/
const (
	Name = "curve_finance_api"

	// https://prices.curve.fi/v1/usd_price/ethereum/0xCd5fE23C85820F7B72D0926FC9b05b43E359b7ee
	URL = "https://prices.curve.fi/v1/usd_price/%s/%s"
)

var DefaultAPIConfig = config.APIConfig{
	Name:             Name,
	Atomic:           false,
	Enabled:          true,
	Timeout:          500 * time.Millisecond,
	Interval:         20 * time.Second,
	ReconnectTimeout: 2000 * time.Millisecond,
	MaxQueries:       1,
	Endpoints:        []config.Endpoint{{URL: URL}},
}

type (
	//{
	//	"data": {
	//		"address": "0xCd5fE23C85820F7B72D0926FC9b05b43E359b7ee",
	//		"usd_price": 1674.1742629502855,
	//		"last_updated": "2025-04-16T06:04:23"
	//	}
	//}

	CurveResponse struct {
		Data CurveData `json:"data"`
	}

	CurveData struct {
		Address     string  `json:"address"`
		UsdPrice    float64 `json:"usd_price"`
		LastUpdated string  `json:"last_updated"`
	}
)
