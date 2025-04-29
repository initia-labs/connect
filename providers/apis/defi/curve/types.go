package curve

type CurveMetadata struct {
	Network           string `json:"network"`
	PoolID            string `json:"pool_id"`
	BaseTokenAddress  string `json:"base_token_address"`
	QuoteTokenAddress string `json:"quote_token_address"`
}

func IsSupportedNetwork(network string) bool {
	networkMap := map[string]bool{
		"ethereum": true,
	}
	if _, ok := networkMap[network]; ok {
		return true
	}
	return false
}
