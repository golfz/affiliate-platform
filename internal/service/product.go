package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jonosize/affiliate-platform/internal/database"
	"github.com/jonosize/affiliate-platform/internal/dto"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/model"
	"github.com/jonosize/affiliate-platform/internal/repository"
	"github.com/jonosize/affiliate-platform/internal/validator"
	"github.com/jonosize/affiliate-platform/pkg/adapters"
	"github.com/jonosize/affiliate-platform/pkg/adapters/mock"
)

// ProductService handles product business logic
type ProductService struct {
	productRepo *repository.ProductRepository
	offerRepo   *repository.OfferRepository
	db          *database.DB
	logger      logger.Logger
}

// NewProductService creates a new product service
func NewProductService(db *database.DB, log logger.Logger) *ProductService {
	return &ProductService{
		productRepo: repository.NewProductRepository(db),
		offerRepo:   repository.NewOfferRepository(db),
		db:          db,
		logger:      log,
	}
}

// CreateProduct creates a product from Lazada and/or Shopee URLs
func (s *ProductService) CreateProduct(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error) {
	if req.LazadaURL == "" && req.ShopeeURL == "" {
		return nil, fmt.Errorf("at least one URL (Lazada or Shopee) must be provided")
	}

	// Get adapters (using mock adapters for now)
	lazadaAdapter, shopeeAdapter, err := mock.GetMockAdapters()
	if err != nil {
		return nil, fmt.Errorf("failed to get adapters: %w", err)
	}

	var productTitle string
	var productImageURL string
	var primaryProductURL string
	var randomSourceID int // To store the source_id if a random product is selected

	// Try to fetch product data from Lazada first if URL is provided
	if req.LazadaURL != "" {
		if _, _, err := validator.ValidateProductURL(req.LazadaURL); err != nil {
			return nil, fmt.Errorf("invalid Lazada URL: %w", err)
		}
		productData, err := lazadaAdapter.FetchProduct(ctx, req.LazadaURL, adapters.SourceTypeURL)
		if err == nil && productData != nil {
			productTitle = productData.Title
			productImageURL = productData.ImageURL
			primaryProductURL = productData.MarketplaceProductURL
			randomSourceID = productData.SourceID // Capture source ID from mock adapter
		} else {
			s.logger.Warn("Failed to fetch Lazada product data", logger.Error(err), logger.String("url", req.LazadaURL))
		}
	}

	// If no Lazada URL or fetching failed, try Shopee if URL is provided
	if productTitle == "" && req.ShopeeURL != "" {
		if _, _, err := validator.ValidateProductURL(req.ShopeeURL); err != nil {
			return nil, fmt.Errorf("invalid Shopee URL: %w", err)
		}
		productData, err := shopeeAdapter.FetchProduct(ctx, req.ShopeeURL, adapters.SourceTypeURL)
		if err == nil && productData != nil {
			productTitle = productData.Title
			productImageURL = productData.ImageURL
			primaryProductURL = productData.MarketplaceProductURL
			randomSourceID = productData.SourceID // Capture source ID from mock adapter
		} else {
			s.logger.Warn("Failed to fetch Shopee product data", logger.Error(err), logger.String("url", req.ShopeeURL))
		}
	}

	// If no specific URL was provided, but we still need a product (e.g., for random selection)
	if productTitle == "" && req.LazadaURL == "" && req.ShopeeURL == "" {
		// Call FetchProduct with a dummy URL to trigger random selection in mock adapter
		// The actual URL doesn't matter here, as mock adapter will pick a random product
		productData, err := lazadaAdapter.FetchProduct(ctx, "https://www.lazada.co.th/products/random", adapters.SourceTypeURL)
		if err == nil && productData != nil {
			productTitle = productData.Title
			productImageURL = productData.ImageURL
			primaryProductURL = productData.MarketplaceProductURL
			randomSourceID = productData.SourceID // Capture source ID from mock adapter
		} else {
			s.logger.Error("Failed to fetch random product data", logger.Error(err))
			return nil, fmt.Errorf("failed to fetch product data from any provided URL or random selection")
		}
	} else if productTitle == "" {
		return nil, fmt.Errorf("failed to fetch product data from any provided URL")
	}

	// Create product
	product := &model.Product{
		Title:    productTitle,
		ImageURL: productImageURL,
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// Fetch offers from adapters
	offers := make([]*model.Offer, 0)

	// Determine which marketplaces to fetch offers for
	// If specific URLs are provided, use those; otherwise try both
	hasLazadaURL := req.LazadaURL != ""
	hasShopeeURL := req.ShopeeURL != ""

	// Check if source_id is available (from FetchProduct)
	// If yes, use FetchOfferBySourceID to get offers for the random product
	hasSourceID := randomSourceID > 0

	// Fetch Lazada offer if:
	// 1. Lazada URL is explicitly provided, OR
	// 2. No URLs are provided (fetch both platforms)
	if hasLazadaURL || (!hasLazadaURL && !hasShopeeURL) {
		var offerData *adapters.OfferData
		var err error

		// If we have source_id from FetchProduct, use FetchOfferBySourceID
		// Otherwise, try FetchOffer with URL
		if hasSourceID {
			// Cast to MockAdapter to access FetchOfferBySourceID
			if mockAdapter, ok := lazadaAdapter.(*mock.MockAdapter); ok {
				offerData, err = mockAdapter.FetchOfferBySourceID(ctx, randomSourceID, adapters.MarketplaceLazada)
			} else {
				// Fallback to FetchOffer if not MockAdapter
				var lazadaOfferURL string
				if hasLazadaURL {
					lazadaOfferURL = req.LazadaURL
				} else {
					lazadaOfferURL = primaryProductURL
				}
				offerData, err = lazadaAdapter.FetchOffer(ctx, lazadaOfferURL)
			}
		} else {
			// Use URL-based FetchOffer
			var lazadaOfferURL string
			if hasLazadaURL {
				lazadaOfferURL = req.LazadaURL
			} else {
				lazadaOfferURL = primaryProductURL
			}
			offerData, err = lazadaAdapter.FetchOffer(ctx, lazadaOfferURL)
		}

		if err == nil && offerData != nil {
			offer := &model.Offer{
				ProductID:             product.ID,
				Marketplace:           model.Marketplace(adapters.MarketplaceLazada),
				StoreName:             offerData.StoreName,
				Price:                 offerData.Price,
				MarketplaceProductURL: offerData.MarketplaceProductURL,
				LastCheckedAt:         time.Now(),
			}
			offers = append(offers, offer)
		}
	}

	// Fetch Shopee offer if:
	// 1. Shopee URL is explicitly provided, OR
	// 2. No URLs are provided (fetch both platforms)
	if hasShopeeURL || (!hasLazadaURL && !hasShopeeURL) {
		var offerData *adapters.OfferData
		var err error

		// If we have source_id from FetchProduct, use FetchOfferBySourceID
		// Otherwise, try FetchOffer with URL
		if hasSourceID {
			// Cast to MockAdapter to access FetchOfferBySourceID
			if mockAdapter, ok := shopeeAdapter.(*mock.MockAdapter); ok {
				offerData, err = mockAdapter.FetchOfferBySourceID(ctx, randomSourceID, adapters.MarketplaceShopee)
			} else {
				// Fallback to FetchOffer if not MockAdapter
				var shopeeOfferURL string
				if hasShopeeURL {
					shopeeOfferURL = req.ShopeeURL
				} else {
					shopeeOfferURL = primaryProductURL
				}
				offerData, err = shopeeAdapter.FetchOffer(ctx, shopeeOfferURL)
			}
		} else {
			// Use URL-based FetchOffer
			var shopeeOfferURL string
			if hasShopeeURL {
				shopeeOfferURL = req.ShopeeURL
			} else {
				shopeeOfferURL = primaryProductURL
			}
			offerData, err = shopeeAdapter.FetchOffer(ctx, shopeeOfferURL)
		}

		if err == nil && offerData != nil {
			offer := &model.Offer{
				ProductID:             product.ID,
				Marketplace:           model.Marketplace(adapters.MarketplaceShopee),
				StoreName:             offerData.StoreName,
				Price:                 offerData.Price,
				MarketplaceProductURL: offerData.MarketplaceProductURL,
				LastCheckedAt:         time.Now(),
			}
			offers = append(offers, offer)
		}
	}

	// Save offers
	for _, offer := range offers {
		if err := s.offerRepo.Upsert(ctx, offer); err != nil {
			s.logger.Error("Failed to save offer", logger.Error(err), logger.String("marketplace", string(offer.Marketplace)))
			// Continue with other offers
		}
	}

	// Convert to response
	response := &dto.ProductResponse{
		ID:        product.ID,
		Title:     product.Title,
		ImageURL:  product.ImageURL,
		CreatedAt: product.CreatedAt,
	}

	// Fetch offers from DB to include in response
	dbOffers, err := s.offerRepo.FindByProductID(ctx, product.ID)
	if err == nil {
		response.Offers = make([]dto.OfferResponse, len(dbOffers))
		for i, o := range dbOffers {
			response.Offers[i] = dto.OfferResponse{
				ID:            o.ID,
				Marketplace:   string(o.Marketplace),
				StoreName:     o.StoreName,
				Price:         o.Price,
				LastCheckedAt: o.LastCheckedAt,
			}
		}
	}

	return response, nil
}

// GetProductOffers gets offers for a product
func (s *ProductService) GetProductOffers(ctx context.Context, productID uuid.UUID) (*dto.ProductOffersResponse, error) {
	// Check if product exists
	product, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Get offers
	offers, err := s.offerRepo.FindByProductID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get offers: %w", err)
	}

	// Convert to response
	response := &dto.ProductOffersResponse{
		ProductID: product.ID,
		Offers:    make([]dto.OfferResponse, len(offers)),
	}

	for i, offer := range offers {
		response.Offers[i] = dto.OfferResponse{
			ID:            offer.ID,
			Marketplace:   string(offer.Marketplace),
			StoreName:     offer.StoreName,
			Price:         offer.Price,
			LastCheckedAt: offer.LastCheckedAt,
		}
	}

	// Calculate best price
	if len(offers) > 0 {
		bestOffer := offers[0] // Already sorted by price ASC
		for _, offer := range offers {
			if offer.Price < bestOffer.Price {
				bestOffer = offer
			}
		}
		response.BestPrice = &dto.BestPrice{
			Marketplace: string(bestOffer.Marketplace),
			Price:       bestOffer.Price,
		}
	}

	return response, nil
}

// GetAllProducts gets all products with pagination
func (s *ProductService) GetAllProducts(ctx context.Context, limit, offset int) ([]*dto.ProductResponse, error) {
	// Get products from repository
	products, _, err := s.productRepo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	// Convert to response
	responses := make([]*dto.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = &dto.ProductResponse{
			ID:        product.ID,
			Title:     product.Title,
			ImageURL:  product.ImageURL,
			CreatedAt: product.CreatedAt,
		}

		// Convert offers
		if len(product.Offers) > 0 {
			responses[i].Offers = make([]dto.OfferResponse, len(product.Offers))
			for j, offer := range product.Offers {
				responses[i].Offers[j] = dto.OfferResponse{
					ID:            offer.ID,
					Marketplace:   string(offer.Marketplace),
					StoreName:     offer.StoreName,
					Price:         offer.Price,
					LastCheckedAt: offer.LastCheckedAt,
				}
			}
		}
	}

	return responses, nil
}

// DeleteProduct deletes a product and all related data
// CASCADE constraints will automatically delete:
// - Offers (ON DELETE CASCADE)
// - Links (ON DELETE CASCADE)
// - CampaignProducts (ON DELETE CASCADE)
// - Clicks (via links, ON DELETE CASCADE)
func (s *ProductService) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	// Check if product exists
	_, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	// Delete product (CASCADE will handle related data)
	if err := s.productRepo.Delete(ctx, productID); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}
