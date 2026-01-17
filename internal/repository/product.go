package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/jonosize/affiliate-platform/internal/database"
	"github.com/jonosize/affiliate-platform/internal/model"
)

// ProductRepository handles product database operations
type ProductRepository struct {
	db *database.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *database.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create creates a new product (uses write DB)
func (r *ProductRepository) Create(ctx context.Context, product *model.Product) error {
	return r.db.Write.WithContext(ctx).Create(product).Error
}

// FindByID finds a product by ID (uses read DB)
func (r *ProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	var product model.Product
	err := r.db.Read.WithContext(ctx).Preload("Offers").First(&product, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// FindAll finds all products (uses read DB)
func (r *ProductRepository) FindAll(ctx context.Context, limit, offset int) ([]*model.Product, int64, error) {
	var products []*model.Product
	var total int64

	// Count total
	if err := r.db.Read.WithContext(ctx).Model(&model.Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Find with pagination
	err := r.db.Read.WithContext(ctx).
		Preload("Offers").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&products).Error

	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// Update updates a product (uses write DB)
func (r *ProductRepository) Update(ctx context.Context, product *model.Product) error {
	return r.db.Write.WithContext(ctx).Save(product).Error
}

// Delete deletes a product (uses write DB)
func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.Write.WithContext(ctx).Delete(&model.Product{}, "id = ?", id).Error
}
