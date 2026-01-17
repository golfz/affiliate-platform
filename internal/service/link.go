package service

import (
	"context"
	"fmt"

	"github.com/jonosize/affiliate-platform/internal/config"
	"github.com/jonosize/affiliate-platform/internal/database"
	"github.com/jonosize/affiliate-platform/internal/dto"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/model"
	"github.com/jonosize/affiliate-platform/internal/repository"
)

// LinkService handles link business logic
type LinkService struct {
	linkRepo     *repository.LinkRepository
	campaignRepo *repository.CampaignRepository
	productRepo  *repository.ProductRepository
	offerRepo    *repository.OfferRepository
	db           *database.DB
	logger       logger.Logger
	cfg          config.Config
}

// NewLinkService creates a new link service
func NewLinkService(db *database.DB, cfg config.Config, log logger.Logger) *LinkService {
	return &LinkService{
		linkRepo:     repository.NewLinkRepository(db),
		campaignRepo: repository.NewCampaignRepository(db),
		productRepo:  repository.NewProductRepository(db),
		offerRepo:    repository.NewOfferRepository(db),
		db:           db,
		logger:       log,
		cfg:          cfg,
	}
}

// generateUniqueShortCode generates a unique short code, retrying on collision
func (s *LinkService) generateUniqueShortCode(ctx context.Context) (string, error) {
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		code, err := generateShortCode()
		if err != nil {
			return "", err
		}

		// Check uniqueness
		exists, err := s.linkRepo.ShortCodeExists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique short code after %d retries", maxRetries)
}

// CreateLink creates an affiliate link
func (s *LinkService) CreateLink(ctx context.Context, req dto.CreateLinkRequest) (*dto.LinkResponse, error) {
	// Validate marketplace
	marketplace := model.Marketplace(req.Marketplace)
	if marketplace != model.MarketplaceLazada && marketplace != model.MarketplaceShopee {
		return nil, fmt.Errorf("invalid marketplace: must be 'lazada' or 'shopee'")
	}

	// Verify product exists
	if _, err := s.productRepo.FindByID(ctx, req.ProductID); err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Verify campaign exists
	campaign, err := s.campaignRepo.FindByID(ctx, req.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("campaign not found: %w", err)
	}

	// Get offer for the product and marketplace
	offer, err := s.offerRepo.FindByProductIDAndMarketplace(ctx, req.ProductID, marketplace)
	if err != nil {
		return nil, fmt.Errorf("offer not found for product and marketplace: %w", err)
	}

	// Verify the offer's marketplace matches the requested marketplace
	if offer.Marketplace != marketplace {
		return nil, fmt.Errorf("offer marketplace mismatch: expected %s, got %s", marketplace, offer.Marketplace)
	}

	// Generate unique short code
	shortCode, err := s.generateUniqueShortCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate short code: %w", err)
	}

	// Build target URL with UTM parameters
	baseURL := offer.MarketplaceProductURL
	utmSource := "affiliate"
	utmMedium := "affiliate"
	utmCampaign := campaign.UTMCampaign

	targetURL, err := buildTargetURL(baseURL, utmCampaign, utmSource, utmMedium)
	if err != nil {
		return nil, fmt.Errorf("failed to build target URL: %w", err)
	}

	// Create link
	link := &model.Link{
		ProductID:   req.ProductID,
		CampaignID:  req.CampaignID,
		Marketplace: marketplace,
		ShortCode:   shortCode,
		TargetURL:   targetURL,
	}

	if err := s.linkRepo.Create(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to create link: %w", err)
	}

	// Build full URL
	apiBaseURL := s.cfg.GetAPIBaseURL()
	if apiBaseURL == "" {
		apiBaseURL = "http://localhost:8080"
	}
	fullURL := fmt.Sprintf("%s/go/%s", apiBaseURL, shortCode)

	// Convert to response
	response := &dto.LinkResponse{
		ID:        link.ID,
		ShortCode: link.ShortCode,
		TargetURL: link.TargetURL,
		FullURL:   fullURL,
	}

	return response, nil
}
