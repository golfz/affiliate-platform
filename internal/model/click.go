package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Click represents a click tracking record
type Click struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	LinkID    uuid.UUID `gorm:"type:uuid;not null;index:idx_clicks_link_id;index:idx_clicks_link_timestamp" json:"link_id"`
	Timestamp time.Time `gorm:"default:now();index:idx_clicks_timestamp;index:idx_clicks_link_timestamp" json:"timestamp"`
	Referrer  string    `gorm:"type:text" json:"referrer"`
	UserAgent string    `gorm:"type:text" json:"user_agent"`
	IPAddress string    `gorm:"type:inet" json:"ip_address"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relationships
	Link Link `gorm:"foreignKey:LinkID" json:"link,omitempty"`
}

// TableName specifies the table name for Click
func (Click) TableName() string {
	return "clicks"
}

// BeforeCreate hook to set UUID if not set
func (c *Click) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
