package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/jonosize/affiliate-platform/internal/model"
)

// ProductRepositoryInterface defines the interface for product repository operations
type ProductRepositoryInterface interface {
	Create(ctx context.Context, product *model.Product) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Product, error)
	FindAll(ctx context.Context, limit, offset int) ([]*model.Product, int64, error)
	Update(ctx context.Context, product *model.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// OfferRepositoryInterface defines the interface for offer repository operations
type OfferRepositoryInterface interface {
	Create(ctx context.Context, offer *model.Offer) error
	FindByProductID(ctx context.Context, productID uuid.UUID) ([]*model.Offer, error)
	FindByProductIDAndMarketplace(ctx context.Context, productID uuid.UUID, marketplace model.Marketplace) (*model.Offer, error)
	Update(ctx context.Context, offer *model.Offer) error
	Upsert(ctx context.Context, offer *model.Offer) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// CampaignRepositoryInterface defines the interface for campaign repository operations
type CampaignRepositoryInterface interface {
	Create(ctx context.Context, campaign *model.Campaign) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Campaign, error)
	FindAll(ctx context.Context, limit, offset int) ([]*model.Campaign, int64, error)
	Update(ctx context.Context, campaign *model.Campaign) error
	Delete(ctx context.Context, id uuid.UUID) error
	AddProducts(ctx context.Context, campaignID uuid.UUID, productIDs []uuid.UUID) error
	RemoveProducts(ctx context.Context, campaignID uuid.UUID, productIDs []uuid.UUID) error
	UpdateCampaignProducts(ctx context.Context, campaignID uuid.UUID, productIDs []uuid.UUID) error
}

// LinkRepositoryInterface defines the interface for link repository operations
type LinkRepositoryInterface interface {
	Create(ctx context.Context, link *model.Link) error
	FindByShortCode(ctx context.Context, shortCode string) (*model.Link, error)
	FindByProductIDAndCampaignID(ctx context.Context, productID, campaignID uuid.UUID) ([]*model.Link, error)
	FindByCampaignID(ctx context.Context, campaignID uuid.UUID) ([]*model.Link, error)
	ShortCodeExists(ctx context.Context, shortCode string) (bool, error)
	Update(ctx context.Context, link *model.Link) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByProductIDAndCampaignID(ctx context.Context, productID, campaignID uuid.UUID) error
	DeleteByCampaignIDAndNotInProducts(ctx context.Context, campaignID uuid.UUID, productIDs []uuid.UUID) error
	CountWithFilters(ctx context.Context, campaignID *uuid.UUID, marketplace *string) (int64, error)
}

// ClickRepositoryInterface defines the interface for click repository operations
type ClickRepositoryInterface interface {
	Create(ctx context.Context, click *model.Click) error
	CountByLinkID(ctx context.Context, linkID uuid.UUID) (int64, error)
	CountByLinkIDAndTimeRange(ctx context.Context, linkID uuid.UUID, startAt, endAt time.Time) (int64, error)
	CountByCampaignID(ctx context.Context, campaignID uuid.UUID) (int64, error)
	FindRecentClicks(ctx context.Context, limit int) ([]model.Click, error)
	CountWithFilters(ctx context.Context, campaignID *uuid.UUID, marketplace *string, startDate, endDate time.Time) (int64, error)
	CountByCampaignWithFilters(ctx context.Context, campaignID *uuid.UUID, marketplace *string, startDate, endDate time.Time) ([]model.CampaignStatResult, error)
	CountByMarketplaceWithFilters(ctx context.Context, campaignID *uuid.UUID, marketplace *string, startDate, endDate time.Time) ([]model.MarketplaceStatResult, error)
	FindTopProductsWithFilters(ctx context.Context, campaignID *uuid.UUID, marketplace *string, startDate, endDate time.Time, limit int) ([]model.TopProductResult, error)
}
