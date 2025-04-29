package curve

import (
	"encoding/json"
	"github.com/skip-mev/connect/v2/providers/base/testutils"
	providertypes "github.com/skip-mev/connect/v2/providers/types"
	"net/http"
	"testing"
	"time"

	"github.com/skip-mev/connect/v2/oracle/config"
	"github.com/skip-mev/connect/v2/oracle/types"
	"github.com/stretchr/testify/require"
)

func TestNewAPIHandler(t *testing.T) {
	tests := []struct {
		name      string
		apiConfig config.APIConfig
		wantErr   bool
	}{
		{
			name:      "success - default config",
			apiConfig: DefaultAPIConfig,
			wantErr:   false,
		},
		{
			name: "failure - wrong API name",
			apiConfig: config.APIConfig{
				Name:    "wrong_name",
				Enabled: true,
			},
			wantErr: true,
		},
		{
			name: "failure - disabled API",
			apiConfig: config.APIConfig{
				Name:    Name,
				Enabled: false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := NewAPIHandler(tt.apiConfig)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, handler)
		})
	}
}

func TestCreateURL(t *testing.T) {
	handler, err := NewAPIHandler(DefaultAPIConfig)
	require.NoError(t, err)

	tests := []struct {
		name     string
		tickers  []types.ProviderTicker
		wantURL  string
		wantErr  bool
		metadata CurveMetadata
	}{
		{
			name:    "failure - empty tickers",
			tickers: []types.ProviderTicker{},
			wantErr: true,
		},
		{
			name: "success - single ticker",
			tickers: []types.ProviderTicker{
				createTickerWithMetadata(t, "ETH", "USD", CurveMetadata{
					Network:          "ethereum",
					BaseTokenAddress: "0x123",
				}),
			},
			metadata: CurveMetadata{
				Network:          "ethereum",
				BaseTokenAddress: "0x123",
			},
			wantURL: "https://prices.curve.fi/v1/usd_price/ethereum",
			wantErr: false,
		},
		{
			name: "failure - missing network",
			tickers: []types.ProviderTicker{
				createTickerWithMetadata(t, "ETH", "USD", CurveMetadata{
					BaseTokenAddress: "0x123",
				}),
			},
			metadata: CurveMetadata{
				BaseTokenAddress: "0x123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := handler.CreateURL(tt.tickers)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantURL, url)
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
				OffChainTicker: "0xCd5fE23C85820F7B72D0926FC9b05b43E359b7ee",
				JSON:           `{"network":"ethereum","base_token_address":"0xCd5fE23C85820F7B72D0926FC9b05b43E359b7ee"}`,
			},
			response: testutils.CreateResponseFromJSON(`{
                "data": [{
                    "address": "0xCd5fE23C85820F7B72D0926FC9b05b43E359b7ee",
                    "usd_price": 1674.1742629502855,
                    "last_updated": "2025-04-16T06:04:23"
                }]
            }`),
			expectedPrice: 1674.1742629502855,
			expectError:   false,
		},
		{
			name: "invalid JSON response",
			ticker: types.DefaultProviderTicker{
				OffChainTicker: "0xAddress",
				JSON:           `{"network":"ethereum","base_token_address":"0xAddress"}`,
			},
			response:    testutils.CreateResponseFromJSON(`{invalid json`),
			expectError: true,
			errorCode:   providertypes.ErrorFailedToDecode,
		},
		{
			name: "empty data response",
			ticker: types.DefaultProviderTicker{
				OffChainTicker: "0xAddress",
				JSON:           `{"network":"ethereum","base_token_address":"0xAddress"}`,
			},
			response:    testutils.CreateResponseFromJSON(`{"data": {whole list of ethereum ...}}`),
			expectError: true,
			errorCode:   providertypes.ErrorFailedToDecode,
		},
		{
			name: "unresolved ticker",
			ticker: types.DefaultProviderTicker{
				OffChainTicker: "0xUnknownAddress",
				JSON:           `{"network":"ethereum","base_token_address":"0xUnknownAddress"}`,
			},
			response:    testutils.CreateResponseFromJSON(`{"detail":"Token data not found"}`),
			expectError: true,
			errorCode:   providertypes.ErrorNoResponse,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h, err := NewAPIHandler(DefaultAPIConfig)
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

// createTickerWithMetadata is a helper function to create a ProviderTicker with metadata
func createTickerWithMetadata(t *testing.T, base, quote string, metadata CurveMetadata) types.ProviderTicker {
	metadataBytes, err := json.Marshal(metadata)
	require.NoError(t, err)

	return types.DefaultProviderTicker{
		OffChainTicker: base + "/" + quote,
		JSON:           string(metadataBytes),
	}
}
