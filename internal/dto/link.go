package dto

import (
	"github.com/google/uuid"
)

// CreateLinkRequest represents the request to create an affiliate link
type CreateLinkRequest struct {
	ProductID   uuid.UUID `json:"product_id" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	CampaignID  uuid.UUID `json:"campaign_id" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Marketplace string    `json:"marketplace" validate:"required,oneof=lazada shopee" example:"lazada"`
}

// LinkResponse represents a link response
type LinkResponse struct {
	ID        uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ShortCode string    `json:"short_code" example:"abc123xyz"`
	TargetURL string    `json:"target_url" example:"https://www.lazada.co.th/products/...?utm_source=...&utm_medium=affiliate&utm_campaign=summer_2025"`
	FullURL   string    `json:"full_url" example:"https://demo.jonosize.com/go/abc123xyz"`
}
