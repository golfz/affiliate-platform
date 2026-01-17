package lazada

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jonosize/affiliate-platform/pkg/adapters"
)

// LazadaAdapter implements MarketplaceAdapter using Lazada Open Platform API
type LazadaAdapter struct {
	appKey      string
	appSecret   string
	accessToken string
	apiURL      string
	httpClient  *http.Client
}

// NewAdapter creates a new Lazada adapter with API credentials
func NewAdapter(appKey, appSecret, accessToken string) *LazadaAdapter {
	return &LazadaAdapter{
		appKey:      appKey,
		appSecret:   appSecret,
		accessToken: accessToken,
		apiURL:      "https://api.lazada.com.my/rest", // Default to Malaysia, can be configured
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// SetAPIURL allows setting custom API URL (for different regions)
func (a *LazadaAdapter) SetAPIURL(apiURL string) {
	a.apiURL = apiURL
}

// Marketplace returns the marketplace identifier
func (a *LazadaAdapter) Marketplace() adapters.Marketplace {
	return adapters.MarketplaceLazada
}

// FetchProduct fetches product details from URL or SKU
func (a *LazadaAdapter) FetchProduct(ctx context.Context, source string, sourceType adapters.SourceType) (*adapters.ProductData, error) {
	var itemID string
	var err error

	// Extract item ID from URL or use SKU directly
	if sourceType == adapters.SourceTypeURL {
		itemID, err = extractItemIDFromURL(source)
		if err != nil {
			return nil, fmt.Errorf("failed to extract item ID from URL: %w", err)
		}
	} else {
		itemID = source // Assume source is the item ID/SKU
	}

	// TODO: Add support for multiple regions (currently defaults to Malaysia)
	// TODO: Add retry logic for API failures
	// TODO: Add rate limiting to respect API quotas
	// Call Lazada API to get product details
	product, err := a.getProduct(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch product from Lazada API: %w", err)
	}

	// Extract first image if available
	imageURL := ""
	if len(product.Data.Images) > 0 {
		imageURL = product.Data.Images[0]
	}

	return &adapters.ProductData{
		Title:                 product.Data.Title,
		ImageURL:              imageURL,
		MarketplaceProductURL: product.Data.URL,
	}, nil
}

// FetchOffer fetches current offer/price
func (a *LazadaAdapter) FetchOffer(ctx context.Context, productURL string) (*adapters.OfferData, error) {
	// Extract item ID from URL
	itemID, err := extractItemIDFromURL(productURL)
	if err != nil {
		return nil, fmt.Errorf("failed to extract item ID from URL: %w", err)
	}

	// TODO: Add support for fetching multiple offers from different sellers
	// TODO: Add caching mechanism to reduce API calls
	// TODO: Add support for promotional prices and discounts
	// Call Lazada API to get product details (includes price)
	product, err := a.getProduct(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch offer from Lazada API: %w", err)
	}

	return &adapters.OfferData{
		StoreName:             product.Data.SellerName,
		Price:                 product.Data.Price,
		MarketplaceProductURL: productURL,
	}, nil
}

// LazadaProductResponse represents the response from Lazada API
type LazadaProductResponse struct {
	Code      string `json:"code"`
	RequestID string `json:"request_id"`
	Data      struct {
		ItemID     string   `json:"item_id"`
		Title      string   `json:"title"`
		Images     []string `json:"images"`
		Price      float64  `json:"price"`
		SellerName string   `json:"seller_name"`
		URL        string   `json:"item_url"`
	} `json:"data"`
	Message string `json:"message"`
}

// getProduct calls Lazada API to get product details
func (a *LazadaAdapter) getProduct(ctx context.Context, itemID string) (*LazadaProductResponse, error) {
	// Build request parameters
	params := url.Values{}
	params.Set("api", "product/get")
	params.Set("app_key", a.appKey)
	params.Set("access_token", a.accessToken)
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("item_id", itemID)
	params.Set("site", "my") // Default to Malaysia, can be configured

	// Generate signature
	signature := a.generateSignature(params)
	params.Set("sign", signature)

	// Build request URL
	reqURL := fmt.Sprintf("%s?%s", a.apiURL, params.Encode())

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Parse response
	var apiResp LazadaProductResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Code != "0" {
		return nil, fmt.Errorf("API error: %s - %s", apiResp.Code, apiResp.Message)
	}

	return &apiResp, nil
}

// generateSignature generates HMAC-SHA256 signature for Lazada API
func (a *LazadaAdapter) generateSignature(params url.Values) string {
	// Sort parameters by key
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" { // Exclude sign parameter
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// Build query string
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s%s", k, params.Get(k)))
	}
	queryString := strings.Join(parts, "")

	// Add app_secret at the end
	signString := queryString + a.appSecret

	// Calculate HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(a.appSecret))
	mac.Write([]byte(signString))
	signature := hex.EncodeToString(mac.Sum(nil))

	return strings.ToUpper(signature)
}

// extractItemIDFromURL extracts item ID from Lazada product URL
// Example: https://www.lazada.co.th/products/i123456-s789012.html -> 123456
func extractItemIDFromURL(productURL string) (string, error) {
	parsedURL, err := url.Parse(productURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Try to extract from path (e.g., /products/i123456-s789012.html)
	path := parsedURL.Path
	if strings.Contains(path, "/products/") {
		parts := strings.Split(path, "/products/")
		if len(parts) > 1 {
			// Extract item ID from format like "i123456-s789012.html"
			itemPart := strings.Split(parts[1], "-")[0]
			if strings.HasPrefix(itemPart, "i") {
				return strings.TrimPrefix(itemPart, "i"), nil
			}
		}
	}

	// Try to extract from query parameters
	itemID := parsedURL.Query().Get("item_id")
	if itemID != "" {
		return itemID, nil
	}

	return "", fmt.Errorf("could not extract item ID from URL: %s", productURL)
}
