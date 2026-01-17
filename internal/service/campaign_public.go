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
