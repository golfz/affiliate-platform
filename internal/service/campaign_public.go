package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jonosize/affiliate-platform/internal/config"
	"github.com/jonosize/affiliate-platform/internal/database"
	"github.com/jonosize/affiliate-platform/internal/dto"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/model"
	"github.com/jonosize/affiliate-platform/internal/repository"
)

// CampaignPublicService handles public campaign business logic
type CampaignPublicService struct {
	campaignRepo *repository.CampaignRepository
	productRepo  *repository.ProductRepository
	offerRepo    *repository.OfferRepository
	linkRepo     *repository.LinkRepository
	cfg          config.Config
	db           *database.DB
	logger       logger.Logger
}

// NewCampaignPublicService creates a new public campaign service
func NewCampaignPublicService(db *database.DB, cfg config.Config, log logger.Logger) *CampaignPublicService {
	return &CampaignPublicService{
		campaignRepo: repository.NewCampaignRepository(db),
		productRepo:  repository.NewProductRepository(db),
		offerRepo:    repository.NewOfferRepository(db),
		linkRepo:     repository.NewLinkRepository(db),
		cfg:          cfg,
		db:           db,
		logger:       log,
	}
}

// GetPublicCampaign gets a public campaign view with products and offers
func (s *CampaignPublicService) GetPublicCampaign(ctx context.Context, campaignID uuid.UUID) (*dto.CampaignPublicResponse, error) {
	// Get campaign
	campaign, err := s.campaignRepo.FindByID(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("campaign not found: %w", err)
	}

	// Check if campaign is active
	// Use UTC for comparison to match database timezone
	now := time.Now().UTC()
	if now.Before(campaign.StartAt) || now.After(campaign.EndAt) {
		return nil, fmt.Errorf("campaign is not active")
	}

	// Build response
	response := &dto.CampaignPublicResponse{
		ID:       campaign.ID,
		Name:     campaign.Name,
		StartAt:  campaign.StartAt,
		EndAt:    campaign.EndAt,
		Products: make([]dto.CampaignProduct, 0),
	}

	// Get products for campaign
	for _, cp := range campaign.CampaignProducts {
		if cp.Product.ID == uuid.Nil {
			continue
		}

		product := cp.Product

		// Get offers for product
		offers, err := s.offerRepo.FindByProductID(ctx, product.ID)
		if err != nil {
			s.logger.Error("Failed to get offers for product", logger.Error(err), logger.String("product_id", product.ID.String()))
			continue
		}

		// Convert offers to DTO
		offerResponses := make([]dto.OfferResponse, 0, len(offers))
		var bestPrice *dto.BestPrice

		if len(offers) > 0 {
			// Sort by price and find best
			bestOffer := offers[0]
			for _, offer := range offers {
				if offer.Price < bestOffer.Price {
					bestOffer = offer
				}
				offerResponses = append(offerResponses, dto.OfferResponse{
					ID:            offer.ID,
					Marketplace:   string(offer.Marketplace),
					StoreName:     offer.StoreName,
					Price:         offer.Price,
					LastCheckedAt: offer.LastCheckedAt,
				})
			}

			bestPrice = &dto.BestPrice{
				Marketplace: string(bestOffer.Marketplace),
				Price:       bestOffer.Price,
			}
		}

		// Get links for this product in the campaign
		links, err := s.linkRepo.FindByProductIDAndCampaignID(ctx, product.ID, campaign.ID)
		if err != nil {
			s.logger.Error("Failed to get links", logger.Error(err))
		}

		// If no links exist, create them automatically
		if len(links) == 0 && len(offers) > 0 {
			s.logger.Info("No links found, creating links automatically", logger.String("product_id", product.ID.String()), logger.String("campaign_id", campaign.ID.String()))
			// Create links for each marketplace that has an offer
			for _, offer := range offers {
				// Generate unique short code
				shortCode, err := s.generateUniqueShortCode(ctx)
				if err != nil {
					s.logger.Warn("Failed to generate short code", logger.Error(err))
					continue
				}

				// Build target URL with UTM parameters
				baseURL := offer.MarketplaceProductURL
				utmSource := "affiliate"
				utmMedium := "affiliate"
				utmCampaign := campaign.UTMCampaign

				targetURL, err := buildTargetURL(baseURL, utmCampaign, utmSource, utmMedium)
				if err != nil {
					s.logger.Warn("Failed to build target URL", logger.Error(err))
					continue
				}

				// Create link
				link := &model.Link{
					ProductID:   product.ID,
					CampaignID:  campaign.ID,
					Marketplace: offer.Marketplace,
					ShortCode:   shortCode,
					TargetURL:   targetURL,
				}

				if err := s.linkRepo.Create(ctx, link); err != nil {
					s.logger.Warn("Failed to create link", logger.Error(err))
					continue
				}
				links = append(links, link)
			}
			// Re-fetch links to get all created links
			if len(links) > 0 {
				links, err = s.linkRepo.FindByProductIDAndCampaignID(ctx, product.ID, campaign.ID)
				if err != nil {
					s.logger.Warn("Failed to re-fetch links after creation", logger.Error(err))
				}
			}
		}

		// Convert links to DTO
		productLinks := make([]dto.ProductLink, len(links))
		apiBaseURL := s.cfg.GetAPIBaseURL()
		for i, link := range links {
			fullURL := apiBaseURL + "/go/" + link.ShortCode
			productLinks[i] = dto.ProductLink{
				Marketplace: string(link.Marketplace),
				ShortCode:   link.ShortCode,
				FullURL:     fullURL,
			}
		}

		response.Products = append(response.Products, dto.CampaignProduct{
			ID:        product.ID,
			Title:     product.Title,
			ImageURL:  product.ImageURL,
			Offers:    offerResponses,
			BestPrice: bestPrice,
			Links:     productLinks,
		})
	}

	return response, nil
}

// generateUniqueShortCode generates a unique short code, retrying on collision
func (s *CampaignPublicService) generateUniqueShortCode(ctx context.Context) (string, error) {
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
