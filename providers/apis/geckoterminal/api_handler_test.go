package geckoterminal

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/skip-mev/connect/v2/oracle/types"
	"github.com/skip-mev/connect/v2/providers/base/testutils"
	providertypes "github.com/skip-mev/connect/v2/providers/types"
)

func TestGetTokenInfo(t *testing.T) {
	testCases := []struct {
		name            string
		metadataJSON    string
		expectedNetwork string
		expectedAddr    string
		expectError     bool
	}{
		{
			name:            "valid metadata with network",
			metadataJSON:    `{"network": "eth", "address": "0x123"}`,
			expectedNetwork: "eth",
			expectedAddr:    "0x123",
			expectError:     false,
		},
		{
			name:            "valid metadata without network (should default to eth)",
			metadataJSON:    `{"address": "0x123"}`,
			expectedNetwork: "eth",
			expectedAddr:    "0x123",
			expectError:     false,
		},
		{
			name:         "invalid json",
			metadataJSON: `{invalid json}`,
			expectError:  true,
		},
		{
			name:         "missing address",
			metadataJSON: `{"network": "eth"}`,
			expectError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h, err := NewAPIHandler(DefaultETHAPIConfig)
			require.NoError(t, err)

			handler := h.(*APIHandler)
			network, addr, err := handler.GetTokenInfo(tc.metadataJSON)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedNetwork, network)
				require.Equal(t, tc.expectedAddr, addr)
			}
		})
	}
}

func TestCreateURL(t *testing.T) {
	testCases := []struct {
		name        string
		ticker      types.ProviderTicker
		expectedURL string
		expectError bool
	}{
		{
			name: "valid ticker",
			ticker: types.DefaultProviderTicker{
				OffChainTicker: "0x123",
				JSON:           `{"network": "eth", "address": "0x123"}`,
			},
			expectedURL: "https://api.geckoterminal.com/api/v2/simple/networks/eth/token_price/0x123",
			expectError: false,
		},
		{
			name: "invalid metadata json",
			ticker: types.DefaultProviderTicker{
				OffChainTicker: "0x123",
				JSON:           `{invalid json}`,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h, err := NewAPIHandler(DefaultETHAPIConfig)
			require.NoError(t, err)

			url, err := h.CreateURL([]types.ProviderTicker{tc.ticker})
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedURL, url)
			}
		})
	}
}

func TestParseResponse(t *testing.T) {
	testCases := []struct {
		name          string
		ticker        types.ProviderTicker
		response      *http.Response
		expectedPrice float64
		expectError   bool
		errorCode     providertypes.ErrorCode
	}{
		{
			name: "valid response",
			ticker: types.DefaultProviderTicker{
				OffChainTicker: "0x123",
				JSON:           `{"network": "eth", "address": "0x123"}`,
			},
			response: testutils.CreateResponseFromJSON(`{
				"data": {
					"type": "simple_token_price",
					"attributes": {
						"token_prices": {
							"0x123": "1.23"
						}
					}
				}
			}`),
			expectedPrice: 1.23,
			expectError:   false,
		},
		{
			name: "invalid response type",
			ticker: types.DefaultProviderTicker{
				OffChainTicker: "0x123",
				JSON:           `{"network": "eth", "address": "0x123"}`,
			},
			response: testutils.CreateResponseFromJSON(`{
				"data": {
					"type": "wrong_type",
					"attributes": {
						"token_prices": {
							"0x123": "1.23"
						}
					}
				}
			}`),
			expectError: true,
			errorCode:   providertypes.ErrorInvalidResponse,
		},
		{
			name: "invalid price format",
			ticker: types.DefaultProviderTicker{
				OffChainTicker: "0x123",
				JSON:           `{"network": "eth", "address": "0x123"}`,
			},
			response: testutils.CreateResponseFromJSON(`{
				"data": {
					"type": "simple_token_price",
					"attributes": {
						"token_prices": {
							"0x123": "invalid_price"
						}
					}
				}
			}`),
			expectError: true,
			errorCode:   providertypes.ErrorFailedToParsePrice,
		},
		{
			name: "missing price for ticker",
			ticker: types.DefaultProviderTicker{
				OffChainTicker: "0x123",
				JSON:           `{"network": "eth", "address": "0x123"}`,
			},
			response: testutils.CreateResponseFromJSON(`{
				"data": {
					"type": "simple_token_price",
					"attributes": {
						"token_prices": {}
					}
				}
			}`),
			expectError: true,
			errorCode:   providertypes.ErrorNoResponse,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h, err := NewAPIHandler(DefaultETHAPIConfig)
			require.NoError(t, err)

			// First create URL to populate cache
			_, err = h.CreateURL([]types.ProviderTicker{tc.ticker})
			require.NoError(t, err)

			resp := h.ParseResponse([]types.ProviderTicker{tc.ticker}, tc.response)

			if tc.expectError {
				require.Len(t, resp.Resolved, 0)
				require.Len(t, resp.UnResolved, 1)
				require.Equal(t, tc.errorCode, resp.UnResolved[tc.ticker].ErrorWithCode.Code())
			} else {
				require.Len(t, resp.Resolved, 1)
				require.Len(t, resp.UnResolved, 0)
				price, _ := resp.Resolved[tc.ticker].Value.Float64()
				require.InDelta(t, tc.expectedPrice, price, 0.0001)
				require.True(t, resp.Resolved[tc.ticker].Timestamp.After(time.Now().Add(-time.Second)))
			}
		})
	}
}
