# Configuration for custom endpoints
This file describes how to configure and use custom endpoints in your Connect connection setup.

```json oracle.json
{
  "providers": {
    "binance_ws": {
      "webSocket": {
        "endpoints": [
          {
            "url": "wss://stream.binance.com/stream"
          }
        ]
      }
    },
    "bitfinex_ws": {
      "webSocket": {
        "endpoints": [
          {
            "url": "wss://api-pub.bitfinex.com/ws/2"
          }
        ]
      }
    },
    "bitget_ws": {
      "webSocket": {
        "endpoints": [
          {
            "url": "wss://ws.bitget.com/v2/ws/public"
          }
        ]
      }
    },
    "bitstamp_ws": {
      "webSocket": {
        "endpoints": [
          {
            "url": "wss://ws.bitstamp.net"
          }
        ]
      }
    },
    "bybit_ws": {
      "webSocket": {
        "endpoints": [
          {
            "url": "wss://stream.bybit.com/v5/public/spot"
          }
        ]
      }
    },
    "coinbase_ws": {
      "webSocket": {
        "endpoints": [
          {
            "url": "wss://ws-feed.exchange.coinbase.com"
          }
        ]
      }
    },
    "crypto_dot_com_ws": {
      "webSocket": {
        "endpoints": [
          {
            "url": "wss://stream.crypto.com/exchange/v1/market"
          }
        ]
      }
    },
    "gate_ws": {
      "webSocket": {
        "endpoints": [
          {
            "url": "wss://api.gateio.ws/ws/v4/"
          }
        ]
      }
    },
    "huobi_ws": {
      "webSocket": {
        "endpoints": [
          {
            "url": "wss://api.huobi.pro/ws"
          }
        ]
      }
    },
    "kraken_ws": {
      "webSocket": {
        "endpoints": [
          {
            "url": "wss://ws.kraken.com"
          }
        ]
      }
    },
    "mexc_ws": {
      "webSocket": {
        "endpoints": [
          {
            "url": "wss://wbs.mexc.com/ws"
          }
        ]
      }
    },
    "okx_ws": {
      "webSocket": {
        "endpoints": [
          {
            "url": "wss://ws.okx.com:8443/ws/v5/public"
          }
        ]
      }
    },
    "coingecko_api": {
      "api": {
        "endpoints": [
          {
            "url": "https://pro-api.coingecko.com/api/v3",
            "authentication": {
              "apiKeyHeader": "x-cg-pro-api-key",
              "apiKey": "api-key"
            }
          }
        ]
      }
    }
  }
}
```
with the `oracle.json` file path, enther the following command to run connect.
```shell
connect --oracle-config path/to/oracle.json
```