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
