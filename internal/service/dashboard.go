package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jonosize/affiliate-platform/internal/database"
	"github.com/jonosize/affiliate-platform/internal/dto"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/model"
	"github.com/jonosize/affiliate-platform/internal/repository"
)

// DashboardService handles dashboard analytics business logic
type DashboardService struct {
	clickRepo    *repository.ClickRepository
	linkRepo     *repository.LinkRepository
	campaignRepo *repository.CampaignRepository
	productRepo  *repository.ProductRepository
	db           *database.DB
	logger       logger.Logger
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(db *database.DB, log logger.Logger) *DashboardService {
	return &DashboardService{
		clickRepo:    repository.NewClickRepository(db),
		linkRepo:     repository.NewLinkRepository(db),
		campaignRepo: repository.NewCampaignRepository(db),
		productRepo:  repository.NewProductRepository(db),
		db:           db,
		logger:       log,
	}
}

// GetDashboardStats returns aggregated dashboard statistics
func (s *DashboardService) GetDashboardStats(ctx context.Context, params dto.DashboardQueryParams) (*dto.DashboardStatsResponse, error) {
	// Build query conditions
	startDate := time.Time{}
	endDate := time.Now()
	if params.StartDate != nil {
		startDate = *params.StartDate
	}
	if params.EndDate != nil {
		endDate = *params.EndDate
	}

	// Get total clicks (with filters)
	totalClicks, err := s.getTotalClicks(ctx, params, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get total clicks: %w", err)
	}

	// Get total links
	totalLinks, err := s.getTotalLinks(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get total links: %w", err)
	}

	// Calculate CTR (Click-Through Rate)
	ctr := 0.0
	if totalLinks > 0 {
		ctr = (float64(totalClicks) / float64(totalLinks)) * 100.0
	}

	// Get campaign stats
	campaignStats, err := s.getCampaignStats(ctx, params, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign stats: %w", err)
	}

	// Get marketplace stats
	marketplaceStats, err := s.getMarketplaceStats(ctx, params, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get marketplace stats: %w", err)
	}

	// Get top products
	topProducts, err := s.getTopProducts(ctx, params, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get top products: %w", err)
	}

	// Get recent clicks
	recentClicks, err := s.getRecentClicks(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent clicks: %w", err)
	}

	return &dto.DashboardStatsResponse{
		TotalClicks:      totalClicks,
		TotalLinks:       totalLinks,
		CTR:              ctr,
		CampaignStats:    campaignStats,
		MarketplaceStats: marketplaceStats,
		TopProducts:      topProducts,
		RecentClicks:     recentClicks,
	}, nil
}

// getTotalClicks gets total click count with filters
func (s *DashboardService) getTotalClicks(ctx context.Context, params dto.DashboardQueryParams, startDate, endDate time.Time) (int64, error) {
	query := s.db.Read.WithContext(ctx).Model(&model.Click{})

	// Apply date range filter
	if !startDate.IsZero() {
		query = query.Where("timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("timestamp <= ?", endDate)
	}

	// Apply campaign filter (via link join)
	if params.CampaignID != nil {
		query = query.Joins("JOIN links ON clicks.link_id = links.id").
			Where("links.campaign_id = ?", *params.CampaignID)
	}

	// Apply marketplace filter (via link join)
	if params.Marketplace != nil {
		query = query.Joins("JOIN links ON clicks.link_id = links.id").
			Where("links.marketplace = ?", *params.Marketplace)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// getTotalLinks gets total link count with filters
func (s *DashboardService) getTotalLinks(ctx context.Context, params dto.DashboardQueryParams) (int64, error) {
	query := s.db.Read.WithContext(ctx).Model(&model.Link{})

	if params.CampaignID != nil {
		query = query.Where("campaign_id = ?", *params.CampaignID)
	}
	if params.Marketplace != nil {
		query = query.Where("marketplace = ?", *params.Marketplace)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// getCampaignStats gets click statistics grouped by campaign
func (s *DashboardService) getCampaignStats(ctx context.Context, params dto.DashboardQueryParams, startDate, endDate time.Time) ([]dto.CampaignStat, error) {
	query := s.db.Read.WithContext(ctx).
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
	if params.Marketplace != nil {
		query = query.Where("links.marketplace = ?", *params.Marketplace)
	}

	// Apply campaign filter
	if params.CampaignID != nil {
		query = query.Where("campaigns.id = ?", *params.CampaignID)
	}

	type result struct {
		CampaignID   uuid.UUID
		CampaignName string
		Clicks       int64
	}

	var results []result
	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	stats := make([]dto.CampaignStat, len(results))
	for i, r := range results {
		stats[i] = dto.CampaignStat{
			CampaignID:   r.CampaignID,
			CampaignName: r.CampaignName,
			Clicks:       r.Clicks,
		}
	}

	return stats, nil
}

// getMarketplaceStats gets click statistics grouped by marketplace
func (s *DashboardService) getMarketplaceStats(ctx context.Context, params dto.DashboardQueryParams, startDate, endDate time.Time) ([]dto.MarketplaceStat, error) {
	query := s.db.Read.WithContext(ctx).
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
	if params.CampaignID != nil {
		query = query.Where("links.campaign_id = ?", *params.CampaignID)
	}

	// Apply marketplace filter
	if params.Marketplace != nil {
		query = query.Where("links.marketplace = ?", *params.Marketplace)
	}

	type result struct {
		Marketplace string
		Clicks      int64
	}

	var results []result
	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	// Calculate total for percentage
	total := int64(0)
	for _, r := range results {
		total += r.Clicks
	}

	stats := make([]dto.MarketplaceStat, len(results))
	for i, r := range results {
		percentage := 0.0
		if total > 0 {
			percentage = (float64(r.Clicks) / float64(total)) * 100.0
		}
		stats[i] = dto.MarketplaceStat{
			Marketplace: r.Marketplace,
			Clicks:      r.Clicks,
			Percentage:  percentage,
		}
	}

	return stats, nil
}

// getTopProducts gets top-performing products by click count
func (s *DashboardService) getTopProducts(ctx context.Context, params dto.DashboardQueryParams, startDate, endDate time.Time) ([]dto.TopProduct, error) {
	query := s.db.Read.WithContext(ctx).
		Table("clicks").
		Select("products.id as product_id, products.title as product_name, links.marketplace, COUNT(clicks.id) as clicks").
		Joins("JOIN links ON clicks.link_id = links.id").
		Joins("JOIN products ON links.product_id = products.id").
		Group("products.id, products.title, links.marketplace").
		Order("clicks DESC").
		Limit(10)

	// Apply date range filter
	if !startDate.IsZero() {
		query = query.Where("clicks.timestamp >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("clicks.timestamp <= ?", endDate)
	}

	// Apply campaign filter
	if params.CampaignID != nil {
		query = query.Where("links.campaign_id = ?", *params.CampaignID)
	}

	// Apply marketplace filter
	if params.Marketplace != nil {
		query = query.Where("links.marketplace = ?", *params.Marketplace)
	}

	type result struct {
		ProductID   uuid.UUID
		ProductName string
		Marketplace string
		Clicks      int64
	}

	var results []result
	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	topProducts := make([]dto.TopProduct, len(results))
	for i, r := range results {
		topProducts[i] = dto.TopProduct{
			ProductID:   r.ProductID,
			ProductName: r.ProductName,
			Clicks:      r.Clicks,
			Marketplace: r.Marketplace,
		}
	}

	return topProducts, nil
}

// getRecentClicks gets the most recent clicks with related data
func (s *DashboardService) getRecentClicks(ctx context.Context, params dto.DashboardQueryParams) ([]dto.RecentClick, error) {
	// Get recent clicks from repository
	clicks, err := s.clickRepo.FindRecentClicks(ctx, 10)
	if err != nil {
		return nil, err
	}

	// Convert to DTO
	recentClicks := make([]dto.RecentClick, 0, len(clicks))
	for _, click := range clicks {
		// Skip if link is not loaded
		if click.Link.ID == uuid.Nil {
			continue
		}

		// Apply filters if specified
		if params.Marketplace != nil && string(click.Link.Marketplace) != *params.Marketplace {
			continue
		}
		if params.CampaignID != nil && click.Link.CampaignID != *params.CampaignID {
			continue
		}

		// Get product name (with fallback)
		productName := "Unknown Product"
		if click.Link.Product.ID != uuid.Nil {
			productName = click.Link.Product.Title
		}

		// Get campaign name (with fallback)
		campaignName := "Unknown Campaign"
		if click.Link.Campaign.ID != uuid.Nil {
			campaignName = click.Link.Campaign.Name
		}

		recentClick := dto.RecentClick{
			DateTime:     click.Timestamp,
			ProductID:    click.Link.ProductID,
			ProductName:  productName,
			Marketplace:  string(click.Link.Marketplace),
			CampaignID:   click.Link.CampaignID,
			CampaignName: campaignName,
		}

		recentClicks = append(recentClicks, recentClick)
	}

	return recentClicks, nil
}
