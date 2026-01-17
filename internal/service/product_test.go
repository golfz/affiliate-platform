package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jonosize/affiliate-platform/internal/dto"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/model"
	"github.com/jonosize/affiliate-platform/pkg/adapters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockProductRepository is a mock implementation of ProductRepositoryInterface
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, product *model.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Product), args.Error(1)
}

func (m *MockProductRepository) FindAll(ctx context.Context, limit, offset int) ([]*model.Product, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*model.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) Update(ctx context.Context, product *model.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockOfferRepository is a mock implementation of OfferRepositoryInterface
type MockOfferRepository struct {
	mock.Mock
}

func (m *MockOfferRepository) Create(ctx context.Context, offer *model.Offer) error {
	args := m.Called(ctx, offer)
	return args.Error(0)
}

func (m *MockOfferRepository) FindByProductID(ctx context.Context, productID uuid.UUID) ([]*model.Offer, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Offer), args.Error(1)
}

func (m *MockOfferRepository) FindByProductIDAndMarketplace(ctx context.Context, productID uuid.UUID, marketplace model.Marketplace) (*model.Offer, error) {
	args := m.Called(ctx, productID, marketplace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Offer), args.Error(1)
}

func (m *MockOfferRepository) Update(ctx context.Context, offer *model.Offer) error {
	args := m.Called(ctx, offer)
	return args.Error(0)
}

func (m *MockOfferRepository) Upsert(ctx context.Context, offer *model.Offer) error {
	args := m.Called(ctx, offer)
	return args.Error(0)
}

func (m *MockOfferRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockMarketplaceAdapter is a mock implementation of adapters.MarketplaceAdapter
type MockMarketplaceAdapter struct {
	mock.Mock
}

func (m *MockMarketplaceAdapter) FetchProduct(ctx context.Context, source string, sourceType adapters.SourceType) (*adapters.ProductData, error) {
	args := m.Called(ctx, source, sourceType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*adapters.ProductData), args.Error(1)
}

func (m *MockMarketplaceAdapter) FetchOffer(ctx context.Context, productURL string) (*adapters.OfferData, error) {
	args := m.Called(ctx, productURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*adapters.OfferData), args.Error(1)
}

func (m *MockMarketplaceAdapter) Marketplace() adapters.Marketplace {
	args := m.Called()
	return args.Get(0).(adapters.Marketplace)
}

// ProductServiceTestSuite is the test suite for ProductService
type ProductServiceTestSuite struct {
	suite.Suite
	service       *ProductService
	productRepo   *MockProductRepository
	offerRepo     *MockOfferRepository
	lazadaAdapter *MockMarketplaceAdapter
	shopeeAdapter *MockMarketplaceAdapter
	logger        logger.Logger
	ctx           context.Context
}

func (suite *ProductServiceTestSuite) SetupTest() {
	suite.productRepo = new(MockProductRepository)
	suite.offerRepo = new(MockOfferRepository)
	suite.lazadaAdapter = new(MockMarketplaceAdapter)
	suite.shopeeAdapter = new(MockMarketplaceAdapter)
	log, err := logger.NewZapLogger("info")
	if err != nil {
		suite.T().Fatal("Failed to create logger:", err)
	}
	suite.logger = log
	suite.ctx = context.Background()
	suite.service = NewProductService(
		suite.productRepo,
		suite.offerRepo,
		suite.lazadaAdapter,
		suite.shopeeAdapter,
		suite.logger,
	)
}

func (suite *ProductServiceTestSuite) TearDownTest() {
	suite.productRepo.AssertExpectations(suite.T())
	suite.offerRepo.AssertExpectations(suite.T())
	suite.lazadaAdapter.AssertExpectations(suite.T())
	suite.shopeeAdapter.AssertExpectations(suite.T())
}

// TestProductService_CreateProduct tests the CreateProduct method
func (suite *ProductServiceTestSuite) TestProductService_CreateProduct() {
	tests := []struct {
		name        string
		req         dto.CreateProductRequest
		setupMock   func()
		wantErr     bool
		errContains string
	}{
		{
			name: "success with Lazada URL",
			req: dto.CreateProductRequest{
				LazadaURL: "https://www.lazada.co.th/products/test-i123456.html",
			},
			setupMock: func() {
				productData := &adapters.ProductData{
					Title:                 "Test Product",
					ImageURL:              "https://example.com/image.jpg",
					MarketplaceProductURL: "https://www.lazada.co.th/products/test-i123456.html",
					SourceID:              1,
				}
				suite.lazadaAdapter.On("FetchProduct", suite.ctx, "https://www.lazada.co.th/products/test-i123456.html", adapters.SourceTypeURL).
					Return(productData, nil).Once()

				offerData := &adapters.OfferData{
					StoreName:             "Test Store",
					Price:                 299.00,
					MarketplaceProductURL: "https://www.lazada.co.th/products/test-i123456.html",
				}
				suite.lazadaAdapter.On("FetchOffer", suite.ctx, "https://www.lazada.co.th/products/test-i123456.html").
					Return(offerData, nil).Once()

				product := &model.Product{
					ID:        uuid.New(),
					Title:     "Test Product",
					ImageURL:  "https://example.com/image.jpg",
					CreatedAt: time.Now(),
				}
				suite.productRepo.On("Create", suite.ctx, mock.AnythingOfType("*model.Product")).
					Run(func(args mock.Arguments) {
						p := args.Get(1).(*model.Product)
						p.ID = product.ID
						p.CreatedAt = product.CreatedAt
					}).
					Return(nil).Once()

				suite.offerRepo.On("Upsert", suite.ctx, mock.AnythingOfType("*model.Offer")).
					Return(nil).Once()

				suite.offerRepo.On("FindByProductID", suite.ctx, product.ID).
					Return([]*model.Offer{
						{
							ID:            uuid.New(),
							ProductID:     product.ID,
							Marketplace:   model.MarketplaceLazada,
							StoreName:     "Test Store",
							Price:         299.00,
							LastCheckedAt: time.Now(),
						},
					}, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "success with Shopee URL",
			req: dto.CreateProductRequest{
				ShopeeURL: "https://shopee.co.th/product/test-i123456",
			},
			setupMock: func() {
				productData := &adapters.ProductData{
					Title:                 "Test Product",
					ImageURL:              "https://example.com/image.jpg",
					MarketplaceProductURL: "https://shopee.co.th/product/test-i123456",
					SourceID:              2,
				}
				suite.shopeeAdapter.On("FetchProduct", suite.ctx, "https://shopee.co.th/product/test-i123456", adapters.SourceTypeURL).
					Return(productData, nil).Once()

				offerData := &adapters.OfferData{
					StoreName:             "Shopee Store",
					Price:                 279.00,
					MarketplaceProductURL: "https://shopee.co.th/product/test-i123456",
				}
				suite.shopeeAdapter.On("FetchOffer", suite.ctx, "https://shopee.co.th/product/test-i123456").
					Return(offerData, nil).Once()

				product := &model.Product{
					ID:        uuid.New(),
					Title:     "Test Product",
					ImageURL:  "https://example.com/image.jpg",
					CreatedAt: time.Now(),
				}
				suite.productRepo.On("Create", suite.ctx, mock.AnythingOfType("*model.Product")).
					Run(func(args mock.Arguments) {
						p := args.Get(1).(*model.Product)
						p.ID = product.ID
						p.CreatedAt = product.CreatedAt
					}).
					Return(nil).Once()

				suite.offerRepo.On("Upsert", suite.ctx, mock.AnythingOfType("*model.Offer")).
					Return(nil).Once()

				suite.offerRepo.On("FindByProductID", suite.ctx, product.ID).
					Return([]*model.Offer{
						{
							ID:            uuid.New(),
							ProductID:     product.ID,
							Marketplace:   model.MarketplaceShopee,
							StoreName:     "Shopee Store",
							Price:         279.00,
							LastCheckedAt: time.Now(),
						},
					}, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "error when both URLs are empty",
			req: dto.CreateProductRequest{
				LazadaURL: "",
				ShopeeURL: "",
			},
			setupMock:   func() {},
			wantErr:     true,
			errContains: "at least one URL",
		},
		{
			name: "error when FetchProduct fails",
			req: dto.CreateProductRequest{
				LazadaURL: "https://www.lazada.co.th/products/test-i123456.html",
			},
			setupMock: func() {
				suite.lazadaAdapter.On("FetchProduct", suite.ctx, "https://www.lazada.co.th/products/test-i123456.html", adapters.SourceTypeURL).
					Return(nil, errors.New("fetch failed")).Once()
			},
			wantErr:     true,
			errContains: "failed to fetch product data",
		},
		{
			name: "error when Create fails",
			req: dto.CreateProductRequest{
				LazadaURL: "https://www.lazada.co.th/products/test-i123456.html",
			},
			setupMock: func() {
				productData := &adapters.ProductData{
					Title:                 "Test Product",
					ImageURL:              "https://example.com/image.jpg",
					MarketplaceProductURL: "https://www.lazada.co.th/products/test-i123456.html",
				}
				suite.lazadaAdapter.On("FetchProduct", suite.ctx, "https://www.lazada.co.th/products/test-i123456.html", adapters.SourceTypeURL).
					Return(productData, nil).Once()

				suite.productRepo.On("Create", suite.ctx, mock.AnythingOfType("*model.Product")).
					Return(errors.New("database error")).Once()
			},
			wantErr:     true,
			errContains: "failed to create product",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset mocks
			suite.productRepo.ExpectedCalls = nil
			suite.offerRepo.ExpectedCalls = nil
			suite.lazadaAdapter.ExpectedCalls = nil
			suite.shopeeAdapter.ExpectedCalls = nil

			// Setup mocks
			tt.setupMock()

			// Execute
			result, err := suite.service.CreateProduct(suite.ctx, tt.req)

			// Assert
			if tt.wantErr {
				assert.Error(suite.T(), err)
				if tt.errContains != "" {
					assert.Contains(suite.T(), err.Error(), tt.errContains)
				}
				assert.Nil(suite.T(), result)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), result)
				assert.NotEqual(suite.T(), uuid.Nil, result.ID)
				assert.Equal(suite.T(), tt.req.LazadaURL != "" || tt.req.ShopeeURL != "", result.Title != "")
			}
		})
	}
}

// TestProductService_GetProductOffers tests the GetProductOffers method
func (suite *ProductServiceTestSuite) TestProductService_GetProductOffers() {
	productID := uuid.New()
	product := &model.Product{
		ID:    productID,
		Title: "Test Product",
	}

	tests := []struct {
		name        string
		productID   uuid.UUID
		setupMock   func()
		wantErr     bool
		errContains string
	}{
		{
			name:      "success",
			productID: productID,
			setupMock: func() {
				suite.productRepo.On("FindByID", suite.ctx, productID).
					Return(product, nil).Once()

				offers := []*model.Offer{
					{
						ID:            uuid.New(),
						ProductID:     productID,
						Marketplace:   model.MarketplaceLazada,
						StoreName:     "Lazada Store",
						Price:         299.00,
						LastCheckedAt: time.Now(),
					},
					{
						ID:            uuid.New(),
						ProductID:     productID,
						Marketplace:   model.MarketplaceShopee,
						StoreName:     "Shopee Store",
						Price:         279.00,
						LastCheckedAt: time.Now(),
					},
				}
				suite.offerRepo.On("FindByProductID", suite.ctx, productID).
					Return(offers, nil).Once()
			},
			wantErr: false,
		},
		{
			name:      "error when product not found",
			productID: productID,
			setupMock: func() {
				suite.productRepo.On("FindByID", suite.ctx, productID).
					Return(nil, errors.New("not found")).Once()
			},
			wantErr:     true,
			errContains: "product not found",
		},
		{
			name:      "error when FindByProductID fails",
			productID: productID,
			setupMock: func() {
				suite.productRepo.On("FindByID", suite.ctx, productID).
					Return(product, nil).Once()

				suite.offerRepo.On("FindByProductID", suite.ctx, productID).
					Return(nil, errors.New("database error")).Once()
			},
			wantErr:     true,
			errContains: "failed to get offers",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset mocks
			suite.productRepo.ExpectedCalls = nil
			suite.offerRepo.ExpectedCalls = nil

			// Setup mocks
			tt.setupMock()

			// Execute
			result, err := suite.service.GetProductOffers(suite.ctx, tt.productID)

			// Assert
			if tt.wantErr {
				assert.Error(suite.T(), err)
				if tt.errContains != "" {
					assert.Contains(suite.T(), err.Error(), tt.errContains)
				}
				assert.Nil(suite.T(), result)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), result)
				assert.Equal(suite.T(), productID, result.ProductID)
				assert.NotNil(suite.T(), result.BestPrice)
				assert.Equal(suite.T(), "shopee", result.BestPrice.Marketplace) // Shopee has lower price
			}
		})
	}
}

// TestProductService_GetAllProducts tests the GetAllProducts method
func (suite *ProductServiceTestSuite) TestProductService_GetAllProducts() {
	tests := []struct {
		name        string
		limit       int
		offset      int
		setupMock   func()
		wantErr     bool
		errContains string
	}{
		{
			name:   "success",
			limit:  10,
			offset: 0,
			setupMock: func() {
				products := []*model.Product{
					{
						ID:        uuid.New(),
						Title:     "Product 1",
						ImageURL:  "https://example.com/image1.jpg",
						CreatedAt: time.Now(),
					},
					{
						ID:        uuid.New(),
						Title:     "Product 2",
						ImageURL:  "https://example.com/image2.jpg",
						CreatedAt: time.Now(),
					},
				}
				suite.productRepo.On("FindAll", suite.ctx, 10, 0).
					Return(products, int64(2), nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "error when FindAll fails",
			limit:  10,
			offset: 0,
			setupMock: func() {
				suite.productRepo.On("FindAll", suite.ctx, 10, 0).
					Return(nil, int64(0), errors.New("database error")).Once()
			},
			wantErr:     true,
			errContains: "failed to get products",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset mocks
			suite.productRepo.ExpectedCalls = nil

			// Setup mocks
			tt.setupMock()

			// Execute
			result, err := suite.service.GetAllProducts(suite.ctx, tt.limit, tt.offset)

			// Assert
			if tt.wantErr {
				assert.Error(suite.T(), err)
				if tt.errContains != "" {
					assert.Contains(suite.T(), err.Error(), tt.errContains)
				}
				assert.Nil(suite.T(), result)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), result)
				assert.Len(suite.T(), result, 2)
			}
		})
	}
}

// TestProductService_DeleteProduct tests the DeleteProduct method
func (suite *ProductServiceTestSuite) TestProductService_DeleteProduct() {
	productID := uuid.New()
	product := &model.Product{
		ID:    productID,
		Title: "Test Product",
	}

	tests := []struct {
		name        string
		productID   uuid.UUID
		setupMock   func()
		wantErr     bool
		errContains string
	}{
		{
			name:      "success",
			productID: productID,
			setupMock: func() {
				suite.productRepo.On("FindByID", suite.ctx, productID).
					Return(product, nil).Once()

				suite.productRepo.On("Delete", suite.ctx, productID).
					Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:      "error when product not found",
			productID: productID,
			setupMock: func() {
				suite.productRepo.On("FindByID", suite.ctx, productID).
					Return(nil, errors.New("not found")).Once()
			},
			wantErr:     true,
			errContains: "product not found",
		},
		{
			name:      "error when Delete fails",
			productID: productID,
			setupMock: func() {
				suite.productRepo.On("FindByID", suite.ctx, productID).
					Return(product, nil).Once()

				suite.productRepo.On("Delete", suite.ctx, productID).
					Return(errors.New("database error")).Once()
			},
			wantErr:     true,
			errContains: "failed to delete product",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset mocks
			suite.productRepo.ExpectedCalls = nil

			// Setup mocks
			tt.setupMock()

			// Execute
			err := suite.service.DeleteProduct(suite.ctx, tt.productID)

			// Assert
			if tt.wantErr {
				assert.Error(suite.T(), err)
				if tt.errContains != "" {
					assert.Contains(suite.T(), err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(suite.T(), err)
			}
		})
	}
}

func TestProductServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ProductServiceTestSuite))
}
