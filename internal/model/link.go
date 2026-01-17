package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Link represents an affiliate link
type Link struct {
	ID          uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProductID   uuid.UUID   `gorm:"type:uuid;not null;index:idx_links_product_campaign" json:"product_id"`
	CampaignID  uuid.UUID   `gorm:"type:uuid;not null;index:idx_links_product_campaign" json:"campaign_id"`
	Marketplace Marketplace `gorm:"type:varchar(20);not null;check:marketplace IN ('lazada', 'shopee')" json:"marketplace"`
	ShortCode   string      `gorm:"type:varchar(20);not null;uniqueIndex:idx_links_short_code" json:"short_code"`
	TargetURL   string      `gorm:"type:text;not null" json:"target_url"`
	CreatedAt   time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Product  Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Campaign Campaign `gorm:"foreignKey:CampaignID" json:"campaign,omitempty"`
	Clicks   []Click  `gorm:"foreignKey:LinkID;constraint:OnDelete:CASCADE" json:"clicks,omitempty"`
}

// TableName specifies the table name for Link
func (Link) TableName() string {
	return "links"
}

// BeforeCreate hook to set UUID if not set
func (l *Link) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}
