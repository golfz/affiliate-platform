package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/jonosize/affiliate-platform/internal/database"
	"github.com/jonosize/affiliate-platform/internal/model"
)

// ClickRepository handles click database operations
type ClickRepository struct {
	db *database.DB
}

// NewClickRepository creates a new click repository
func NewClickRepository(db *database.DB) *ClickRepository {
	return &ClickRepository{db: db}
}

// Create creates a new click event (uses write DB)
func (r *ClickRepository) Create(ctx context.Context, click *model.Click) error {
	return r.db.Write.WithContext(ctx).Create(click).Error
}

// CountByLinkID counts clicks for a link (uses read DB)
func (r *ClickRepository) CountByLinkID(ctx context.Context, linkID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Read.WithContext(ctx).
		Model(&model.Click{}).
		Where("link_id = ?", linkID).
		Count(&count).Error
	return count, err
}

// CountByLinkIDAndTimeRange counts clicks for a link within a time range (uses read DB)
func (r *ClickRepository) CountByLinkIDAndTimeRange(ctx context.Context, linkID uuid.UUID, startAt, endAt time.Time) (int64, error) {
	var count int64
	err := r.db.Read.WithContext(ctx).
		Model(&model.Click{}).
		Where("link_id = ? AND timestamp >= ? AND timestamp <= ?", linkID, startAt, endAt).
		Count(&count).Error
	return count, err
}

// FindByCampaignID counts clicks for a campaign (uses read DB)
func (r *ClickRepository) CountByCampaignID(ctx context.Context, campaignID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Read.WithContext(ctx).
		Model(&model.Click{}).
		Joins("JOIN links ON clicks.link_id = links.id").
		Where("links.campaign_id = ?", campaignID).
		Count(&count).Error
	return count, err
}

// FindRecentClicks finds the most recent clicks with related data (uses read DB)
func (r *ClickRepository) FindRecentClicks(ctx context.Context, limit int) ([]model.Click, error) {
	var clicks []model.Click
	err := r.db.Read.WithContext(ctx).
		Preload("Link").
		Preload("Link.Product").
		Preload("Link.Campaign").
		Order("timestamp DESC").
		Limit(limit).
		Find(&clicks).Error
	return clicks, err
}

// CountWithFilters counts clicks with optional filters (uses read DB)
func (r *ClickRepository) CountWithFilters(ctx context.Context, campaignID *uuid.UUID, marketplace *string, startDate, endDate time.Time) (int64, error) {
	query := r.db.Read.WithContext(ctx).Model(&model.Click{})

	// Apply date range filter
	if !startDate.IsZero() {
		query = query.Where("timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("timestamp <= ?", endDate)
	}

	// Apply campaign filter (via link join)
	if campaignID != nil {
		query = query.Joins("JOIN links ON clicks.link_id = links.id").
			Where("links.campaign_id = ?", *campaignID)
	}

	// Apply marketplace filter (via link join)
	if marketplace != nil {
		query = query.Joins("JOIN links ON clicks.link_id = links.id").
			Where("links.marketplace = ?", *marketplace)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// CountByCampaignWithFilters counts clicks grouped by campaign with filters (uses read DB)
func (r *ClickRepository) CountByCampaignWithFilters(ctx context.Context, campaignID *uuid.UUID, marketplace *string, startDate, endDate time.Time) ([]model.CampaignStatResult, error) {
	query := r.db.Read.WithContext(ctx).
		Table("clicks").
		Select("campaigns.id as campaign_id, campaigns.name as campaign_name, COUNT(clicks.id) as clicks").
		Joins("JOIN links ON clicks.link_id = links.id").
		Joins("JOIN campaigns ON links.campaign_id = campaigns.id").
		Group("campaigns.id, campaigns.name")

	// Apply date range filter
	if !startDate.IsZero() {
		query = query.Where("clicks.timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("clicks.timestamp <= ?", endDate)
	}

	// Apply marketplace filter
	if marketplace != nil {
		query = query.Where("links.marketplace = ?", *marketplace)
	}

	// Apply campaign filter
	if campaignID != nil {
		query = query.Where("campaigns.id = ?", *campaignID)
	}

	var results []model.CampaignStatResult
	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

// CountByMarketplaceWithFilters counts clicks grouped by marketplace with filters (uses read DB)
func (r *ClickRepository) CountByMarketplaceWithFilters(ctx context.Context, campaignID *uuid.UUID, marketplace *string, startDate, endDate time.Time) ([]model.MarketplaceStatResult, error) {
	query := r.db.Read.WithContext(ctx).
		Table("clicks").
		Select("links.marketplace, COUNT(clicks.id) as clicks").
		Joins("JOIN links ON clicks.link_id = links.id").
		Group("links.marketplace")

	// Apply date range filter
	if !startDate.IsZero() {
		query = query.Where("clicks.timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("clicks.timestamp <= ?", endDate)
	}

	// Apply campaign filter
	if campaignID != nil {
		query = query.Where("links.campaign_id = ?", *campaignID)
	}

	// Apply marketplace filter
	if marketplace != nil {
		query = query.Where("links.marketplace = ?", *marketplace)
	}

	var results []model.MarketplaceStatResult
	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

// FindTopProductsWithFilters finds top products by click count with filters (uses read DB)
func (r *ClickRepository) FindTopProductsWithFilters(ctx context.Context, campaignID *uuid.UUID, marketplace *string, startDate, endDate time.Time, limit int) ([]model.TopProductResult, error) {
	query := r.db.Read.WithContext(ctx).
		Table("clicks").
		Select("products.id as product_id, products.title as product_name, links.marketplace, COUNT(clicks.id) as clicks").
		Joins("JOIN links ON clicks.link_id = links.id").
		Joins("JOIN products ON links.product_id = products.id").
		Group("products.id, products.title, links.marketplace").
		Order("clicks DESC").
		Limit(limit)

	// Apply date range filter
	if !startDate.IsZero() {
		query = query.Where("clicks.timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("clicks.timestamp <= ?", endDate)
	}

	// Apply campaign filter
	if campaignID != nil {
		query = query.Where("links.campaign_id = ?", *campaignID)
	}

	// Apply marketplace filter
	if marketplace != nil {
		query = query.Where("links.marketplace = ?", *marketplace)
	}

	var results []model.TopProductResult
	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}
