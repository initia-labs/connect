package geckoterminal

import (
	"time"

	"github.com/skip-mev/connect/v2/oracle/config"
)

// NOTE: All documentation for this file can be located on the GeckoTerminal
// API specification:
//
// - https://api.geckoterminal.com/api/v2
// - https://www.geckoterminal.com/dex-api.

const (
	// Name is the name of the GeckoTerminal provider.
	Name = "gecko_terminal_api"

	// URL is the root URL for the GeckoTerminal API.
	ETH_URL = "https://api.geckoterminal.com/api/v2/networks/eth/tokens/%s"
)

// DefaultETHAPIConfig is the default configuration for querying Ethereum mainnet tokens
// on the GeckoTerminal API.
var DefaultETHAPIConfig = config.APIConfig{
	Name:             Name,
	Atomic:           false,
	Enabled:          true,
	Timeout:          500 * time.Millisecond,
	Interval:         20 * time.Second,
	ReconnectTimeout: 2000 * time.Millisecond,
	MaxQueries:       1,
	Endpoints:        []config.Endpoint{{URL: ETH_URL}},
}

type (
	// GeckoTerminalResponse is the expected response returned by the GeckoTerminal API.
	// The response is json formatted.
	// Example response:
	// {
	//   "data": {
	//     "id": "eth_0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
	//     "type": "token",
	//     "attributes": {
	//       "address": "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
	//       "name": "Wrapped Ether",
	//       "symbol": "WETH",
	//       "decimals": 18,
	//       "image_url": "https://coin-images.coingecko.com/coins/images/2518/large/weth.png?1696503332",
	//       "coingecko_coin_id": "weth",
	//       "total_supply": "2853226432822898981820569.0",
	//       "price_usd": "1793.24",
	//       "fdv_usd": "5114070988.97486",
	//       "total_reserve_in_usd": "1275089731.73157810664933938896",
	//       "volume_usd": {
	//         "h24": "608777978.450451"
	//       },
	//       "market_cap_usd": "5113847843.5746"
	//     },
	//     "relationships": {
	//       "top_pools": {
	//         "data": [
	//           {
	//             "id": "eth_0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640",
	//             "type": "pool"
	//           }
	//         ]
	//       }
	//     }
	//   }
	// }
	GeckoTerminalResponse struct { //nolint
		Data GeckoTerminalData `json:"data"`
	}

	// GeckoTerminalData is the data field in the GeckoTerminalResponse.
	GeckoTerminalData struct { //nolint
		ID            string                  `json:"id"`
		Type          string                  `json:"type"`
		Attributes    GeckoTerminalAttributes `json:"attributes"`
		Relationships struct {
			TopPools struct {
				Data []struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
			} `json:"top_pools"`
		} `json:"relationships"`
	}

	// GeckoTerminalAttributes is the attributes field in the GeckoTerminalData.
	GeckoTerminalAttributes struct { //nolint
		Address           string `json:"address"`
		Name              string `json:"name"`
		Symbol            string `json:"symbol"`
		Decimals          int    `json:"decimals"`
		ImageURL          string `json:"image_url"`
		CoingeckoCoinID   string `json:"coingecko_coin_id"`
		TotalSupply       string `json:"total_supply"`
		PriceUSD          string `json:"price_usd"`
		FDVUSD            string `json:"fdv_usd"`
		TotalReserveInUSD string `json:"total_reserve_in_usd"`
		VolumeUSD         struct {
			H24 string `json:"h24"`
		} `json:"volume_usd"`
		MarketCapUSD string `json:"market_cap_usd"`
	}

	// MetadataJSON represents the structure of the metadata JSON string
	MetadataJSON struct {
		Address       string `json:"address"`
		BaseDecimals  int    `json:"base_decimals"`
		QuoteDecimals int    `json:"quote_decimals"`
		Invert        bool   `json:"invert"`
	}
)
