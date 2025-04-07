package geckoterminal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/skip-mev/connect/v2/oracle/config"
	"github.com/skip-mev/connect/v2/oracle/types"
	"github.com/skip-mev/connect/v2/pkg/math"
	providertypes "github.com/skip-mev/connect/v2/providers/types"
)

var _ types.PriceAPIDataHandler = (*APIHandler)(nil)

// APIHandler implements the PriceAPIDataHandler interface for GeckoTerminal.
type APIHandler struct {
	// apiCfg is the config for the GeckoTerminal API.
	api config.APIConfig
	// cache maintains the latest set of tickers seen by the handler.
	cache types.ProviderTickers
}

// NewAPIHandler returns a new GeckoTerminal PriceAPIDataHandler.
func NewAPIHandler(
	api config.APIConfig,
) (types.PriceAPIDataHandler, error) {
	if api.Name != Name {
		return nil, fmt.Errorf("expected api config name %s, got %s", Name, api.Name)
	}

	if !api.Enabled {
		return nil, fmt.Errorf("api config for %s is not enabled", Name)
	}

	if err := api.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid api config for %s: %w", Name, err)
	}

	return &APIHandler{
		api:   api,
		cache: types.NewProviderTickers(),
	}, nil
}

// getBaseAddress extracts the base address from the metadata JSON.
func (h *APIHandler) getBaseAddress(metadataJSONStr string) (string, error) {
	var metadata MetadataJSON
	if err := json.Unmarshal([]byte(metadataJSONStr), &metadata); err != nil {
		return "", fmt.Errorf("failed to parse metadata JSON: %w", err)
	}

	addresses := strings.Split(metadata.Address, ",")

	if len(addresses) < 2 {
		return "", fmt.Errorf("not enough addresses in metadata")
	}

	// Get the 2nd address
	// {\"address\":\"0X87428A53E14D24AB19C6CA4939B4DF93B8996CA9/
	// UNISWAP_V3,0X8236A87084F8B84306F72007F36F2618A5634494/
	// UNISWAP_V3,0X2260FAC5E5542A773AA44FBCFEDF7C193BC2C599\",\"base_decimals\":8,\"quote_decimals\":8,\"invert\":true}
	secondAddr := addresses[1]

	var targetAddress string
	if strings.Contains(secondAddr, "/") {
		parts := strings.Split(secondAddr, "/")
		targetAddress = parts[0]
	} else {
		targetAddress = secondAddr
	}

	return targetAddress, nil
}

// CreateURL returns the URL that is used to fetch data from the GeckoTerminal API for the
// given tickers. Note that the GeckoTerminal API supports fetching multiple spot prices
// iff they are all on the same chain.
func (h *APIHandler) CreateURL(
	tickers []types.ProviderTicker,
) (string, error) {
	if len(tickers) == 0 {
		return "", fmt.Errorf("no tickers provided")
	}

	metadataJSONStr := tickers[0].GetJSON()

	targetAddress, err := h.getBaseAddress(metadataJSONStr)
	if err != nil {
		return "", err
	}

	h.cache.Add(tickers[0])

	return fmt.Sprintf(h.api.Endpoints[0].URL, targetAddress), nil
}

// ParseResponse parses the response from the GeckoTerminal API. The response is expected
// to contain multiple spot prices for a given token address. Note that all of the tokens
// are shared on the same chain.
func (h *APIHandler) ParseResponse(
	tickers []types.ProviderTicker,
	resp *http.Response,
) types.PriceResponse {
	var result GeckoTerminalResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return types.NewPriceResponseWithErr(
			tickers,
			providertypes.NewErrorWithCode(err, providertypes.ErrorFailedToDecode),
		)
	}

	var (
		resolved   = make(types.ResolvedPrices)
		unresolved = make(types.UnResolvedPrices)
	)

	price := result.Data.Attributes.PriceUSD
	if price == "" {
		err := fmt.Errorf("no price found in response")
		return types.NewPriceResponseWithErr(
			tickers,
			providertypes.NewErrorWithCode(err, providertypes.ErrorNoResponse),
		)
	}

	priceFloat, err := math.Float64StringToBigFloat(price)
	if err != nil {
		wErr := fmt.Errorf("failed to convert price %s to big.Float: %w", price, err)
		unresolved[tickers[0]] = providertypes.UnresolvedResult{
			ErrorWithCode: providertypes.NewErrorWithCode(wErr, providertypes.ErrorFailedToParsePrice),
		}
		return types.NewPriceResponse(resolved, unresolved)
	}

	resolved[tickers[0]] = types.NewPriceResult(priceFloat, time.Now().UTC())

	// Add all expected tickers that did not return a response to the unresolved map
	for _, ticker := range tickers {
		_, resolvedOk := resolved[ticker]
		_, unresolvedOk := unresolved[ticker]

		if !resolvedOk && !unresolvedOk {
			err := fmt.Errorf("received no price response")
			unresolved[ticker] = providertypes.UnresolvedResult{
				ErrorWithCode: providertypes.NewErrorWithCode(err, providertypes.ErrorNoResponse),
			}
		}
	}

	return types.NewPriceResponse(resolved, unresolved)
}
