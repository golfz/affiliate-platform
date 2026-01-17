package mock

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	mathrand "math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jonosize/affiliate-platform/pkg/adapters"
)

var (
	globalRand     *mathrand.Rand
	globalRandMu   sync.Mutex
	globalRandOnce sync.Once
)

// getGlobalRand returns a thread-safe global random generator
func getGlobalRand() *mathrand.Rand {
	globalRandOnce.Do(func() {
		// Use crypto/rand for better seed
		var seed int64
		b := make([]byte, 8)
		if _, err := rand.Read(b); err == nil {
			seed = int64(binary.BigEndian.Uint64(b))
		} else {
			// Fallback to time-based seed
			seed = time.Now().UnixNano()
		}
		globalRand = mathrand.New(mathrand.NewSource(seed))
	})
	return globalRand
}

//go:embed fixtures/*.json
var fixturesFS embed.FS

// MockAdapter implements MarketplaceAdapter using JSON fixtures
type MockAdapter struct {
	products    map[string]*Product // key: source_id
	offers      map[string][]*Offer // key: product_source_id
	mu          sync.RWMutex
	rand        *mathrand.Rand       // Random generator for product selection (uses global rand)
	marketplace adapters.Marketplace // The marketplace this adapter represents
}

// ProductFixture represents a product from fixtures (merged structure)
type ProductFixture struct {
	SourceID  int        `json:"source_id"`
	Title     string     `json:"title"`
	ImageURL  string     `json:"image_url"`
	Platforms []Platform `json:"platforms"`
}

// Platform represents a marketplace platform/offer
type Platform struct {
	Marketplace string  `json:"marketplace"`
	StoreName   string  `json:"store_name"`
	Price       float64 `json:"price"`
	URL         string  `json:"url"`
}

// Product represents a product in the adapter (internal structure)
type Product struct {
	SourceID    int // Converted from fixture
	Title       string
	ImageURL    string
	URL         string // Primary URL (first platform's URL)
	Marketplace string // Primary marketplace (first platform's marketplace)
}

// Offer represents an offer in the adapter (internal structure)
type Offer struct {
	ProductSourceID int // Converted from fixture
	Marketplace     string
	StoreName       string
	Price           float64
	URL             string
}

// NewAdapter creates a new mock adapter and loads fixtures
func NewAdapter() (*MockAdapter, error) {
	adapter := &MockAdapter{
		products: make(map[string]*Product),
		offers:   make(map[string][]*Offer),
		rand:     getGlobalRand(), // Use shared global random generator
	}

	// Load fixtures
	if err := adapter.loadFixtures(); err != nil {
		return nil, fmt.Errorf("failed to load fixtures: %w", err)
	}

	return adapter, nil
}

// NewAdapterForMarketplace creates a new mock adapter for a specific marketplace
func NewAdapterForMarketplace(marketplace adapters.Marketplace) (*MockAdapter, error) {
	// Use global random generator for consistent randomness across adapter instances
	adapter := &MockAdapter{
		products:    make(map[string]*Product),
		offers:      make(map[string][]*Offer),
		rand:        getGlobalRand(), // Use shared global random generator
		marketplace: marketplace,
	}

	// Load fixtures
	if err := adapter.loadFixtures(); err != nil {
		return nil, fmt.Errorf("failed to load fixtures: %w", err)
	}

	return adapter, nil
}

func (a *MockAdapter) loadFixtures() error {
	// Load products.json (merged structure)
	productsData, err := fixturesFS.ReadFile("fixtures/products.json")
	if err != nil {
		return fmt.Errorf("failed to read products.json: %w", err)
	}
	var fixtures []ProductFixture
	if err := json.Unmarshal(productsData, &fixtures); err != nil {
		return fmt.Errorf("failed to unmarshal products.json: %w", err)
	}

	// Convert fixtures to internal Product and Offer structures
	for i := range fixtures {
		fixture := fixtures[i]
		sourceIDStr := strconv.Itoa(fixture.SourceID)

		// Determine primary URL and marketplace (use first platform)
		primaryURL := ""
		primaryMarketplace := ""
		if len(fixture.Platforms) > 0 {
			primaryURL = fixture.Platforms[0].URL
			primaryMarketplace = fixture.Platforms[0].Marketplace
		}

		// Create Product
		product := &Product{
			SourceID:    fixture.SourceID,
			Title:       fixture.Title,
			ImageURL:    fixture.ImageURL,
			URL:         primaryURL,
			Marketplace: primaryMarketplace,
		}
		a.products[sourceIDStr] = product

		// Create Offers from platforms
		for j := range fixture.Platforms {
			platform := fixture.Platforms[j]
			offer := &Offer{
				ProductSourceID: fixture.SourceID,
				Marketplace:     platform.Marketplace,
				StoreName:       platform.StoreName,
				Price:           platform.Price,
				URL:             platform.URL,
			}
			a.offers[sourceIDStr] = append(a.offers[sourceIDStr], offer)
		}
	}

	return nil
}

// FetchProduct fetches product details from source
func (a *MockAdapter) FetchProduct(ctx context.Context, source string, sourceType adapters.SourceType) (*adapters.ProductData, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Try to find product by source_id first
	// Convert source string to int for comparison
	if sourceID, err := strconv.Atoi(source); err == nil {
		if product, exists := a.products[source]; exists && product.SourceID == sourceID {
			return &adapters.ProductData{
				Title:                 product.Title,
				ImageURL:              product.ImageURL,
				MarketplaceProductURL: product.URL,
			}, nil
		}
	}

	// Try to find product by URL pattern matching
	// This allows matching any Lazada/Shopee URL to a mock product
	// For random selection, we select from ALL products, not just matching marketplace
	if sourceType == adapters.SourceTypeURL {
		var matchingProducts []*Product

		// Collect ALL products (not just matching marketplace) for random selection
		// This ensures better randomness when adding products
		for _, product := range a.products {
			matchingProducts = append(matchingProducts, product)
		}

		// If found products, randomly select one from all available products
		if len(matchingProducts) > 0 {
			// Sort by source_id for consistent ordering (map iteration order is non-deterministic)
			for i := 0; i < len(matchingProducts)-1; i++ {
				for j := i + 1; j < len(matchingProducts); j++ {
					if matchingProducts[i].SourceID > matchingProducts[j].SourceID {
						matchingProducts[i], matchingProducts[j] = matchingProducts[j], matchingProducts[i]
					}
				}
			}

			// Randomly select a product using shuffle for better distribution
			var product *Product
			if a.rand == nil {
				// Fallback to first product if rand is nil
				product = matchingProducts[0]
			} else {
				// Use thread-safe shuffle for better random distribution
				globalRandMu.Lock()
				// Shuffle the products array to get better randomness
				// This ensures each product has equal chance of being selected
				for i := len(matchingProducts) - 1; i > 0; i-- {
					j := a.rand.Intn(i + 1)
					matchingProducts[i], matchingProducts[j] = matchingProducts[j], matchingProducts[i]
				}
				// Select the first product after shuffle
				product = matchingProducts[0]
				globalRandMu.Unlock()
			}

			return &adapters.ProductData{
				Title:                 product.Title,
				ImageURL:              product.ImageURL,
				MarketplaceProductURL: product.URL,
				SourceID:              product.SourceID,
			}, nil
		}
	}

	// If still not found, randomly select a product from all available products
	// This ensures random selection when URL doesn't match any specific product
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

		// Randomly select a product from all available products
		var index int
		if a.rand == nil {
			// Fallback: use hash-based selection if rand is nil
			hash := 0
			for _, char := range source {
				hash = hash*31 + int(char)
			}
			if hash < 0 {
				hash = -hash
			}
			index = hash % len(productList)
		} else {
			// Use thread-safe random selection
			globalRandMu.Lock()
			index = a.rand.Intn(len(productList))
			globalRandMu.Unlock()
		}

		product := productList[index]

		return &adapters.ProductData{
			Title:                 product.Title,
			ImageURL:              product.ImageURL,
			MarketplaceProductURL: product.URL,
			SourceID:              product.SourceID,
		}, nil
	}

	return nil, fmt.Errorf("product not found: %s", source)
}

// FetchOfferBySourceID fetches offer by source_id and marketplace
// This is used when we know the source_id from FetchProduct and want to get the offer for a specific marketplace
func (a *MockAdapter) FetchOfferBySourceID(ctx context.Context, sourceID int, marketplace adapters.Marketplace) (*adapters.OfferData, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	sourceIDStr := strconv.Itoa(sourceID)
	offers, exists := a.offers[sourceIDStr]
	if !exists {
		return nil, fmt.Errorf("offers not found for source_id: %d", sourceID)
	}

	// Find offer matching the marketplace
	for _, offer := range offers {
		if offer.Marketplace == string(marketplace) {
			return &adapters.OfferData{
				StoreName:             offer.StoreName,
				Price:                 offer.Price,
				MarketplaceProductURL: offer.URL,
			}, nil
		}
	}

	return nil, fmt.Errorf("offer not found for source_id: %d in marketplace: %s", sourceID, marketplace)
}

// FetchOffer fetches current offer/price
func (a *MockAdapter) FetchOffer(ctx context.Context, productURL string) (*adapters.OfferData, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Note: FetchOffer now relies on FetchOfferBySourceID being called directly from CreateProduct
	// when source_id is available in ProductData. This method is kept for backward compatibility
	// and for cases where URL matching is needed.

	// Find offer by URL - MUST match both URL and marketplace
	adapterMarketplace := a.Marketplace()

	// Normalize URLs for comparison (remove query parameters and trailing slashes)
	normalizeURL := func(url string) string {
		url = strings.TrimSpace(url)
		url = strings.TrimSuffix(url, "/")
		// Remove query parameters for comparison
		if idx := strings.Index(url, "?"); idx != -1 {
			url = url[:idx]
		}
		return url
	}

	normalizedProductURL := normalizeURL(productURL)

	// First, try exact match
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

	// If exact match fails, try normalized URL matching
	// This handles cases where the URL might have query parameters or slight variations
	for _, offers := range a.offers {
		for _, offer := range offers {
			if offer.Marketplace == string(adapterMarketplace) {
				normalizedOfferURL := normalizeURL(offer.URL)
				if normalizedOfferURL == normalizedProductURL {
					return &adapters.OfferData{
						StoreName:             offer.StoreName,
						Price:                 offer.Price,
						MarketplaceProductURL: offer.URL,
					}, nil
				}
			}
		}
	}

	// If still not found, return error - DO NOT fallback to random offer
	// This prevents returning offers from wrong products
	// The old fallback logic was returning the first offer matching the marketplace,
	// which could be from a completely different product, causing incorrect links
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
