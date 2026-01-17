package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jonosize/affiliate-platform/internal/config"
	"github.com/jonosize/affiliate-platform/internal/dto"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockCampaignRepository is a mock implementation of CampaignRepositoryInterface
type MockCampaignRepository struct {
	mock.Mock
}

func (m *MockCampaignRepository) Create(ctx context.Context, campaign *model.Campaign) error {
	args := m.Called(ctx, campaign)
	return args.Error(0)
}

func (m *MockCampaignRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Campaign, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Campaign), args.Error(1)
}

func (m *MockCampaignRepository) FindAll(ctx context.Context, limit, offset int) ([]*model.Campaign, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*model.Campaign), args.Get(1).(int64), args.Error(2)
}

func (m *MockCampaignRepository) Update(ctx context.Context, campaign *model.Campaign) error {
	args := m.Called(ctx, campaign)
	return args.Error(0)
}

func (m *MockCampaignRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCampaignRepository) AddProducts(ctx context.Context, campaignID uuid.UUID, productIDs []uuid.UUID) error {
	args := m.Called(ctx, campaignID, productIDs)
	return args.Error(0)
}

func (m *MockCampaignRepository) RemoveProducts(ctx context.Context, campaignID uuid.UUID, productIDs []uuid.UUID) error {
	args := m.Called(ctx, campaignID, productIDs)
	return args.Error(0)
}

func (m *MockCampaignRepository) UpdateCampaignProducts(ctx context.Context, campaignID uuid.UUID, productIDs []uuid.UUID) error {
	args := m.Called(ctx, campaignID, productIDs)
	return args.Error(0)
}

// LinkServiceTestSuite is the test suite for LinkService
type LinkServiceTestSuite struct {
	suite.Suite
	service      *LinkService
	linkRepo     *MockLinkRepository
	campaignRepo *MockCampaignRepository
	productRepo  *MockProductRepository
	offerRepo    *MockOfferRepository
	cfg          config.Config
	logger       logger.Logger
	ctx          context.Context
}

func (suite *LinkServiceTestSuite) SetupTest() {
	suite.linkRepo = new(MockLinkRepository)
	suite.campaignRepo = new(MockCampaignRepository)
	suite.productRepo = new(MockProductRepository)
	suite.offerRepo = new(MockOfferRepository)
	suite.cfg = &MockConfig{apiBaseURL: "https://api.example.com"}
	log, err := logger.NewZapLogger("info")
	if err != nil {
		suite.T().Fatal("Failed to create logger:", err)
	}
	suite.logger = log
	suite.ctx = context.Background()
	suite.service = NewLinkService(
		suite.linkRepo,
		suite.campaignRepo,
		suite.productRepo,
		suite.offerRepo,
		suite.cfg,
		suite.logger,
	)
}

func (suite *LinkServiceTestSuite) TearDownTest() {
	suite.linkRepo.AssertExpectations(suite.T())
	suite.campaignRepo.AssertExpectations(suite.T())
	suite.productRepo.AssertExpectations(suite.T())
	suite.offerRepo.AssertExpectations(suite.T())
}

// MockConfig is a simple mock config for testing
type MockConfig struct {
	apiBaseURL string
}

func (m *MockConfig) GetAPIBaseURL() string {
	return m.apiBaseURL
}

// Implement all Config interface methods (stubs for testing)
func (m *MockConfig) GetDatabaseWriteHost() string           { return "" }
func (m *MockConfig) GetDatabaseWritePort() int              { return 0 }
func (m *MockConfig) GetDatabaseWriteUser() string           { return "" }
func (m *MockConfig) GetDatabaseWritePassword() string       { return "" }
func (m *MockConfig) GetDatabaseWriteDBName() string         { return "" }
func (m *MockConfig) GetDatabaseWriteSSLMode() string        { return "" }
func (m *MockConfig) GetDatabaseReadHost() string            { return "" }
func (m *MockConfig) GetDatabaseReadPort() int               { return 0 }
func (m *MockConfig) GetDatabaseReadUser() string            { return "" }
func (m *MockConfig) GetDatabaseReadPassword() string        { return "" }
func (m *MockConfig) GetDatabaseReadDBName() string          { return "" }
func (m *MockConfig) GetDatabaseReadSSLMode() string         { return "" }
func (m *MockConfig) GetDatabaseMaxOpenConns() int           { return 0 }
func (m *MockConfig) GetDatabaseMaxIdleConns() int           { return 0 }
func (m *MockConfig) GetDatabaseConnMaxLifetime() int        { return 0 }
func (m *MockConfig) GetDatabaseWriteURL() string            { return "" }
func (m *MockConfig) GetDatabaseReadURL() string             { return "" }
func (m *MockConfig) GetRedisURL() string                    { return "" }
func (m *MockConfig) GetServerPort() string                  { return "" }
func (m *MockConfig) GetServerHost() string                  { return "" }
func (m *MockConfig) GetPriceRefreshCron() string            { return "" }
func (m *MockConfig) GetMockMode() bool                      { return false }
func (m *MockConfig) GetBasicAuthUsername() string           { return "" }
func (m *MockConfig) GetBasicAuthPassword() string           { return "" }
func (m *MockConfig) GetAllSettings() map[string]interface{} { return nil }

// TestLinkService_CreateLink tests the CreateLink method
func (suite *LinkServiceTestSuite) TestLinkService_CreateLink() {
	productID := uuid.New()
	campaignID := uuid.New()
	product := &model.Product{
		ID:    productID,
		Title: "Test Product",
	}
	campaign := &model.Campaign{
		ID:          campaignID,
		Name:        "Test Campaign",
		UTMCampaign: "test_campaign",
	}
	offer := &model.Offer{
		ID:                    uuid.New(),
		ProductID:             productID,
		Marketplace:           model.MarketplaceLazada,
		StoreName:             "Test Store",
		Price:                 299.00,
		MarketplaceProductURL: "https://www.lazada.co.th/products/test.html",
	}

	tests := []struct {
		name        string
		req         dto.CreateLinkRequest
		setupMock   func()
		wantErr     bool
		errContains string
	}{
		{
			name: "success with Lazada marketplace",
			req: dto.CreateLinkRequest{
				ProductID:   productID,
				CampaignID:  campaignID,
				Marketplace: "lazada",
			},
			setupMock: func() {
				suite.productRepo.On("FindByID", suite.ctx, productID).
					Return(product, nil).Once()

				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(campaign, nil).Once()

				suite.offerRepo.On("FindByProductIDAndMarketplace", suite.ctx, productID, model.MarketplaceLazada).
					Return(offer, nil).Once()

				suite.linkRepo.On("ShortCodeExists", suite.ctx, mock.AnythingOfType("string")).
					Return(false, nil).Once()

				suite.linkRepo.On("Create", suite.ctx, mock.MatchedBy(func(link *model.Link) bool {
					return link.ProductID == productID &&
						link.CampaignID == campaignID &&
						link.Marketplace == model.MarketplaceLazada &&
						link.ShortCode != "" &&
						link.TargetURL != ""
				})).Run(func(args mock.Arguments) {
					link := args.Get(1).(*model.Link)
					link.ID = uuid.New() // Set ID after creation
				}).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "error when marketplace is invalid",
			req: dto.CreateLinkRequest{
				ProductID:   productID,
				CampaignID:  campaignID,
				Marketplace: "invalid",
			},
			setupMock:   func() {},
			wantErr:     true,
			errContains: "invalid marketplace",
		},
		{
			name: "error when product not found",
			req: dto.CreateLinkRequest{
				ProductID:   productID,
				CampaignID:  campaignID,
				Marketplace: "lazada",
			},
			setupMock: func() {
				suite.productRepo.On("FindByID", suite.ctx, productID).
					Return(nil, errors.New("not found")).Once()
			},
			wantErr:     true,
			errContains: "product not found",
		},
		{
			name: "error when campaign not found",
			req: dto.CreateLinkRequest{
				ProductID:   productID,
				CampaignID:  campaignID,
				Marketplace: "lazada",
			},
			setupMock: func() {
				suite.productRepo.On("FindByID", suite.ctx, productID).
					Return(product, nil).Once()

				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(nil, errors.New("not found")).Once()
			},
			wantErr:     true,
			errContains: "campaign not found",
		},
		{
			name: "error when offer not found",
			req: dto.CreateLinkRequest{
				ProductID:   productID,
				CampaignID:  campaignID,
				Marketplace: "lazada",
			},
			setupMock: func() {
				suite.productRepo.On("FindByID", suite.ctx, productID).
					Return(product, nil).Once()

				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(campaign, nil).Once()

				suite.offerRepo.On("FindByProductIDAndMarketplace", suite.ctx, productID, model.MarketplaceLazada).
					Return(nil, errors.New("not found")).Once()
			},
			wantErr:     true,
			errContains: "offer not found",
		},
		{
			name: "error when short code generation fails",
			req: dto.CreateLinkRequest{
				ProductID:   productID,
				CampaignID:  campaignID,
				Marketplace: "lazada",
			},
			setupMock: func() {
				suite.productRepo.On("FindByID", suite.ctx, productID).
					Return(product, nil).Once()

				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(campaign, nil).Once()

				suite.offerRepo.On("FindByProductIDAndMarketplace", suite.ctx, productID, model.MarketplaceLazada).
					Return(offer, nil).Once()

				// Simulate all short codes exist (collision)
				suite.linkRepo.On("ShortCodeExists", suite.ctx, mock.AnythingOfType("string")).
					Return(true, nil).Times(10) // maxRetries
			},
			wantErr:     true,
			errContains: "failed to generate unique short code",
		},
		{
			name: "error when Create fails",
			req: dto.CreateLinkRequest{
				ProductID:   productID,
				CampaignID:  campaignID,
				Marketplace: "lazada",
			},
			setupMock: func() {
				suite.productRepo.On("FindByID", suite.ctx, productID).
					Return(product, nil).Once()

				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(campaign, nil).Once()

				suite.offerRepo.On("FindByProductIDAndMarketplace", suite.ctx, productID, model.MarketplaceLazada).
					Return(offer, nil).Once()

				suite.linkRepo.On("ShortCodeExists", suite.ctx, mock.AnythingOfType("string")).
					Return(false, nil).Once()

				suite.linkRepo.On("Create", suite.ctx, mock.AnythingOfType("*model.Link")).
					Return(errors.New("database error")).Once()
			},
			wantErr:     true,
			errContains: "failed to create link",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset mocks
			suite.linkRepo.ExpectedCalls = nil
			suite.campaignRepo.ExpectedCalls = nil
			suite.productRepo.ExpectedCalls = nil
			suite.offerRepo.ExpectedCalls = nil

			// Setup mocks
			tt.setupMock()

			// Execute
			result, err := suite.service.CreateLink(suite.ctx, tt.req)

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
				assert.NotEmpty(suite.T(), result.ShortCode)
				assert.NotEmpty(suite.T(), result.TargetURL)
				assert.Contains(suite.T(), result.FullURL, result.ShortCode)
			}
		})
	}
}

func TestLinkServiceTestSuite(t *testing.T) {
	suite.Run(t, new(LinkServiceTestSuite))
}
