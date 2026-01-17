package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jonosize/affiliate-platform/internal/database"
	"github.com/jonosize/affiliate-platform/internal/model"
)

// LinkRepository handles link database operations
type LinkRepository struct {
	db *database.DB
}

// NewLinkRepository creates a new link repository
func NewLinkRepository(db *database.DB) *LinkRepository {
	return &LinkRepository{db: db}
}

// Create creates a new link (uses write DB)
func (r *LinkRepository) Create(ctx context.Context, link *model.Link) error {
	return r.db.Write.WithContext(ctx).Create(link).Error
}

// FindByShortCode finds a link by short code (uses read DB)
func (r *LinkRepository) FindByShortCode(ctx context.Context, shortCode string) (*model.Link, error) {
	var link model.Link
	err := r.db.Read.WithContext(ctx).
		Preload("Product").
		Preload("Campaign").
		Where("short_code = ?", shortCode).
		First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// FindByProductIDAndCampaignID finds links by product and campaign (uses read DB)
func (r *LinkRepository) FindByProductIDAndCampaignID(ctx context.Context, productID, campaignID uuid.UUID) ([]*model.Link, error) {
	var links []*model.Link
	err := r.db.Read.WithContext(ctx).
		Where("product_id = ? AND campaign_id = ?", productID, campaignID).
		Find(&links).Error
	if err != nil {
		return nil, err
	}
	return links, nil
}

// ShortCodeExists checks if a short code already exists (uses read DB)
func (r *LinkRepository) ShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	var count int64
	err := r.db.Read.WithContext(ctx).
		Model(&model.Link{}).
		Where("short_code = ?", shortCode).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Update updates a link (uses write DB)
func (r *LinkRepository) Update(ctx context.Context, link *model.Link) error {
	return r.db.Write.WithContext(ctx).Save(link).Error
}

// Delete deletes a link (uses write DB)
func (r *LinkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.Write.WithContext(ctx).Delete(&model.Link{}, "id = ?", id).Error
}

// FindByCampaignID finds all links for a campaign (uses read DB)
func (r *LinkRepository) FindByCampaignID(ctx context.Context, campaignID uuid.UUID) ([]*model.Link, error) {
	var links []*model.Link
	err := r.db.Read.WithContext(ctx).
		Where("campaign_id = ?", campaignID).
		Find(&links).Error
	if err != nil {
		return nil, err
	}
	return links, nil
}

// DeleteByProductIDAndCampaignID deletes links for a specific product in a campaign (uses write DB)
func (r *LinkRepository) DeleteByProductIDAndCampaignID(ctx context.Context, productID, campaignID uuid.UUID) error {
	return r.db.Write.WithContext(ctx).
		Where("product_id = ? AND campaign_id = ?", productID, campaignID).
		Delete(&model.Link{}).Error
}

// DeleteByCampaignIDAndNotInProducts deletes links for products not in the provided list (uses write DB)
func (r *LinkRepository) DeleteByCampaignIDAndNotInProducts(ctx context.Context, campaignID uuid.UUID, productIDs []uuid.UUID) error {
	if len(productIDs) == 0 {
		// If no products, delete all links for this campaign
		return r.db.Write.WithContext(ctx).
			Where("campaign_id = ?", campaignID).
			Delete(&model.Link{}).Error
	}
	return r.db.Write.WithContext(ctx).
		Where("campaign_id = ? AND product_id NOT IN ?", campaignID, productIDs).
		Delete(&model.Link{}).Error
}

// CountWithFilters counts links with optional filters (uses read DB)
func (r *LinkRepository) CountWithFilters(ctx context.Context, campaignID *uuid.UUID, marketplace *string) (int64, error) {
	query := r.db.Read.WithContext(ctx).Model(&model.Link{})

	if campaignID != nil {
		query = query.Where("campaign_id = ?", *campaignID)
	}
	if marketplace != nil {
		query = query.Where("marketplace = ?", *marketplace)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
