package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Product represents a product entity
type Product struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title     string    `gorm:"type:varchar(500);not null" json:"title"`
	ImageURL  string    `gorm:"type:text" json:"image_url"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Offers []Offer `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"offers,omitempty"`
}

// TableName specifies the table name for Product
func (Product) TableName() string {
	return "products"
}

// BeforeCreate hook to set UUID if not set
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
