package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Marketplace type (matching adapters.Marketplace for compatibility)
type Marketplace string

const (
	MarketplaceLazada Marketplace = "lazada"
	MarketplaceShopee Marketplace = "shopee"
)

// Offer represents a price offer from a marketplace
type Offer struct {
	ID                    uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProductID             uuid.UUID   `gorm:"type:uuid;not null;index" json:"product_id"`
	Marketplace           Marketplace `gorm:"type:varchar(20);not null;check:marketplace IN ('lazada', 'shopee');index" json:"marketplace"`
	StoreName             string      `gorm:"type:varchar(200)" json:"store_name"`
	Price                 float64     `gorm:"type:decimal(10,2);not null;check:price >= 0" json:"price"`
	MarketplaceProductURL string      `gorm:"type:text;not null" json:"marketplace_product_url"`
	LastCheckedAt         time.Time   `gorm:"default:now();index" json:"last_checked_at"`
	CreatedAt             time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time   `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// TableName specifies the table name for Offer
func (Offer) TableName() string {
	return "offers"
}

// BeforeCreate hook to set UUID if not set
func (o *Offer) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}
