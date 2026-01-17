package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jonosize/affiliate-platform/internal/config"
	"github.com/jonosize/affiliate-platform/internal/database"
	"github.com/jonosize/affiliate-platform/internal/dto"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/model"
	"github.com/jonosize/affiliate-platform/internal/repository"
)

// CampaignService handles campaign business logic
type CampaignService struct {
	campaignRepo *repository.CampaignRepository
	db           *database.DB
	logger       logger.Logger
	cfg          config.Config
}

// NewCampaignService creates a new campaign service
func NewCampaignService(db *database.DB, cfg config.Config, log logger.Logger) *CampaignService {
	return &CampaignService{
		campaignRepo: repository.NewCampaignRepository(db),
		db:           db,
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

	// Validate product IDs (required)
	if len(req.ProductIDs) == 0 {
		return nil, fmt.Errorf("at least one product is required")
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

	// Add products to campaign (required)
	if err := s.campaignRepo.AddProducts(ctx, campaign.ID, req.ProductIDs); err != nil {
		return nil, fmt.Errorf("failed to add products to campaign: %w", err)
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
