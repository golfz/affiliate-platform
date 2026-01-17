package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateCampaignRequest represents the request to create a campaign
type CreateCampaignRequest struct {
	Name        string      `json:"name" validate:"required" example:"Summer Deal 2025"`
	UTMCampaign string      `json:"utm_campaign" validate:"required" example:"summer_2025"`
	StartAt     time.Time   `json:"start_at" validate:"required" example:"2025-06-01T00:00:00Z"`
	EndAt       time.Time   `json:"end_at" validate:"required" example:"2025-08-31T23:59:59Z"`
	ProductIDs  []uuid.UUID `json:"product_ids,omitempty" example:"[\"123e4567-e89b-12d3-a456-426614174000\"]"`
}

// CampaignResponse represents a campaign response
type CampaignResponse struct {
	ID          uuid.UUID   `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name        string      `json:"name" example:"Summer Deal 2025"`
	UTMCampaign string      `json:"utm_campaign" example:"summer_2025"`
	StartAt     time.Time   `json:"start_at" example:"2025-06-01T00:00:00Z"`
	EndAt       time.Time   `json:"end_at" example:"2025-08-31T23:59:59Z"`
	CreatedAt   time.Time   `json:"created_at" example:"2025-01-15T10:00:00Z"`
	ProductIDs  []uuid.UUID `json:"product_ids,omitempty" example:"[\"123e4567-e89b-12d3-a456-426614174000\"]"` // Product IDs in this campaign
}

// CampaignPublicResponse represents a public campaign response (for public landing page)
type CampaignPublicResponse struct {
	ID       uuid.UUID         `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name     string            `json:"name" example:"Summer Deal 2025"`
	StartAt  time.Time         `json:"start_at" example:"2025-06-01T00:00:00Z"`
	EndAt    time.Time         `json:"end_at" example:"2025-08-31T23:59:59Z"`
	Products []CampaignProduct `json:"products"`
}

// CampaignProduct represents a product in a campaign (public view)
type CampaignProduct struct {
	ID        uuid.UUID       `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Title     string          `json:"title" example:"Product Title"`
	ImageURL  string          `json:"image_url" example:"https://example.com/image.jpg"`
	Offers    []OfferResponse `json:"offers"`
	BestPrice *BestPrice      `json:"best_price,omitempty"`
	Links     []ProductLink   `json:"links,omitempty"` // Links for this product in the campaign
}

// ProductLink represents an affiliate link for a product
type ProductLink struct {
	Marketplace string `json:"marketplace" example:"lazada"`
	ShortCode   string `json:"short_code" example:"abc123xyz"`
	FullURL     string `json:"full_url" example:"https://demo.jonosize.com/go/abc123xyz"`
}

// UpdateCampaignRequest represents the request to update a campaign
type UpdateCampaignRequest struct {
	Name        string      `json:"name,omitempty" example:"Summer Deal 2025"`
	UTMCampaign string      `json:"utm_campaign,omitempty" example:"summer_2025"`
	StartAt     *time.Time  `json:"start_at,omitempty" example:"2025-06-01T00:00:00Z"`
	EndAt       *time.Time  `json:"end_at,omitempty" example:"2025-08-31T23:59:59Z"`
	ProductIDs  []uuid.UUID `json:"product_ids,omitempty" example:"[\"123e4567-e89b-12d3-a456-426614174000\"]"`
}

// UpdateCampaignProductsRequest represents the request to update products in a campaign
type UpdateCampaignProductsRequest struct {
	ProductIDs []uuid.UUID `json:"product_ids" example:"[\"123e4567-e89b-12d3-a456-426614174000\"]"`
}
