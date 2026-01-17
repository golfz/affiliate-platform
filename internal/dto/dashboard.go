package dto

import (
	"time"

	"github.com/google/uuid"
)

// DashboardStatsResponse represents dashboard statistics
type DashboardStatsResponse struct {
	TotalClicks      int64             `json:"total_clicks" example:"1250"`
	TotalLinks       int64             `json:"total_links" example:"45"`
	CTR              float64           `json:"ctr" example:"2.78"` // Click-through rate (percentage)
	CampaignStats    []CampaignStat    `json:"campaign_stats"`
	MarketplaceStats []MarketplaceStat `json:"marketplace_stats"`
	TopProducts      []TopProduct      `json:"top_products"`
	RecentClicks     []RecentClick     `json:"recent_clicks"`
}

// CampaignStat represents click statistics for a campaign
type CampaignStat struct {
	CampaignID   uuid.UUID `json:"campaign_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	CampaignName string    `json:"campaign_name" example:"Summer Deal 2025"`
	Clicks       int64     `json:"clicks" example:"450"`
}

// MarketplaceStat represents click statistics by marketplace
type MarketplaceStat struct {
	Marketplace string  `json:"marketplace" example:"lazada"`
	Clicks      int64   `json:"clicks" example:"750"`
	Percentage  float64 `json:"percentage" example:"60.0"`
}

// TopProduct represents a top-performing product
type TopProduct struct {
	ProductID   uuid.UUID `json:"product_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ProductName string    `json:"product_name" example:"Product Title"`
	Clicks      int64     `json:"clicks" example:"120"`
	Marketplace string    `json:"marketplace" example:"lazada"`
}

// RecentClick represents a recent click event
type RecentClick struct {
	DateTime     time.Time `json:"datetime" example:"2025-01-15T10:30:00Z"`
	ProductID    uuid.UUID `json:"product_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ProductName  string    `json:"product_name" example:"Product Title"`
	Marketplace  string    `json:"marketplace" example:"lazada"`
	CampaignID   uuid.UUID `json:"campaign_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	CampaignName string    `json:"campaign_name" example:"Summer Deal 2025"`
}

// DashboardQueryParams represents query parameters for dashboard filtering
type DashboardQueryParams struct {
	CampaignID  *uuid.UUID `json:"campaign_id,omitempty"`
	Marketplace *string    `json:"marketplace,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}
