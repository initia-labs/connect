package initia

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/skip-mev/connect/v2/oracle/config"
	"github.com/skip-mev/connect/v2/oracle/types"
)

const (
	Name      = "initia_api"
	Path      = "indexer/price/v1/prices/%s"
	Separator = "/"
)

var DefaultAPIConfig = config.APIConfig{
	Enabled:          true,
	Name:             Name,
	Timeout:          5 * time.Second,
	Interval:         3000 * time.Millisecond,
	ReconnectTimeout: 5 * time.Second,
	MaxQueries:       10,
	Atomic:           false,
	BatchSize:        1,
	Endpoints: []config.Endpoint{
		{
			URL: "https://dex-api.initia.xyz",
		},
	},
	MaxBlockHeightAge: 30 * time.Second,
}

func CreateURL(baseURL, denom string) (string, error) {
	return strings.Join(
		[]string{
			baseURL,
			fmt.Sprintf(Path, denom),
		},
		Separator,
	), nil
}

type metadataCache struct {
	metadataPerTicker map[string]InitiaMetadata
	mtx               sync.RWMutex
}

func newMetadataCache() *metadataCache {
	return &metadataCache{
		metadataPerTicker: make(map[string]InitiaMetadata),
	}
}

func (mc *metadataCache) getMetadataPerTicker(ticker types.ProviderTicker) (InitiaMetadata, bool) {
	mc.mtx.RLock()
	defer mc.mtx.RUnlock()

	metadata, ok := mc.metadataPerTicker[ticker.String()]
	return metadata, ok
}

func (mc *metadataCache) updateMetadataCache(ticker types.ProviderTicker) (InitiaMetadata, error) {
	mc.mtx.Lock()
	defer mc.mtx.Unlock()
	if metadata, ok := mc.metadataPerTicker[ticker.String()]; ok {
		return metadata, nil
	}
	metadata, err := unmarshalMetadataJSON(ticker.GetJSON())
	if err != nil {
		return InitiaMetadata{}, fmt.Errorf("error unmarshalling metadata for ticker %s: %w", ticker.String(), err)
	}
	if err := metadata.ValidateBasic(); err != nil {
		return InitiaMetadata{}, fmt.Errorf("metadata for ticker %s is invalid: %w", ticker.String(), err)
	}
	mc.metadataPerTicker[ticker.String()] = metadata
	return metadata, nil
}

type InitiaMetadata struct {
	BaseTokenDenom  string `json:"base_token_denom"`
	QuoteTokenDenom string `json:"quote_token_denom"`
	LPDenom         string `json:"lp_denom"`
}

func (im *InitiaMetadata) ValidateBasic() error {
	if im.BaseTokenDenom == "" || im.QuoteTokenDenom == "" || im.LPDenom == "" {
		return fmt.Errorf("base token denom, quote token denom, or lp cannot be empty")
	}
	return nil
}

func unmarshalMetadataJSON(metadata string) (InitiaMetadata, error) {
	var initiaMetadata InitiaMetadata
	if err := json.Unmarshal([]byte(metadata), &initiaMetadata); err != nil {
		return InitiaMetadata{}, err
	}
	return initiaMetadata, nil
}

type InitiaSpotPrice struct {
	Prices map[string]float64 `json:"prices"`
}

type WrappedInitiaSpotPrice struct {
	InitiaSpotPrice
	Timestamp int64 `json:"timestamp"`
}
