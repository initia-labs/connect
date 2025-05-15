package initia

import (
	"context"
	"fmt"
	"github.com/skip-mev/connect/v2/pkg/math"
	"math/big"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/skip-mev/connect/v2/oracle/config"
	oracletypes "github.com/skip-mev/connect/v2/oracle/types"
	"github.com/skip-mev/connect/v2/providers/base/api/metrics"
	providertypes "github.com/skip-mev/connect/v2/providers/types"
)

var _ oracletypes.PriceAPIFetcher = &APIPriceFetcher{}

type APIPriceFetcher struct {
	api               config.APIConfig
	client            Client
	logger            *zap.Logger
	metadataPerTicker *metadataCache
}

func NewAPIPriceFetcher(logger *zap.Logger, api config.APIConfig, apiMetrics metrics.APIMetrics) (*APIPriceFetcher, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if err := api.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid api config: %w", err)
	}
	if api.Name != Name {
		return nil, fmt.Errorf("invalid api name; expected %s, got %s", Name, api.Name)
	}
	if !api.Enabled {
		return nil, fmt.Errorf("api is disabled")
	}
	if apiMetrics == nil {
		return nil, fmt.Errorf("metrics cannot be nil")
	}

	client, err := NewMultiClientFromEndpoints(logger, api, apiMetrics)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	return &APIPriceFetcher{
		api:               api,
		client:            client,
		logger:            logger.With(zap.String("fetcher", Name)),
		metadataPerTicker: newMetadataCache(),
	}, nil
}

func (pf *APIPriceFetcher) Fetch(
	ctx context.Context,
	tickers []oracletypes.ProviderTicker,
) oracletypes.PriceResponse {
	resolved := make(oracletypes.ResolvedPrices)
	unresolved := make(oracletypes.UnResolvedPrices)

	g, ctx := errgroup.WithContext(ctx)
	unresolvedMutex := sync.Mutex{}
	resolvedMutex := sync.Mutex{}
	g.SetLimit(pf.api.MaxQueries)

	unresolvedTickerCallback := func(ticker oracletypes.ProviderTicker, err providertypes.ErrorWithCode) {
		unresolvedMutex.Lock()
		defer unresolvedMutex.Unlock()
		unresolved[ticker] = providertypes.UnresolvedResult{
			ErrorWithCode: err,
		}
	}

	resolvedTickerCallback := func(ticker oracletypes.ProviderTicker, price *big.Float) {
		resolvedMutex.Lock()
		defer resolvedMutex.Unlock()
		resolved[ticker] = oracletypes.NewPriceResult(price, time.Now().UTC())
	}

	pf.logger.Info("fetching for tickers", zap.Any("tickers", tickers))
	
	for _, ticker := range tickers {
		_, found := pf.metadataPerTicker.getMetadataPerTicker(ticker)
		if !found {
			_, err := pf.metadataPerTicker.updateMetadataCache(ticker)
			if err != nil {
				pf.logger.Debug("failed to update metadata cache", zap.Error(err))
			}
		}
	}

	for _, ticker := range tickers {
		g.Go(func() error {
			ticker := ticker
			var err error

			metadata, found := pf.metadataPerTicker.getMetadataPerTicker(ticker)
			if !found {
				err = fmt.Errorf("no metadata found for ticker %s", ticker.String())
				unresolvedTickerCallback(ticker, providertypes.NewErrorWithCode(err, providertypes.ErrorTickerMetadataNotFound))
				return nil
			}
			callCtx, cancel := context.WithTimeout(ctx, pf.api.Timeout)
			defer cancel()

			resp, err := pf.client.SpotPrice(callCtx, metadata.BaseTokenDenom)
			if err != nil {
				unresolvedTickerCallback(ticker, providertypes.NewErrorWithCode(err, providertypes.ErrorAPIGeneral))
				return nil
			}

			rawPrice, found := resp.Prices[metadata.BaseTokenDenom]
			if !found {
				unresolvedTickerCallback(ticker, providertypes.NewErrorWithCode(fmt.Errorf("price not found in response"), providertypes.ErrorAPIGeneral))
				return nil
			}

			price := math.Float64ToBigFloat(rawPrice)
			resolvedTickerCallback(ticker, price)

			return nil
		})
	}

	_ = g.Wait()

	return oracletypes.NewPriceResponse(resolved, unresolved)
}
