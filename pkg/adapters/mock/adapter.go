package mock

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/jonosize/affiliate-platform/pkg/adapters"
)

//go:embed fixtures/*.json
var fixturesFS embed.FS

// MockAdapter implements MarketplaceAdapter using JSON fixtures
type MockAdapter struct {
	products    map[string]*Product // key: source_id
	offers      map[string][]*Offer // key: product_source_id
	mu          sync.RWMutex
	rand        *rand.Rand           // Random generator for product selection
	marketplace adapters.Marketplace // The marketplace this adapter represents
}

// Product represents a product from fixtures
type Product struct {
	SourceID    string `json:"source_id"`
	SourceType  string `json:"source_type"`
	Marketplace string `json:"marketplace"`
	Title       string `json:"title"`
	ImageURL    string `json:"image_url"`
	URL         string `json:"url"`
}

// Offer represents an offer from fixtures
type Offer struct {
	ProductSourceID string  `json:"product_source_id"`
	Marketplace     string  `json:"marketplace"`
	StoreName       string  `json:"store_name"`
	Price           float64 `json:"price"`
	Currency        string  `json:"currency"`
	URL             string  `json:"url"`
	InStock         bool    `json:"in_stock"`
}

// NewAdapter creates a new mock adapter and loads fixtures
func NewAdapter() (*MockAdapter, error) {
	adapter := &MockAdapter{
		products: make(map[string]*Product),
		offers:   make(map[string][]*Offer),
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())), // Initialize random generator
	}

	// Load fixtures
	if err := adapter.loadFixtures(); err != nil {
		return nil, fmt.Errorf("failed to load fixtures: %w", err)
	}

	return adapter, nil
}

// NewAdapterForMarketplace creates a new mock adapter for a specific marketplace
func NewAdapterForMarketplace(marketplace adapters.Marketplace) (*MockAdapter, error) {
	adapter := &MockAdapter{
		products:    make(map[string]*Product),
		offers:      make(map[string][]*Offer),
		rand:        rand.New(rand.NewSource(time.Now().UnixNano())),
		marketplace: marketplace,
	}

	// Load fixtures
	if err := adapter.loadFixtures(); err != nil {
		return nil, fmt.Errorf("failed to load fixtures: %w", err)
	}

	return adapter, nil
}

func (a *MockAdapter) loadFixtures() error {
	// Load products.json
	productsData, err := fixturesFS.ReadFile("fixtures/products.json")
	if err != nil {
		return fmt.Errorf("failed to read products.json: %w", err)
	}
	var products []Product
	if err := json.Unmarshal(productsData, &products); err != nil {
		return fmt.Errorf("failed to unmarshal products.json: %w", err)
	}
	for i := range products {
		p := products[i] // Create a copy to avoid loop variable reuse bug
		a.products[p.SourceID] = &p
	}

	// Load offers.json
	offersData, err := fixturesFS.ReadFile("fixtures/offers.json")
	if err != nil {
		return fmt.Errorf("failed to read offers.json: %w", err)
	}
	var offers []Offer
	if err := json.Unmarshal(offersData, &offers); err != nil {
		return fmt.Errorf("failed to unmarshal offers.json: %w", err)
	}

	for i := range offers {
		o := offers[i] // Create a copy to avoid loop variable reuse bug
		a.offers[o.ProductSourceID] = append(a.offers[o.ProductSourceID], &o)
	}

	return nil
}

// FetchProduct fetches product details from source
func (a *MockAdapter) FetchProduct(ctx context.Context, source string, sourceType adapters.SourceType) (*adapters.ProductData, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Try to find product by source_id first
	for _, product := range a.products {
		if product.SourceID == source {
			return &adapters.ProductData{
				Title:                 product.Title,
				ImageURL:              product.ImageURL,
				MarketplaceProductURL: product.URL,
			}, nil
		}
	}

	// Try to find product by URL pattern matching
	// This allows matching any Lazada/Shopee URL to a mock product
	if sourceType == adapters.SourceTypeURL {
		sourceLower := strings.ToLower(source)
		var matchingProducts []*Product

		// Collect all products from matching marketplace
		for _, product := range a.products {
			if (strings.Contains(sourceLower, "lazada") && product.Marketplace == "lazada") ||
				(strings.Contains(sourceLower, "shopee") && product.Marketplace == "shopee") {
				matchingProducts = append(matchingProducts, product)
			}
		}

		// If found matching products, randomly select one from matching marketplace
		if len(matchingProducts) > 0 {
			// Sort by source_id for consistent ordering (map iteration order is non-deterministic)
			for i := 0; i < len(matchingProducts)-1; i++ {
				for j := i + 1; j < len(matchingProducts); j++ {
					if matchingProducts[i].SourceID > matchingProducts[j].SourceID {
						matchingProducts[i], matchingProducts[j] = matchingProducts[j], matchingProducts[i]
					}
				}
			}

			// Randomly select a product from matching marketplace
			var index int
			if a.rand == nil {
				index = 0 // Fallback to first product if rand is nil
			} else {
				index = a.rand.Intn(len(matchingProducts))
			}
			product := matchingProducts[index]

			return &adapters.ProductData{
				Title:                 product.Title,
				ImageURL:              product.ImageURL,
				MarketplaceProductURL: product.URL,
			}, nil
		}
	}

	// If still not found, return a product based on source hash for consistency
	// This ensures the same URL always returns the same product
	if len(a.products) > 0 {
		// Convert products map to slice for consistent ordering (sorted by source_id)
		productList := make([]*Product, 0, len(a.products))
		for _, product := range a.products {
			productList = append(productList, product)
		}

		// Sort by source_id for consistent ordering
		for i := 0; i < len(productList)-1; i++ {
			for j := i + 1; j < len(productList); j++ {
				if productList[i].SourceID > productList[j].SourceID {
					productList[i], productList[j] = productList[j], productList[i]
				}
			}
		}

		// Use source string as seed to pick a product consistently
		hash := 0
		for _, char := range source {
			hash = hash*31 + int(char)
		}
		if hash < 0 {
			hash = -hash
		}
		index := hash % len(productList)

		product := productList[index]
		return &adapters.ProductData{
			Title:                 product.Title,
			ImageURL:              product.ImageURL,
			MarketplaceProductURL: product.URL,
		}, nil
	}

	return nil, fmt.Errorf("product not found: %s", source)
}

// FetchOffer fetches current offer/price
func (a *MockAdapter) FetchOffer(ctx context.Context, productURL string) (*adapters.OfferData, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Find offer by URL - MUST match both URL and marketplace
	adapterMarketplace := a.Marketplace()
	for _, offers := range a.offers {
		for _, offer := range offers {
			// Only return offers that match both URL and the adapter's marketplace
			if offer.URL == productURL && offer.Marketplace == string(adapterMarketplace) {
				return &adapters.OfferData{
					StoreName:             offer.StoreName,
					Price:                 offer.Price,
					MarketplaceProductURL: offer.URL,
				}, nil
			}
		}
	}

	// Return first offer matching the adapter's marketplace as fallback
	// This ensures Lazada adapter returns Lazada offers and Shopee adapter returns Shopee offers
	for _, offers := range a.offers {
		for _, offer := range offers {
			// Filter by marketplace to ensure we return the correct marketplace's offer
			if offer.Marketplace == string(adapterMarketplace) {
				return &adapters.OfferData{
					StoreName:             offer.StoreName,
					Price:                 offer.Price,
					MarketplaceProductURL: offer.URL,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("offer not found for URL: %s in marketplace: %s", productURL, adapterMarketplace)
}

// Marketplace returns the marketplace identifier
func (a *MockAdapter) Marketplace() adapters.Marketplace {
	// If marketplace is explicitly set, return it
	if a.marketplace != "" {
		return a.marketplace
	}
	// Otherwise, return first available marketplace from fixtures (backward compatibility)
	if len(a.products) > 0 {
		for _, product := range a.products {
			if product.Marketplace == "lazada" {
				return adapters.MarketplaceLazada
			}
			if product.Marketplace == "shopee" {
				return adapters.MarketplaceShopee
			}
		}
	}
	return adapters.MarketplaceLazada
}
