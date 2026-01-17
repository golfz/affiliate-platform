package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateProductRequest represents the request to create a product
type CreateProductRequest struct {
	Source     string `json:"source" validate:"required" example:"https://www.lazada.co.th/products/example-i123456.html"`
	SourceType string `json:"sourceType" validate:"required,oneof=url sku" example:"url"`
}

// ProductResponse represents a product response
type ProductResponse struct {
	ID        uuid.UUID       `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Title     string          `json:"title" example:"Product Title"`
	ImageURL  string          `json:"image_url" example:"https://example.com/image.jpg"`
	Offers    []OfferResponse `json:"offers,omitempty"`
	CreatedAt time.Time       `json:"created_at" example:"2025-01-15T10:00:00Z"`
}

// OfferResponse represents an offer response
type OfferResponse struct {
	ID            uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Marketplace   string    `json:"marketplace" example:"lazada"`
	StoreName     string    `json:"store_name" example:"Store Name"`
	Price         float64   `json:"price" example:"299.00"`
	LastCheckedAt time.Time `json:"last_checked_at" example:"2025-01-15T10:00:00Z"`
}

// ProductOffersResponse represents the response for product offers
type ProductOffersResponse struct {
	ProductID uuid.UUID       `json:"product_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Offers    []OfferResponse `json:"offers"`
	BestPrice *BestPrice      `json:"best_price,omitempty"`
}

// BestPrice represents the best price offer
type BestPrice struct {
	Marketplace string  `json:"marketplace" example:"shopee"`
	Price       float64 `json:"price" example:"279.00"`
}
