package bitget

import (
	"fmt"
	"strings"
	"time"

	"github.com/skip-mev/connect/v2/oracle/types"
	"github.com/skip-mev/connect/v2/pkg/math"
	providertypes "github.com/skip-mev/connect/v2/providers/types"
)

func (h *WebSocketHandler) parseTickerUpdate(
	resp TickerUpdateMessage,
) (types.PriceResponse, error) {
	var (
		resolved   = make(types.ResolvedPrices)
		unresolved = make(types.UnResolvedPrices)
	)

	if !strings.Contains(resp.Arg.Channel, string(TickerChannel)) {
		return types.NewPriceResponse(resolved, unresolved), fmt.Errorf("invalid channel %s", resp.Arg.Channel)
	}

	for _, data := range resp.Data {
		ticker, ok := h.cache.FromOffChainTicker(data.InstId)
		if !ok {
			return types.NewPriceResponse(resolved, unresolved), fmt.Errorf("unknown ticker %s", data.InstId)
		}

		price, err := math.Float64StringToBigFloat(data.LastPr)
		if err != nil {
			wErr := fmt.Errorf("failed to convert price to big.Float: %w", err)
			unresolved[ticker] = providertypes.UnresolvedResult{
				ErrorWithCode: providertypes.NewErrorWithCode(wErr, providertypes.ErrorFailedToParsePrice),
			}
			return types.NewPriceResponse(resolved, unresolved), nil
		}

		resolved[ticker] = types.NewPriceResult(price, time.Now().UTC())
	}

	return types.NewPriceResponse(resolved, unresolved), nil
}
