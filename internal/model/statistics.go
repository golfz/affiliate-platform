package model

import (
	"github.com/google/uuid"
)

// CampaignStatResult represents campaign statistics result from repository
type CampaignStatResult struct {
	CampaignID   uuid.UUID
	CampaignName string
	Clicks       int64
}

// MarketplaceStatResult represents marketplace statistics result from repository
type MarketplaceStatResult struct {
	Marketplace string
	Clicks      int64
}

// TopProductResult represents top product statistics result from repository
type TopProductResult struct {
	ProductID   uuid.UUID
	ProductName string
	Marketplace string
	Clicks      int64
}
