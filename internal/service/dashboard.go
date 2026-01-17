package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jonosize/affiliate-platform/internal/dto"
	"github.com/jonosize/affiliate-platform/internal/logger"
)

// DashboardService handles dashboard analytics business logic
type DashboardService struct {
	clickRepo    ClickRepositoryInterface
	linkRepo     LinkRepositoryInterface
	campaignRepo CampaignRepositoryInterface
	productRepo  ProductRepositoryInterface
	logger       logger.Logger
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(
	clickRepo ClickRepositoryInterface,
	linkRepo LinkRepositoryInterface,
	campaignRepo CampaignRepositoryInterface,
	productRepo ProductRepositoryInterface,
	log logger.Logger,
) *DashboardService {
	return &DashboardService{
		clickRepo:    clickRepo,
		linkRepo:     linkRepo,
		campaignRepo: campaignRepo,
		productRepo:  productRepo,
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
	return s.clickRepo.CountWithFilters(ctx, params.CampaignID, params.Marketplace, startDate, endDate)
}

// getTotalLinks gets total link count with filters
func (s *DashboardService) getTotalLinks(ctx context.Context, params dto.DashboardQueryParams) (int64, error) {
	return s.linkRepo.CountWithFilters(ctx, params.CampaignID, params.Marketplace)
}

// getCampaignStats gets click statistics grouped by campaign
func (s *DashboardService) getCampaignStats(ctx context.Context, params dto.DashboardQueryParams, startDate, endDate time.Time) ([]dto.CampaignStat, error) {
	results, err := s.clickRepo.CountByCampaignWithFilters(ctx, params.CampaignID, params.Marketplace, startDate, endDate)
	if err != nil {
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
	results, err := s.clickRepo.CountByMarketplaceWithFilters(ctx, params.CampaignID, params.Marketplace, startDate, endDate)
	if err != nil {
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
	results, err := s.clickRepo.FindTopProductsWithFilters(ctx, params.CampaignID, params.Marketplace, startDate, endDate, 10)
	if err != nil {
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
