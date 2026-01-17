package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/jonosize/affiliate-platform/internal/database"
	"github.com/jonosize/affiliate-platform/internal/model"
)

// CampaignRepository handles campaign database operations
type CampaignRepository struct {
	db *database.DB
}

// NewCampaignRepository creates a new campaign repository
func NewCampaignRepository(db *database.DB) *CampaignRepository {
	return &CampaignRepository{db: db}
}

// Create creates a new campaign (uses write DB)
func (r *CampaignRepository) Create(ctx context.Context, campaign *model.Campaign) error {
	return r.db.Write.WithContext(ctx).Create(campaign).Error
}

// FindByID finds a campaign by ID (uses read DB)
func (r *CampaignRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Campaign, error) {
	var campaign model.Campaign
	err := r.db.Read.WithContext(ctx).
		Preload("CampaignProducts.Product").
		Preload("Links").
		First(&campaign, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &campaign, nil
}

// FindAll finds all campaigns (uses read DB)
func (r *CampaignRepository) FindAll(ctx context.Context, limit, offset int) ([]*model.Campaign, int64, error) {
	var campaigns []*model.Campaign
	var total int64

	// Count total
	if err := r.db.Read.WithContext(ctx).Model(&model.Campaign{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Find with pagination
	err := r.db.Read.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&campaigns).Error

	if err != nil {
		return nil, 0, err
	}

	return campaigns, total, nil
}

// Update updates a campaign (uses write DB)
func (r *CampaignRepository) Update(ctx context.Context, campaign *model.Campaign) error {
	return r.db.Write.WithContext(ctx).Save(campaign).Error
}

// Delete deletes a campaign (uses write DB)
func (r *CampaignRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.Write.WithContext(ctx).Delete(&model.Campaign{}, "id = ?", id).Error
}

// AddProducts adds products to a campaign (uses write DB)
func (r *CampaignRepository) AddProducts(ctx context.Context, campaignID uuid.UUID, productIDs []uuid.UUID) error {
	for _, productID := range productIDs {
		campaignProduct := &model.CampaignProduct{
			CampaignID: campaignID,
			ProductID:  productID,
		}
		if err := r.db.Write.WithContext(ctx).Create(campaignProduct).Error; err != nil {
			// Ignore duplicate errors (UNIQUE constraint)
			continue
		}
	}
	return nil
}

// RemoveProducts removes products from a campaign (uses write DB)
func (r *CampaignRepository) RemoveProducts(ctx context.Context, campaignID uuid.UUID, productIDs []uuid.UUID) error {
	return r.db.Write.WithContext(ctx).
		Where("campaign_id = ? AND product_id IN ?", campaignID, productIDs).
		Delete(&model.CampaignProduct{}).Error
}

// UpdateCampaignProducts replaces all products in a campaign (uses write DB)
// This removes all existing products and adds the new ones
func (r *CampaignRepository) UpdateCampaignProducts(ctx context.Context, campaignID uuid.UUID, productIDs []uuid.UUID) error {
	// Remove all existing products
	if err := r.db.Write.WithContext(ctx).
		Where("campaign_id = ?", campaignID).
		Delete(&model.CampaignProduct{}).Error; err != nil {
		return err
	}

	// Add new products
	if len(productIDs) > 0 {
		return r.AddProducts(ctx, campaignID, productIDs)
	}

	return nil
}
