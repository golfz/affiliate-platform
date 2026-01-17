package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jonosize/affiliate-platform/internal/database"
	"github.com/jonosize/affiliate-platform/internal/model"
)

// OfferRepository handles offer database operations
type OfferRepository struct {
	db *database.DB
}

// NewOfferRepository creates a new offer repository
func NewOfferRepository(db *database.DB) *OfferRepository {
	return &OfferRepository{db: db}
}

// Create creates a new offer (uses write DB)
func (r *OfferRepository) Create(ctx context.Context, offer *model.Offer) error {
	return r.db.Write.WithContext(ctx).Create(offer).Error
}

// FindByProductID finds all offers for a product (uses read DB)
func (r *OfferRepository) FindByProductID(ctx context.Context, productID uuid.UUID) ([]*model.Offer, error) {
	var offers []*model.Offer
	err := r.db.Read.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("price ASC").
		Find(&offers).Error
	if err != nil {
		return nil, err
	}
	return offers, nil
}

// FindByProductIDAndMarketplace finds an offer by product ID and marketplace (uses read DB)
func (r *OfferRepository) FindByProductIDAndMarketplace(ctx context.Context, productID uuid.UUID, marketplace model.Marketplace) (*model.Offer, error) {
	var offer model.Offer
	err := r.db.Read.WithContext(ctx).
		Where("product_id = ? AND marketplace = ?", productID, marketplace).
		First(&offer).Error
	if err != nil {
		return nil, err
	}
	return &offer, nil
}

// Update updates an offer (uses write DB)
func (r *OfferRepository) Update(ctx context.Context, offer *model.Offer) error {
	return r.db.Write.WithContext(ctx).Save(offer).Error
}

// Upsert creates or updates an offer (uses write DB)
func (r *OfferRepository) Upsert(ctx context.Context, offer *model.Offer) error {
	return r.db.Write.WithContext(ctx).
		Where("product_id = ? AND marketplace = ?", offer.ProductID, offer.Marketplace).
		Assign(*offer).
		FirstOrCreate(offer).Error
}

// Delete deletes an offer (uses write DB)
func (r *OfferRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.Write.WithContext(ctx).Delete(&model.Offer{}, "id = ?", id).Error
}
