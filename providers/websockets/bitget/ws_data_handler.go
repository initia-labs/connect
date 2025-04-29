package bitget

import (
	"bytes"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/skip-mev/connect/v2/oracle/config"
	"github.com/skip-mev/connect/v2/oracle/types"
	"github.com/skip-mev/connect/v2/providers/base/websocket/handlers"
)

var _ types.PriceWebSocketDataHandler = (*WebSocketHandler)(nil)

type WebSocketHandler struct {
	logger *zap.Logger

	ws config.WebSocketConfig

	cache types.ProviderTickers
}

func NewWebSocketDataHandler(
	logger *zap.Logger,
	ws config.WebSocketConfig,
) (types.PriceWebSocketDataHandler, error) {
	if ws.Name != Name {
		return nil, fmt.Errorf("expected websocket config name %s, got %s", Name, ws.Name)
	}

	if !ws.Enabled {
		return nil, fmt.Errorf("websocket config for %s is not enabled", Name)
	}

	if err := ws.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid websocket config for %s: %w", Name, err)
	}

	return &WebSocketHandler{
		logger: logger,
		ws:     ws,
		cache:  types.NewProviderTickers(),
	}, nil
}

func (h *WebSocketHandler) HandleMessage(
	message []byte,
) (types.PriceResponse, []handlers.WebsocketEncodedMessage, error) {
	var (
		resp          types.PriceResponse
		subscribeResp SubscriptionResponse
		//baseResp      BaseResponse
		update TickerUpdateMessage
	)

	// ping
	if string(message) == string(OperationPong) {
		h.logger.Debug("received ping response")
		return resp, nil, nil
	}

	// subscription response
	dec := json.NewDecoder(bytes.NewReader(message))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&subscribeResp); err == nil {
		if subscribeResp.Event == string(OperationSubscribe) {
			h.logger.Debug("received subscription response")
			return resp, nil, nil
		}
	}

	// price
	if err := json.Unmarshal(message, &update); err == nil {
		resp, err = h.parseTickerUpdate(update)
		if err != nil {
			return resp, nil, fmt.Errorf("failed to parse ticker update message: %w", err)
		}
	} else {
		h.logger.Debug("failed to unmarshal ticker update message", zap.Error(err), zap.Binary("message", message))
		return resp, nil, err
	}

	return resp, nil, nil
}

func (h *WebSocketHandler) CreateMessages(
	tickers []types.ProviderTicker,
) ([]handlers.WebsocketEncodedMessage, error) {
	pairs := make([]string, 0)

	for _, ticker := range tickers {
		pairs = append(pairs, ticker.GetOffChainTicker())
		h.cache.Add(ticker)
	}

	return h.NewSubscriptionRequestMessage(pairs)
}

func (h *WebSocketHandler) HeartBeatMessages() ([]handlers.WebsocketEncodedMessage, error) {
	return []handlers.WebsocketEncodedMessage{[]byte(OperationPing)}, nil
}

func (h *WebSocketHandler) Copy() types.PriceWebSocketDataHandler {
	return &WebSocketHandler{
		logger: h.logger,
		ws:     h.ws,
		cache:  types.NewProviderTickers(),
	}
}
