package geckoterminal

import (
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/skip-mev/connect/v2/oracle/config"
	"github.com/skip-mev/connect/v2/oracle/types"
	"github.com/stretchr/testify/require"
)

func TestNewAPIHandler(t *testing.T) {
	tests := []struct {
		name        string
		apiConfig   config.APIConfig
		expectError bool
	}{
		{
			name: "valid config",
			apiConfig: config.APIConfig{
				Name:    Name,
				Enabled: true,
				Endpoints: []config.Endpoint{
					{
						URL: "https://api.geckoterminal.com/api/v2/networks/ethereum/tokens/%s",
					},
				},
				MaxQueries:       1,
				Atomic:           false,
				Timeout:          500 * time.Millisecond,
				Interval:         20 * time.Second,
				ReconnectTimeout: 2000 * time.Millisecond,
			},
			expectError: false,
		},
		{
			name: "wrong name",
			apiConfig: config.APIConfig{
				Name:    "wrong_name",
				Enabled: true,
			},
			expectError: true,
		},
		{
			name: "disabled",
			apiConfig: config.APIConfig{
				Name:    Name,
				Enabled: false,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := NewAPIHandler(tt.apiConfig)
			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, handler)
			} else {
				require.NoError(t, err)
				require.NotNil(t, handler)
				require.Equal(t, tt.apiConfig, handler.(*APIHandler).api)
			}
		})
	}
}

func TestGetBaseAddress(t *testing.T) {
	tests := []struct {
		name         string
		metadataJSON string
		expectedAddr string
		expectError  bool
	}{
		{
			name: "valid metadata",
			metadataJSON: `{
				"address": "0X87428A53E14D24AB19C6CA4939B4DF93B8996CA9/UNISWAP_V3,0X8236A87084F8B84306F72007F36F2618A5634494/UNISWAP_V3,0X2260FAC5E5542A773AA44FBCFEDF7C193BC2C599",
				"base_decimals": 8,
				"quote_decimals": 8,
				"invert": true
			}`,
			expectedAddr: "0X8236A87084F8B84306F72007F36F2618A5634494",
			expectError:  false,
		},
		{
			name: "invalid json",
			metadataJSON: `{
				invalid json
			}`,
			expectedAddr: "",
			expectError:  true,
		},
		{
			name: "insufficient addresses",
			metadataJSON: `{
				"address": "0X87428A53E14D24AB19C6CA4939B4DF93B8996CA9",
				"base_decimals": 8,
				"quote_decimals": 8,
				"invert": true
			}`,
			expectedAddr: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &APIHandler{}
			addr, err := handler.getBaseAddress(tt.metadataJSON)
			if tt.expectError {
				require.Error(t, err)
				require.Empty(t, addr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedAddr, addr)
			}
		})
	}
}

func TestCreateURL(t *testing.T) {
	tests := []struct {
		name        string
		apiConfig   config.APIConfig
		tickers     []types.ProviderTicker
		expectedURL string
		expectError bool
	}{
		{
			name: "valid ticker",
			apiConfig: config.APIConfig{
				Name:    Name,
				Enabled: true,
				Endpoints: []config.Endpoint{
					{
						URL: "https://api.geckoterminal.com/api/v2/networks/ethereum/tokens/%s",
					},
				},
			},
			tickers: []types.ProviderTicker{
				types.NewProviderTicker(
					"LBTC-USD",
					`{"address":"0X87428A53E14D24AB19C6CA4939B4DF93B8996CA9/UNISWAP_V3,0X8236A87084F8B84306F72007F36F2618A5634494/UNISWAP_V3,0X2260FAC5E5542A773AA44FBCFEDF7C193BC2C599","base_decimals":8,"quote_decimals":8,"invert":true}`,
				),
			},
			expectedURL: "https://api.geckoterminal.com/api/v2/networks/ethereum/tokens/0X8236A87084F8B84306F72007F36F2618A5634494",
			expectError: false,
		},
		{
			name: "no tickers",
			apiConfig: config.APIConfig{
				Name:    Name,
				Enabled: true,
				Endpoints: []config.Endpoint{
					{
						URL: "https://api.geckoterminal.com/api/v2/networks/ethereum/tokens/%s",
					},
				},
			},
			tickers:     []types.ProviderTicker{},
			expectedURL: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := NewAPIHandler(tt.apiConfig)
			require.NoError(t, err)

			url, err := handler.CreateURL(tt.tickers)
			if tt.expectError {
				require.Error(t, err)
				require.Empty(t, url)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedURL, url)
			}
		})
	}
}

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name          string
		tickers       []types.ProviderTicker
		responseBody  string
		expectedPrice *big.Float
		expectError   bool
	}{
		{
			name: "valid response",
			tickers: []types.ProviderTicker{
				types.NewProviderTicker(
					"LBTC-USD",
					`{"address":"0X87428A53E14D24AB19C6CA4939B4DF93B8996CA9/UNISWAP_V3,0X8236A87084F8B84306F72007F36F2618A5634494/UNISWAP_V3,0X2260FAC5E5542A773AA44FBCFEDF7C193BC2C599","base_decimals":8,"quote_decimals":8,"invert":true}`,
				),
			},
			responseBody: `{
				"data": {
					"attributes": {
						"price_usd": "50000.00"
					}
				}
			}`,
			expectedPrice: big.NewFloat(50000.00),
			expectError:   false,
		},
		{
			name: "invalid json response",
			tickers: []types.ProviderTicker{
				types.NewProviderTicker(
					"LBTC-USD",
					`{"address":"0X87428A53E14D24AB19C6CA4939B4DF93B8996CA9/UNISWAP_V3,0X8236A87084F8B84306F72007F36F2618A5634494/UNISWAP_V3,0X2260FAC5E5542A773AA44FBCFEDF7C193BC2C599","base_decimals":8,"quote_decimals":8,"invert":true}`,
				),
			},
			responseBody: `{
				invalid json
			}`,
			expectedPrice: nil,
			expectError:   true,
		},
		{
			name: "missing price in response",
			tickers: []types.ProviderTicker{
				types.NewProviderTicker(
					"LBTC-USD",
					`{"address":"0X87428A53E14D24AB19C6CA4939B4DF93B8996CA9/UNISWAP_V3,0X8236A87084F8B84306F72007F36F2618A5634494/UNISWAP_V3,0X2260FAC5E5542A773AA44FBCFEDF7C193BC2C599","base_decimals":8,"quote_decimals":8,"invert":true}`,
				),
			},
			responseBody: `{
				"data": {
					"attributes": {}
				}
			}`,
			expectedPrice: nil,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &APIHandler{
				api: config.APIConfig{
					Name:    Name,
					Enabled: true,
				},
			}

			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create a response object
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
			}

			// Parse the response
			result := handler.ParseResponse(tt.tickers, resp)

			if tt.expectError {
				require.NotEmpty(t, result.UnResolved)
				require.Empty(t, result.Resolved)
			} else {
				require.Empty(t, result.UnResolved)
				require.NotEmpty(t, result.Resolved)

				price, ok := result.Resolved[tt.tickers[0]]
				require.True(t, ok)
				require.NotNil(t, price)
				require.Equal(t, tt.expectedPrice, price.Value)
				require.True(t, time.Since(price.Timestamp) < time.Second)
			}
		})
	}
}
