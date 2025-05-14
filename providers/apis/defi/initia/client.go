package initia

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/skip-mev/connect/v2/oracle/config"
	connecthttp "github.com/skip-mev/connect/v2/pkg/http"
	"github.com/skip-mev/connect/v2/providers/base/api/metrics"
)

var (
	_ Client = &ClientImpl{}
	_ Client = &MultiClientImpl{}
)

type Client interface {
	SpotPrice(ctx context.Context, denom string) (WrappedInitiaSpotPrice, error)
}

type ClientImpl struct {
	api         config.APIConfig
	apiMetrics  metrics.APIMetrics
	redactedURL string
	endpoint    config.Endpoint
	httpClient  *connecthttp.Client
}

func NewClient(api config.APIConfig, apiMetrics metrics.APIMetrics, endpoint config.Endpoint) (Client, error) {
	if err := api.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}
	if api.Name != Name {
		return nil, fmt.Errorf("invalid config: name (%s) expected (%s)", api.Name, Name)
	}
	if !api.Enabled {
		return nil, fmt.Errorf("invalid config: disabled (%v)", api.Enabled)
	}
	if apiMetrics == nil {
		return nil, fmt.Errorf("invalid config: apiMetrics is nil")
	}
	redactedURL := metrics.RedactedEndpointURL(0)

	return &ClientImpl{
		api:         api,
		apiMetrics:  apiMetrics,
		redactedURL: redactedURL,
		endpoint:    endpoint,
		httpClient:  connecthttp.NewClient(),
	}, nil
}

func (c *ClientImpl) SpotPrice(ctx context.Context, denom string) (WrappedInitiaSpotPrice, error) {
	start := time.Now()
	defer func() {
		c.apiMetrics.ObserveProviderResponseLatency(c.api.Name, c.redactedURL, time.Since(start))
	}()

	urlEncodedDenom := strings.Replace(denom, "/", "%2F", 1)

	url, err := CreateURL(c.endpoint.URL, urlEncodedDenom)
	if err != nil {
		return WrappedInitiaSpotPrice{}, err
	}

	resp, err := c.httpClient.GetWithContext(ctx, url)
	if err != nil {
		return WrappedInitiaSpotPrice{}, err
	}

	c.apiMetrics.AddHTTPStatusCode(c.api.Name, resp)

	var response InitiaSpotPrice
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return WrappedInitiaSpotPrice{}, err
	}

	c.apiMetrics.AddHTTPStatusCode(c.api.Name, resp)

	return WrappedInitiaSpotPrice{
		InitiaSpotPrice: response,
		Timestamp:       start.Unix(),
	}, nil
}

type MultiClientImpl struct {
	logger     *zap.Logger
	api        config.APIConfig
	apiMetrics metrics.APIMetrics
	clients    []Client
}

func NewMultiClientFromEndpoints(logger *zap.Logger, api config.APIConfig, apiMetrics metrics.APIMetrics) (Client, error) {
	if err := api.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}
	if api.Name != Name {
		return nil, fmt.Errorf("invalid config: name (%s) expected (%s)", api.Name, Name)
	}
	if !api.Enabled {
		return nil, fmt.Errorf("invalid config: disabled (%v)", api.Enabled)
	}
	if apiMetrics == nil {
		return nil, fmt.Errorf("invalid config: apiMetrics is nil")
	}

	clients := make([]Client, 0, len(api.Endpoints))
	for _, endpoint := range api.Endpoints {
		c, err := NewClient(api, apiMetrics, endpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to create client: %w", err)
		}
		clients = append(clients, c)
	}

	return &MultiClientImpl{
		logger:     logger,
		api:        api,
		apiMetrics: apiMetrics,
		clients:    clients,
	}, nil
}

func (mc *MultiClientImpl) SpotPrice(ctx context.Context, denom string) (WrappedInitiaSpotPrice, error) {
	resps := make([]WrappedInitiaSpotPrice, len(mc.clients))

	var wg sync.WaitGroup
	wg.Add(len(mc.clients))

	for i := range mc.clients {
		url := mc.api.Endpoints[i].URL

		index := i
		go func(index int, client Client) {
			defer wg.Done()
			resp, err := client.SpotPrice(ctx, denom)
			if err != nil {
				mc.logger.Error("failed to spot price in sub client", zap.String("url", url), zap.Error(err))
				return
			}

			mc.logger.Debug("successfully fetched accounts", zap.String("url", url))

			resps[index] = resp

		}(index, mc.clients[i])
	}

	wg.Wait()

	return mc.latestSpotPriceResponse(resps)
}

func (mc *MultiClientImpl) latestSpotPriceResponse(responses []WrappedInitiaSpotPrice) (WrappedInitiaSpotPrice, error) {
	if len(responses) == 0 {
		return WrappedInitiaSpotPrice{}, fmt.Errorf("no responses found")
	}

	latest := int64(0)
	latestIndex := 0

	for i, resp := range responses {
		if resp.Timestamp > latest {
			latest = resp.Timestamp
			latestIndex = i
		}
	}

	return responses[latestIndex], nil
}
