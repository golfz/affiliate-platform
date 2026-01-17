# Lazada Adapter

Real Lazada Open Platform API adapter implementation.

## Requirements

To use the real Lazada API, you need:

1. **Lazada Open Platform Account**
   - Register at: https://open.lazada.com/
   - Create an application to get `app_key` and `app_secret`

2. **Access Token**
   - Obtain an access token through OAuth flow
   - See: https://open.lazada.com/apps/doc/authorize

3. **API Credentials**
   - `app_key`: Your application key
   - `app_secret`: Your application secret
   - `access_token`: OAuth access token

## Usage

```go
import "github.com/jonosize/affiliate-platform/pkg/adapters/lazada"

// Create adapter with credentials
adapter := lazada.NewAdapter(
    "YOUR_APP_KEY",
    "YOUR_APP_SECRET", 
    "YOUR_ACCESS_TOKEN",
)

// Optional: Set API URL for different regions
// Thailand: https://api.lazada.co.th/rest
// Malaysia: https://api.lazada.com.my/rest
// Singapore: https://api.lazada.sg/rest
adapter.SetAPIURL("https://api.lazada.co.th/rest")

// Fetch product by URL
product, err := adapter.FetchProduct(ctx, "https://www.lazada.co.th/products/i123456-s789012.html", adapters.SourceTypeURL)
if err != nil {
    // Handle error
}

// Fetch offer/price
offer, err := adapter.FetchOffer(ctx, "https://www.lazada.co.th/products/i123456-s789012.html")
if err != nil {
    // Handle error
}
```

## API Documentation

- Official API Docs: https://open.lazada.com/apps/doc/api?path=%2Fproducts%2Fget
- Authentication: https://open.lazada.com/apps/doc/authorize

## Limitations

- Requires valid Lazada seller account or partner access
- Access token may expire and need refresh
- API rate limits may apply
- Some products may only be accessible by their owners

## Integration

To use this adapter instead of mock adapter, update `internal/service/product.go`:

```go
// Replace mock adapter with real Lazada adapter
lazadaAdapter := lazada.NewAdapter(
    cfg.GetLazadaAppKey(),
    cfg.GetLazadaAppSecret(),
    cfg.GetLazadaAccessToken(),
)
```
