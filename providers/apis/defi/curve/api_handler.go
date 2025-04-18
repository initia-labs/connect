package curve

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

type APIHandler struct {
	api config.APIConfig

	cache types.ProviderTickers
}

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

func (h *APIHandler) CreateURL(
	tickers []types.ProviderTicker,
) (string, error) {
	var addresses = make([]string, len(tickers))
	var network string
	for i, ticker := range tickers {
		h.cache.Add(ticker)
		metadataJSON := ticker.GetJSON()
		var metadata CurveMetadata
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
			return h.api.Endpoints[0].URL, fmt.Errorf("failed to parse metadata JSON: %w", err)
		}
		network = metadata.Network
		addresses[i] = metadata.BaseTokenAddress
	}

	return fmt.Sprintf(h.api.Endpoints[0].URL, network, strings.Join(addresses, ",")), nil
}

func (h *APIHandler) ParseResponse(
	tickers []types.ProviderTicker,
	resp *http.Response,
) types.PriceResponse {
	var result CurveResponse
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

	data := result.Data
	ticker, ok := h.cache.FromOffChainTicker(data.Address)
	if !ok {
		err := fmt.Errorf("no ticker for address %s", data.Address)
		return types.NewPriceResponseWithErr(
			tickers,
			providertypes.NewErrorWithCode(err, providertypes.ErrorUnknownPair),
		)
	}

	price := math.Float64ToBigFloat(data.UsdPrice)

	resolved[ticker] = types.NewPriceResult(price, time.Now().UTC())

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
