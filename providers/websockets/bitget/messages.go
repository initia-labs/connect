package bitget

import (
	"encoding/json"
	"fmt"
	"math"

	connectmath "github.com/skip-mev/connect/v2/pkg/math"
	"github.com/skip-mev/connect/v2/providers/base/websocket/handlers"
)

type (
	Operation string

	Channel string
)

const (
	OperationSubscribe Operation = "subscribe"

	TickerChannel Channel = "ticker"

	OperationPing Operation = "ping"

	OperationPong Operation = "pong"
)

type BaseRequest struct {
	Op string `json:"op"`
}

//{
//	"op":"subscribe",
//	"args":[
//	{
//		"instType":"SPOT",
//		"channel":"ticker",
//		"instId":"BTCUSDT"
//	}
//	]
//}

type SubscriptionRequest struct {
	BaseRequest
	Args []ArgsData `json:"args"`
}

type ArgsData struct {
	InstType string `json:"instType"`
	Channel  string `json:"channel"`
	InstId   string `json:"instId"`
}

func (h *WebSocketHandler) NewSubscriptionRequestMessage(tickers []string) ([]handlers.WebsocketEncodedMessage, error) {
	numTickers := len(tickers)
	if numTickers == 0 {
		return nil, fmt.Errorf("tickers cannot be empty")
	}

	numBatches := int(math.Ceil(float64(numTickers) / float64(h.ws.MaxSubscriptionsPerBatch)))
	msgs := make([]handlers.WebsocketEncodedMessage, numBatches)
	for i := 0; i < numBatches; i++ {
		start := i * h.ws.MaxSubscriptionsPerBatch
		end := connectmath.Min((i+1)*h.ws.MaxSubscriptionsPerBatch, numTickers)

		args := make([]ArgsData, end-start)
		for j := 0; j < end-start; j++ {
			args[j] = ArgsData{
				InstType: "SPOT",
				Channel:  string(TickerChannel),
				InstId:   tickers[start+j],
			}
		}

		bz, err := json.Marshal(SubscriptionRequest{
			BaseRequest: BaseRequest{
				Op: string(OperationSubscribe),
			},
			Args: args,
		})
		if err != nil {
			return msgs, fmt.Errorf("unable to marshal message: %w", err)
		}

		msgs[i] = bz
	}

	return msgs, nil
}

type BaseResponse struct {
	Action string   `json:"action"`
	Arg    ArgsData `json:"arg"`
	Ts     int64    `json:"ts"`
}

//{
//	"event": "subscribe",
//	"arg": {
//	"instType": "SPOT",
//	"channel": "ticker",
//	"instId": "BTCUSDT"
//}

type SubscriptionResponse struct {
	Event string   `json:"event"`
	Arg   ArgsData `json:"arg"`
}

//{
//	"action": "snapshot",
//	"arg": {
//	"instType": "SPOT",
//	"channel": "ticker",
//	"instId": "BTCUSDT"
//	},
//	"data": [
//		{
//			"instId": "BTCUSDT",
//			"lastPr": "93655.32",
//			"open24h": "93217.41",
//			"high24h": "94409.16",
//			"low24h": "92301.72",
//			"change24h": "0.01277",
//			"bidPr": "93655.32",
//			"askPr": "93655.33",
//			"bidSz": "1.101507",
//			"askSz": "1.650978",
//			"baseVolume": "7833.994548",
//			"quoteVolume": "731157068.63006",
//			"openUtc": "93966.82",
//			"changeUtc24h": "-0.00331",
//			"ts": "1745577375326"
//		}
//	],
//	"ts": 1745577375332
//}

type TickerUpdateMessage struct {
	BaseResponse
	Data []TickerUpdateData `json:"data"`
}

type TickerUpdateData struct {
	InstId       string `json:"instId"`
	LastPr       string `json:"lastPr"`
	Open24H      string `json:"open24h"`
	High24H      string `json:"high24h"`
	Low24H       string `json:"low24h"`
	Change24H    string `json:"change24h"`
	BidPr        string `json:"bidPr"`
	AskPr        string `json:"askPr"`
	BidSz        string `json:"bidSz"`
	AskSz        string `json:"askSz"`
	BaseVolume   string `json:"baseVolume"`
	QuoteVolume  string `json:"quoteVolume"`
	OpenUtc      string `json:"openUtc"`
	ChangeUtc24H string `json:"changeUtc24h"`
	Ts           int64  `json:"ts,string"` // this is quoted string in JSON
}
