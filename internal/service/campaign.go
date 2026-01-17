package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/jonosize/affiliate-platform/internal/config"
	"github.com/jonosize/affiliate-platform/internal/dto"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/model"
)

// CampaignService handles campaign business logic
type CampaignService struct {
	campaignRepo CampaignRepositoryInterface
	linkRepo     LinkRepositoryInterface
	offerRepo    OfferRepositoryInterface
	productRepo  ProductRepositoryInterface
	logger       logger.Logger
	cfg          config.Config
}

// NewCampaignService creates a new campaign service
func NewCampaignService(
	campaignRepo CampaignRepositoryInterface,
	linkRepo LinkRepositoryInterface,
	offerRepo OfferRepositoryInterface,
	productRepo ProductRepositoryInterface,
	cfg config.Config,
	log logger.Logger,
) *CampaignService {
	return &CampaignService{
		campaignRepo: campaignRepo,
		linkRepo:     linkRepo,
		offerRepo:    offerRepo,
		productRepo:  productRepo,
		logger:       log,
		cfg:          cfg,
	}
}

// CreateCampaign creates a new campaign
func (s *CampaignService) CreateCampaign(ctx context.Context, req dto.CreateCampaignRequest) (*dto.CampaignResponse, error) {
	// Validate dates
	if req.EndAt.Before(req.StartAt) || req.EndAt.Equal(req.StartAt) {
		return nil, fmt.Errorf("end_at must be after start_at")
	}

	// Validate UTM campaign format (alphanumeric + underscore, max 100 chars)
	if len(req.UTMCampaign) > 100 {
		return nil, fmt.Errorf("utm_campaign must be 100 characters or less")
	}

	// Create campaign
	campaign := &model.Campaign{
		Name:        req.Name,
		UTMCampaign: req.UTMCampaign,
		StartAt:     req.StartAt,
		EndAt:       req.EndAt,
	}

	if err := s.campaignRepo.Create(ctx, campaign); err != nil {
		return nil, fmt.Errorf("failed to create campaign: %w", err)
	}

	// Add products to campaign (optional - can be added later)
	if len(req.ProductIDs) > 0 {
		if err := s.campaignRepo.AddProducts(ctx, campaign.ID, req.ProductIDs); err != nil {
			// If adding products fails, attempt to delete the campaign to prevent orphaned data
			s.logger.Error("Failed to add products to campaign, attempting to rollback campaign creation", logger.Error(err), logger.String("campaign_id", campaign.ID.String()))
			if deleteErr := s.campaignRepo.Delete(ctx, campaign.ID); deleteErr != nil {
				s.logger.Error("Failed to rollback campaign after product add failure", logger.Error(deleteErr), logger.String("campaign_id", campaign.ID.String()))
			}
			return nil, fmt.Errorf("failed to add products to campaign: %w", err)
		}
		// Automatically create links for products
		if err := s.createLinksForProducts(ctx, campaign.ID, req.ProductIDs); err != nil {
			s.logger.Warn("Failed to create links for products", logger.Error(err), logger.String("campaign_id", campaign.ID.String()))
			// Don't fail campaign creation if link creation fails
		}
	}

	// Convert to response
	response := &dto.CampaignResponse{
		ID:          campaign.ID,
		Name:        campaign.Name,
		UTMCampaign: campaign.UTMCampaign,
		StartAt:     campaign.StartAt,
		EndAt:       campaign.EndAt,
		CreatedAt:   campaign.CreatedAt,
	}

	return response, nil
}

// GetCampaign gets a campaign by ID
func (s *CampaignService) GetCampaign(ctx context.Context, id uuid.UUID) (*model.Campaign, error) {
	return s.campaignRepo.FindByID(ctx, id)
}

// GetCampaignResponse gets a campaign by ID and returns as DTO
func (s *CampaignService) GetCampaignResponse(ctx context.Context, id uuid.UUID) (*dto.CampaignResponse, error) {
	campaign, err := s.campaignRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("campaign not found: %w", err)
	}

	// Extract product IDs from campaign products
	productIDs := make([]uuid.UUID, 0, len(campaign.CampaignProducts))
	for _, cp := range campaign.CampaignProducts {
		if cp.ProductID != uuid.Nil {
			productIDs = append(productIDs, cp.ProductID)
		}
	}

	response := &dto.CampaignResponse{
		ID:          campaign.ID,
		Name:        campaign.Name,
		UTMCampaign: campaign.UTMCampaign,
		StartAt:     campaign.StartAt,
		EndAt:       campaign.EndAt,
		CreatedAt:   campaign.CreatedAt,
		ProductIDs:  productIDs,
	}

	return response, nil
}

// GetAllCampaigns gets all campaigns with pagination
func (s *CampaignService) GetAllCampaigns(ctx context.Context, limit, offset int) ([]*dto.CampaignResponse, error) {
	// Get campaigns from repository
	campaigns, _, err := s.campaignRepo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaigns: %w", err)
	}

	// Convert to response
	responses := make([]*dto.CampaignResponse, len(campaigns))
	for i, campaign := range campaigns {
		responses[i] = &dto.CampaignResponse{
			ID:          campaign.ID,
			Name:        campaign.Name,
			UTMCampaign: campaign.UTMCampaign,
			StartAt:     campaign.StartAt,
			EndAt:       campaign.EndAt,
			CreatedAt:   campaign.CreatedAt,
		}
	}

	return responses, nil
}

// DeleteCampaign deletes a campaign and all related data
// CASCADE constraints will automatically delete:
// - CampaignProducts (ON DELETE CASCADE)
// - Links (ON DELETE CASCADE)
// - Clicks (via links, ON DELETE CASCADE)
func (s *CampaignService) DeleteCampaign(ctx context.Context, campaignID uuid.UUID) error {
	// Check if campaign exists
	_, err := s.campaignRepo.FindByID(ctx, campaignID)
	if err != nil {
		return fmt.Errorf("campaign not found: %w", err)
	}

	// Delete campaign (CASCADE will handle related data)
	if err := s.campaignRepo.Delete(ctx, campaignID); err != nil {
		return fmt.Errorf("failed to delete campaign: %w", err)
	}

	return nil
}

// UpdateCampaign updates a campaign
func (s *CampaignService) UpdateCampaign(ctx context.Context, campaignID uuid.UUID, req dto.UpdateCampaignRequest) (*dto.CampaignResponse, error) {
	// Check if campaign exists
	campaign, err := s.campaignRepo.FindByID(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("campaign not found: %w", err)
	}

	// Update campaign fields if provided
	if req.Name != "" {
		campaign.Name = req.Name
	}
	if req.UTMCampaign != "" {
		// Validate UTM campaign format (alphanumeric + underscore, max 100 chars)
		if len(req.UTMCampaign) > 100 {
			return nil, fmt.Errorf("utm_campaign must be 100 characters or less")
		}
		campaign.UTMCampaign = req.UTMCampaign
	}
	if req.StartAt != nil {
		campaign.StartAt = *req.StartAt
	}
	if req.EndAt != nil {
		campaign.EndAt = *req.EndAt
	}

	// Validate dates
	if campaign.EndAt.Before(campaign.StartAt) || campaign.EndAt.Equal(campaign.StartAt) {
		return nil, fmt.Errorf("end_at must be after start_at")
	}

	// Update campaign
	if err := s.campaignRepo.Update(ctx, campaign); err != nil {
		return nil, fmt.Errorf("failed to update campaign: %w", err)
	}

	// Get current product IDs for link sync (needed if UTM campaign changes)
	var productIDsToSync []uuid.UUID

	// Update products if provided
	if req.ProductIDs != nil {
		s.logger.Info("Updating campaign products", logger.String("campaign_id", campaignID.String()), logger.Int("new_product_count", len(req.ProductIDs)))
		if err := s.campaignRepo.UpdateCampaignProducts(ctx, campaignID, req.ProductIDs); err != nil {
			return nil, fmt.Errorf("failed to update campaign products: %w", err)
		}
		productIDsToSync = req.ProductIDs
		s.logger.Info("Campaign products updated, starting link synchronization", logger.String("campaign_id", campaignID.String()))
	} else if req.UTMCampaign != "" {
		// If only UTM campaign changed, get current products to update their links
		currentCampaign, err := s.campaignRepo.FindByID(ctx, campaignID)
		if err == nil {
			productIDsToSync = make([]uuid.UUID, 0, len(currentCampaign.CampaignProducts))
			for _, cp := range currentCampaign.CampaignProducts {
				if cp.ProductID != uuid.Nil {
					productIDsToSync = append(productIDsToSync, cp.ProductID)
				}
			}
		}
	}

	// Sync links if products changed or UTM campaign changed
	// This ensures target URLs are updated when UTM campaign changes
	// Also sync if products were set to empty list (to delete all links)
	if req.ProductIDs != nil || req.UTMCampaign != "" {
		// Automatically sync links for products (add new, remove unused, update URLs)
		// This will also handle the case where productIDsToSync is empty (delete all links)
		if err := s.createLinksForProducts(ctx, campaignID, productIDsToSync); err != nil {
			s.logger.Warn("Failed to sync links for products", logger.Error(err), logger.String("campaign_id", campaignID.String()))
			// Don't fail campaign update if link creation fails
		}
	}

	// Get updated campaign with products
	updatedCampaign, err := s.campaignRepo.FindByID(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated campaign: %w", err)
	}

	// Convert to response
	response := &dto.CampaignResponse{
		ID:          updatedCampaign.ID,
		Name:        updatedCampaign.Name,
		UTMCampaign: updatedCampaign.UTMCampaign,
		StartAt:     updatedCampaign.StartAt,
		EndAt:       updatedCampaign.EndAt,
		CreatedAt:   updatedCampaign.CreatedAt,
	}

	return response, nil
}

// UpdateCampaignProducts updates the products in a campaign
// This replaces all existing products with the new list
func (s *CampaignService) UpdateCampaignProducts(ctx context.Context, campaignID uuid.UUID, productIDs []uuid.UUID) error {
	// Check if campaign exists
	_, err := s.campaignRepo.FindByID(ctx, campaignID)
	if err != nil {
		return fmt.Errorf("campaign not found: %w", err)
	}

	// Update products
	if err := s.campaignRepo.UpdateCampaignProducts(ctx, campaignID, productIDs); err != nil {
		return fmt.Errorf("failed to update campaign products: %w", err)
	}

	// Automatically create links for products
	if err := s.createLinksForProducts(ctx, campaignID, productIDs); err != nil {
		s.logger.Warn("Failed to create links for products", logger.Error(err), logger.String("campaign_id", campaignID.String()))
		// Don't fail update if link creation fails
	}

	return nil
}

// createLinksForProducts creates affiliate links for products in a campaign
// It synchronizes links: removes unused links and creates missing links
func (s *CampaignService) createLinksForProducts(ctx context.Context, campaignID uuid.UUID, productIDs []uuid.UUID) error {
	s.logger.Info("Starting link synchronization for campaign", logger.String("campaign_id", campaignID.String()), logger.Int("product_count", len(productIDs)))

	// Get campaign to get UTM campaign name
	campaign, err := s.campaignRepo.FindByID(ctx, campaignID)
	if err != nil {
		return fmt.Errorf("failed to get campaign: %w", err)
	}

	// Step 1: Remove links for products that are no longer in the campaign
	if err := s.linkRepo.DeleteByCampaignIDAndNotInProducts(ctx, campaignID, productIDs); err != nil {
		s.logger.Warn("Failed to remove unused links", logger.Error(err), logger.String("campaign_id", campaignID.String()))
		// Continue anyway - we'll still create missing links
	} else {
		s.logger.Info("Removed links for products no longer in campaign", logger.String("campaign_id", campaignID.String()))
	}

	// Step 2: For each product, sync links (remove unused, create missing)
	for _, productID := range productIDs {
		s.logger.Info("Processing product for link sync", logger.String("product_id", productID.String()), logger.String("campaign_id", campaignID.String()))

		// Get offers for this product
		offers, err := s.offerRepo.FindByProductID(ctx, productID)
		if err != nil {
			s.logger.Warn("Failed to get offers for product", logger.Error(err), logger.String("product_id", productID.String()))
			continue
		}

		s.logger.Info("Product offers retrieved", logger.String("product_id", productID.String()), logger.Int("offers_count", len(offers)))

		// Get existing links for this product in the campaign
		existingLinks, err := s.linkRepo.FindByProductIDAndCampaignID(ctx, productID, campaignID)
		if err != nil {
			s.logger.Warn("Failed to get existing links", logger.Error(err), logger.String("product_id", productID.String()))
			continue
		}

		s.logger.Info("Existing links retrieved", logger.String("product_id", productID.String()), logger.Int("existing_links_count", len(existingLinks)))

		// Build a map of marketplaces that have offers
		offerMarketplaces := make(map[model.Marketplace]bool)
		for _, offer := range offers {
			offerMarketplaces[offer.Marketplace] = true
		}

		// Remove links for marketplaces that no longer have offers
		for _, existingLink := range existingLinks {
			if !offerMarketplaces[existingLink.Marketplace] {
				s.logger.Info("Removing link for marketplace without offer", logger.String("product_id", productID.String()), logger.String("marketplace", string(existingLink.Marketplace)), logger.String("link_id", existingLink.ID.String()))
				if err := s.linkRepo.Delete(ctx, existingLink.ID); err != nil {
					s.logger.Warn("Failed to delete unused link", logger.Error(err), logger.String("link_id", existingLink.ID.String()))
				}
			}
		}

		// Re-fetch links after deletion to get current state
		existingLinks, err = s.linkRepo.FindByProductIDAndCampaignID(ctx, productID, campaignID)
		if err != nil {
			s.logger.Warn("Failed to re-fetch links after deletion", logger.Error(err), logger.String("product_id", productID.String()))
			existingLinks = []*model.Link{} // Use empty slice if fetch fails
		}

		// Build a map of existing link marketplaces (after deletion)
		existingLinkMarketplacesMap := make(map[model.Marketplace]bool)
		for _, existingLink := range existingLinks {
			existingLinkMarketplacesMap[existingLink.Marketplace] = true
		}

		// Update existing links' target URLs if UTM campaign changed, or create new links
		for _, offer := range offers {
			if existingLinkMarketplacesMap[offer.Marketplace] {
				// Link exists - check if we need to update target URL
				var existingLink *model.Link
				for _, link := range existingLinks {
					if link.Marketplace == offer.Marketplace {
						existingLink = link
						break
					}
				}

				if existingLink != nil {
					// Build new target URL with current UTM campaign
					baseURL := offer.MarketplaceProductURL
					utmSource := "affiliate"
					utmMedium := "affiliate"
					utmCampaign := campaign.UTMCampaign

					newTargetURL, err := buildTargetURL(baseURL, utmCampaign, utmSource, utmMedium)
					if err != nil {
						s.logger.Warn("Failed to build target URL for update", logger.Error(err))
						continue
					}

					// Update if target URL changed
					if existingLink.TargetURL != newTargetURL {
						existingLink.TargetURL = newTargetURL
						if err := s.linkRepo.Update(ctx, existingLink); err != nil {
							s.logger.Warn("Failed to update link target URL", logger.Error(err), logger.String("link_id", existingLink.ID.String()))
							continue
						}
						s.logger.Info("Updated existing link target URL", logger.String("product_id", productID.String()), logger.String("marketplace", string(offer.Marketplace)), logger.String("link_id", existingLink.ID.String()))
					} else {
						s.logger.Info("Link already exists, skipping", logger.String("product_id", productID.String()), logger.String("marketplace", string(offer.Marketplace)))
					}
				}
				continue // Link already exists (or was updated)
			}

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
				ProductID:   productID,
				CampaignID:  campaignID,
				Marketplace: offer.Marketplace,
				ShortCode:   shortCode,
				TargetURL:   targetURL,
			}

			if err := s.linkRepo.Create(ctx, link); err != nil {
				s.logger.Warn("Failed to create link", logger.Error(err), logger.String("product_id", productID.String()), logger.String("marketplace", string(offer.Marketplace)))
				continue
			}

			s.logger.Info("Created new link", logger.String("product_id", productID.String()), logger.String("marketplace", string(offer.Marketplace)), logger.String("short_code", shortCode))
		}
	}

	s.logger.Info("Link synchronization completed", logger.String("campaign_id", campaignID.String()))

	return nil
}

// generateUniqueShortCode generates a unique short code, retrying on collision
func (s *CampaignService) generateUniqueShortCode(ctx context.Context) (string, error) {
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
