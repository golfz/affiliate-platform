package mock

import (
	"github.com/jonosize/affiliate-platform/pkg/adapters"
)

// GetMockAdapters returns both Lazada and Shopee mock adapters
func GetMockAdapters() (adapters.MarketplaceAdapter, adapters.MarketplaceAdapter, error) {
	lazadaAdapter, err := NewAdapterForMarketplace(adapters.MarketplaceLazada)
	if err != nil {
		return nil, nil, err
	}

	shopeeAdapter, err := NewAdapterForMarketplace(adapters.MarketplaceShopee)
	if err != nil {
		return nil, nil, err
	}

	return lazadaAdapter, shopeeAdapter, nil
}
