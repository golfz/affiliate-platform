package adapters

import (
	"context"
)

// MarketplaceAdapter defines the interface for marketplace adapters
type MarketplaceAdapter interface {
	// FetchProduct fetches product details from URL or SKU
	FetchProduct(ctx context.Context, source string, sourceType SourceType) (*ProductData, error)

	// FetchOffer fetches current offer/price
	FetchOffer(ctx context.Context, productURL string) (*OfferData, error)

	// Marketplace returns the marketplace identifier
	Marketplace() Marketplace
}

type SourceType string

const (
	SourceTypeURL SourceType = "url"
	SourceTypeSKU SourceType = "sku"
)

type Marketplace string

const (
	MarketplaceLazada Marketplace = "lazada"
	MarketplaceShopee Marketplace = "shopee"
)

type ProductData struct {
	Title                 string `json:"title"`
	ImageURL              string `json:"image_url"`
	MarketplaceProductURL string `json:"marketplace_product_url"`
	SourceID              int    `json:"source_id,omitempty"` // Optional: source_id from mock adapter (0 if not set)
}

type OfferData struct {
	StoreName             string  `json:"store_name"`
	Price                 float64 `json:"price"`
	MarketplaceProductURL string  `json:"marketplace_product_url"`
}
