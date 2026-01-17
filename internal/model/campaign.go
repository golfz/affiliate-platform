package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Campaign represents a marketing campaign
type Campaign struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(200);not null" json:"name"`
	UTMCampaign string    `gorm:"type:varchar(100);not null" json:"utm_campaign"`
	StartAt     time.Time `gorm:"not null;index:idx_campaigns_dates" json:"start_at"`
	EndAt       time.Time `gorm:"not null;index:idx_campaigns_dates;check:end_at > start_at" json:"end_at"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	CampaignProducts []CampaignProduct `gorm:"foreignKey:CampaignID;constraint:OnDelete:CASCADE" json:"campaign_products,omitempty"`
	Links            []Link            `gorm:"foreignKey:CampaignID;constraint:OnDelete:CASCADE" json:"links,omitempty"`
}

// TableName specifies the table name for Campaign
func (Campaign) TableName() string {
	return "campaigns"
}

// BeforeCreate hook to set UUID if not set
func (c *Campaign) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// CampaignProduct represents the many-to-many relationship between Campaign and Product
type CampaignProduct struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CampaignID uuid.UUID `gorm:"type:uuid;not null;index:idx_campaign_products_campaign;uniqueIndex:idx_campaign_product_unique" json:"campaign_id"`
	ProductID  uuid.UUID `gorm:"type:uuid;not null;index:idx_campaign_products_product;uniqueIndex:idx_campaign_product_unique" json:"product_id"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relationships
	Campaign Campaign `gorm:"foreignKey:CampaignID" json:"campaign,omitempty"`
	Product  Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// TableName specifies the table name for CampaignProduct
func (CampaignProduct) TableName() string {
	return "campaign_products"
}

// BeforeCreate hook to set UUID if not set
func (cp *CampaignProduct) BeforeCreate(tx *gorm.DB) error {
	if cp.ID == uuid.Nil {
		cp.ID = uuid.New()
	}
	return nil
}
