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

// CreateProduct creates a product from URL or SKU
func (s *ProductService) CreateProduct(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error) {
	// Determine source type
	sourceType := adapters.SourceType(req.SourceType)
	if sourceType != adapters.SourceTypeURL && sourceType != adapters.SourceTypeSKU {
		sourceType = adapters.SourceTypeURL // Default to URL
	}

	// Validate URL if source type is URL
	var marketplace adapters.Marketplace
	if sourceType == adapters.SourceTypeURL {
		var err error
		marketplace, sourceType, err = validator.ValidateProductURL(req.Source)
		if err != nil {
			return nil, fmt.Errorf("invalid product URL: %w", err)
		}
	}

	// Get adapter (using mock adapter for now)
	lazadaAdapter, shopeeAdapter, err := mock.GetMockAdapters()
	if err != nil {
		return nil, fmt.Errorf("failed to get adapters: %w", err)
	}

	// Fetch product data from adapter
	var adapter adapters.MarketplaceAdapter
	var productData *adapters.ProductData

	// Try to determine marketplace from source if not already determined
	if marketplace == "" {
		// Try both adapters
		productData, err = lazadaAdapter.FetchProduct(ctx, req.Source, sourceType)
		if err == nil {
			adapter = lazadaAdapter
			marketplace = adapters.MarketplaceLazada
		} else {
			productData, err = shopeeAdapter.FetchProduct(ctx, req.Source, sourceType)
			if err == nil {
				adapter = shopeeAdapter
				marketplace = adapters.MarketplaceShopee
			}
		}
	} else {
		// Use determined marketplace
		if marketplace == adapters.MarketplaceLazada {
			adapter = lazadaAdapter
		} else {
			adapter = shopeeAdapter
		}
		productData, err = adapter.FetchProduct(ctx, req.Source, sourceType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch product data: %w", err)
	}

	// Create product
	product := &model.Product{
		Title:    productData.Title,
		ImageURL: productData.ImageURL,
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

	// Fetch Lazada offer if:
	// 1. Lazada URL is explicitly provided, OR
	// 2. Primary marketplace is Lazada and no explicit URLs are provided
	if hasLazadaURL || (marketplace == adapters.MarketplaceLazada && !hasLazadaURL && !hasShopeeURL) {
		var lazadaOfferURL string
		if hasLazadaURL {
			lazadaOfferURL = req.LazadaURL
		} else {
			lazadaOfferURL = productData.MarketplaceProductURL
		}

		offerData, err := lazadaAdapter.FetchOffer(ctx, lazadaOfferURL)
		if err == nil {
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
	// 2. Primary marketplace is Shopee and no explicit URLs are provided
	if hasShopeeURL || (marketplace == adapters.MarketplaceShopee && !hasLazadaURL && !hasShopeeURL) {
		var shopeeOfferURL string
		if hasShopeeURL {
			shopeeOfferURL = req.ShopeeURL
		} else {
			shopeeOfferURL = productData.MarketplaceProductURL
		}

		offerData, err := shopeeAdapter.FetchOffer(ctx, shopeeOfferURL)
		if err == nil {
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
