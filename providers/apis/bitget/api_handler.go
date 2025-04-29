package bitget

import (
	"encoding/json"
	"fmt"
	"github.com/skip-mev/connect/v2/oracle/config"
	"github.com/skip-mev/connect/v2/oracle/types"
	"github.com/skip-mev/connect/v2/pkg/math"
	providertypes "github.com/skip-mev/connect/v2/providers/types"
	"net/http"
	"time"
)

var _ types.PriceAPIDataHandler = (*APIHandler)(nil)

type APIHandler struct {
	// api is the config for the Coinbase API.
	api config.APIConfig
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
		api: api,
	}, nil
}

func (h *APIHandler) CreateURL(
	tickers []types.ProviderTicker,
) (string, error) {
	if len(tickers) != 1 {
		return "", fmt.Errorf("expected 1 ticker, got %d", len(tickers))
	}
	return fmt.Sprintf(h.api.Endpoints[0].URL, tickers[0].GetOffChainTicker()), nil
}

func (h *APIHandler) ParseResponse(
	tickers []types.ProviderTicker,
	resp *http.Response,
) types.PriceResponse {
	if len(tickers) != 1 {
		return types.NewPriceResponseWithErr(
			tickers,
			providertypes.NewErrorWithCode(
				fmt.Errorf("expected 1 ticker, got %d", len(tickers)),
				providertypes.ErrorInvalidResponse,
			),
		)
	}

	var result BitgetResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return types.NewPriceResponseWithErr(
			tickers,
			providertypes.NewErrorWithCode(err, providertypes.ErrorFailedToDecode),
		)
	}

	ticker := tickers[0]
	price, err := math.Float64StringToBigFloat(result.Data[0].LastPr)
	if err != nil {
		return types.NewPriceResponseWithErr(
			tickers, providertypes.NewErrorWithCode(err, providertypes.ErrorFailedToParsePrice),
		)
	}

	return types.NewPriceResponse(
		types.ResolvedPrices{
			ticker: types.NewPriceResult(price, time.Now().UTC()),
		}, nil)
}
