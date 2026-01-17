package service

import (
	"context"
	"errors"
	"testing"
	"time"

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

// MockLinkRepository is a mock implementation of LinkRepositoryInterface
type MockLinkRepository struct {
	mock.Mock
}

func (m *MockLinkRepository) Create(ctx context.Context, link *model.Link) error {
	args := m.Called(ctx, link)
	return args.Error(0)
}

func (m *MockLinkRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Link, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Link), args.Error(1)
}

func (m *MockLinkRepository) FindByShortCode(ctx context.Context, shortCode string) (*model.Link, error) {
	args := m.Called(ctx, shortCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Link), args.Error(1)
}

func (m *MockLinkRepository) FindByProductIDAndCampaignID(ctx context.Context, productID, campaignID uuid.UUID) ([]*model.Link, error) {
	args := m.Called(ctx, productID, campaignID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Link), args.Error(1)
}

func (m *MockLinkRepository) FindByCampaignID(ctx context.Context, campaignID uuid.UUID) ([]*model.Link, error) {
	args := m.Called(ctx, campaignID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Link), args.Error(1)
}

func (m *MockLinkRepository) Update(ctx context.Context, link *model.Link) error {
	args := m.Called(ctx, link)
	return args.Error(0)
}

func (m *MockLinkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockLinkRepository) DeleteByProductIDAndCampaignID(ctx context.Context, productID, campaignID uuid.UUID) error {
	args := m.Called(ctx, productID, campaignID)
	return args.Error(0)
}

func (m *MockLinkRepository) DeleteByCampaignIDAndNotInProducts(ctx context.Context, campaignID uuid.UUID, productIDs []uuid.UUID) error {
	args := m.Called(ctx, campaignID, productIDs)
	return args.Error(0)
}

func (m *MockLinkRepository) ShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	args := m.Called(ctx, shortCode)
	return args.Bool(0), args.Error(1)
}

func (m *MockLinkRepository) CountWithFilters(ctx context.Context, campaignID *uuid.UUID, marketplace *string) (int64, error) {
	args := m.Called(ctx, campaignID, marketplace)
	return args.Get(0).(int64), args.Error(1)
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

// MockConfig is a mock implementation of config.Config
type MockConfig struct {
	apiBaseURL string
}

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
func (m *MockConfig) GetAPIBaseURL() string                  { return m.apiBaseURL }
func (m *MockConfig) GetPriceRefreshCron() string            { return "" }
func (m *MockConfig) GetMockMode() bool                      { return false }
func (m *MockConfig) GetBasicAuthUsername() string           { return "" }
func (m *MockConfig) GetBasicAuthPassword() string           { return "" }
func (m *MockConfig) GetAllSettings() map[string]interface{} { return nil }

// CampaignServiceTestSuite is the test suite for CampaignService
type CampaignServiceTestSuite struct {
	suite.Suite
	service      *CampaignService
	campaignRepo *MockCampaignRepository
	linkRepo     *MockLinkRepository
	offerRepo    *MockOfferRepository
	productRepo  *MockProductRepository
	cfg          config.Config
	logger       logger.Logger
	ctx          context.Context
}

func (suite *CampaignServiceTestSuite) SetupTest() {
	suite.campaignRepo = new(MockCampaignRepository)
	suite.linkRepo = new(MockLinkRepository)
	suite.offerRepo = new(MockOfferRepository)
	suite.productRepo = new(MockProductRepository)
	suite.cfg = &MockConfig{apiBaseURL: "https://api.example.com"}
	log, err := logger.NewZapLogger("info")
	if err != nil {
		suite.T().Fatal("Failed to create logger:", err)
	}
	suite.logger = log
	suite.ctx = context.Background()
	suite.service = NewCampaignService(
		suite.campaignRepo,
		suite.linkRepo,
		suite.offerRepo,
		suite.productRepo,
		suite.cfg,
		suite.logger,
	)
}

func (suite *CampaignServiceTestSuite) TearDownTest() {
	suite.campaignRepo.AssertExpectations(suite.T())
	suite.linkRepo.AssertExpectations(suite.T())
	suite.offerRepo.AssertExpectations(suite.T())
	suite.productRepo.AssertExpectations(suite.T())
}

// TestCampaignService_CreateCampaign tests the CreateCampaign method
func (suite *CampaignServiceTestSuite) TestCampaignService_CreateCampaign() {
	startAt := time.Now().Add(24 * time.Hour)
	endAt := time.Now().Add(30 * 24 * time.Hour)

	tests := []struct {
		name        string
		req         dto.CreateCampaignRequest
		setupMock   func()
		wantErr     bool
		errContains string
	}{
		{
			name: "success without products",
			req: dto.CreateCampaignRequest{
				Name:        "Test Campaign",
				UTMCampaign: "test_campaign",
				StartAt:     startAt,
				EndAt:       endAt,
				ProductIDs:  []uuid.UUID{},
			},
			setupMock: func() {
				campaign := &model.Campaign{
					ID:          uuid.New(),
					Name:        "Test Campaign",
					UTMCampaign: "test_campaign",
					StartAt:     startAt,
					EndAt:       endAt,
					CreatedAt:   time.Now(),
				}
				suite.campaignRepo.On("Create", suite.ctx, mock.MatchedBy(func(c *model.Campaign) bool {
					return c.Name == "Test Campaign" &&
						c.UTMCampaign == "test_campaign" &&
						c.StartAt.Equal(startAt) &&
						c.EndAt.Equal(endAt)
				})).Run(func(args mock.Arguments) {
					c := args.Get(1).(*model.Campaign)
					c.ID = campaign.ID
					c.CreatedAt = campaign.CreatedAt
				}).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "success with products",
			req: dto.CreateCampaignRequest{
				Name:        "Test Campaign",
				UTMCampaign: "test_campaign",
				StartAt:     startAt,
				EndAt:       endAt,
				ProductIDs:  []uuid.UUID{uuid.New()},
			},
			setupMock: func() {
				campaignID := uuid.New()
				productID := []uuid.UUID{uuid.New()}[0] // Get first product ID from the test case

				suite.campaignRepo.On("Create", suite.ctx, mock.AnythingOfType("*model.Campaign")).
					Run(func(args mock.Arguments) {
						c := args.Get(1).(*model.Campaign)
						c.ID = campaignID
						c.CreatedAt = time.Now()
					}).Return(nil).Once()

				suite.campaignRepo.On("AddProducts", suite.ctx, campaignID, []uuid.UUID{productID}).
					Return(nil).Once()

				// Mock createLinksForProducts calls
				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(&model.Campaign{
						ID:          campaignID,
						UTMCampaign: "test_campaign",
					}, nil).Once()

				suite.offerRepo.On("FindByProductID", suite.ctx, productID).
					Return([]*model.Offer{}, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "error when end_at is before start_at",
			req: dto.CreateCampaignRequest{
				Name:        "Test Campaign",
				UTMCampaign: "test_campaign",
				StartAt:     endAt,
				EndAt:       startAt, // Reversed
				ProductIDs:  []uuid.UUID{},
			},
			setupMock:   func() {},
			wantErr:     true,
			errContains: "end_at must be after start_at",
		},
		{
			name: "error when end_at equals start_at",
			req: dto.CreateCampaignRequest{
				Name:        "Test Campaign",
				UTMCampaign: "test_campaign",
				StartAt:     startAt,
				EndAt:       startAt, // Same time
				ProductIDs:  []uuid.UUID{},
			},
			setupMock:   func() {},
			wantErr:     true,
			errContains: "end_at must be after start_at",
		},
		{
			name: "error when UTM campaign is too long",
			req: dto.CreateCampaignRequest{
				Name:        "Test Campaign",
				UTMCampaign: string(make([]byte, 101)), // 101 characters
				StartAt:     startAt,
				EndAt:       endAt,
				ProductIDs:  []uuid.UUID{},
			},
			setupMock:   func() {},
			wantErr:     true,
			errContains: "utm_campaign must be 100 characters or less",
		},
		{
			name: "error when Create fails",
			req: dto.CreateCampaignRequest{
				Name:        "Test Campaign",
				UTMCampaign: "test_campaign",
				StartAt:     startAt,
				EndAt:       endAt,
				ProductIDs:  []uuid.UUID{},
			},
			setupMock: func() {
				suite.campaignRepo.On("Create", suite.ctx, mock.AnythingOfType("*model.Campaign")).
					Return(errors.New("database error")).Once()
			},
			wantErr:     true,
			errContains: "failed to create campaign",
		},
		{
			name: "error when AddProducts fails",
			req: dto.CreateCampaignRequest{
				Name:        "Test Campaign",
				UTMCampaign: "test_campaign",
				StartAt:     startAt,
				EndAt:       endAt,
				ProductIDs:  []uuid.UUID{uuid.New()},
			},
			setupMock: func() {
				campaignID := uuid.New()
				suite.campaignRepo.On("Create", suite.ctx, mock.AnythingOfType("*model.Campaign")).
					Run(func(args mock.Arguments) {
						c := args.Get(1).(*model.Campaign)
						c.ID = campaignID
					}).Return(nil).Once()

				suite.campaignRepo.On("AddProducts", suite.ctx, campaignID, mock.AnythingOfType("[]uuid.UUID")).
					Return(errors.New("database error")).Once()

				suite.campaignRepo.On("Delete", suite.ctx, campaignID).
					Return(nil).Once() // Rollback
			},
			wantErr:     true,
			errContains: "failed to add products to campaign",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset mocks
			suite.campaignRepo.ExpectedCalls = nil
			suite.linkRepo.ExpectedCalls = nil
			suite.offerRepo.ExpectedCalls = nil
			suite.productRepo.ExpectedCalls = nil

			// Capture productID from request for use in setupMock
			var productID uuid.UUID
			if len(tt.req.ProductIDs) > 0 {
				productID = tt.req.ProductIDs[0]
			}

			// Setup mocks - update setupMock to use captured productID
			if tt.name == "success with products" {
				// Override setupMock for this specific test
				campaignID := uuid.New()
				suite.campaignRepo.On("Create", suite.ctx, mock.AnythingOfType("*model.Campaign")).
					Run(func(args mock.Arguments) {
						c := args.Get(1).(*model.Campaign)
						c.ID = campaignID
						c.CreatedAt = time.Now()
					}).Return(nil).Once()

				suite.campaignRepo.On("AddProducts", suite.ctx, campaignID, mock.MatchedBy(func(ids []uuid.UUID) bool {
					return len(ids) == 1 && ids[0] == productID
				})).Return(nil).Once()

				// Mock createLinksForProducts calls
				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(&model.Campaign{
						ID:          campaignID,
						UTMCampaign: "test_campaign",
					}, nil).Once()

				// Mock DeleteByCampaignIDAndNotInProducts (called by createLinksForProducts)
				suite.linkRepo.On("DeleteByCampaignIDAndNotInProducts", suite.ctx, campaignID, mock.AnythingOfType("[]uuid.UUID")).
					Return(nil).Once()

				suite.offerRepo.On("FindByProductID", suite.ctx, productID).
					Return([]*model.Offer{}, nil).Once()

				// Mock FindByProductIDAndCampaignID (called by createLinksForProducts - called twice: once before deletion, once after)
				suite.linkRepo.On("FindByProductIDAndCampaignID", suite.ctx, productID, campaignID).
					Return([]*model.Link{}, nil).Twice()
			} else {
				tt.setupMock()
			}

			// Execute
			result, err := suite.service.CreateCampaign(suite.ctx, tt.req)

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
				assert.Equal(suite.T(), tt.req.Name, result.Name)
				assert.Equal(suite.T(), tt.req.UTMCampaign, result.UTMCampaign)
			}
		})
	}
}

// TestCampaignService_GetCampaignResponse tests the GetCampaignResponse method
func (suite *CampaignServiceTestSuite) TestCampaignService_GetCampaignResponse() {
	campaignID := uuid.New()
	productID := uuid.New()

	tests := []struct {
		name        string
		campaignID  uuid.UUID
		setupMock   func()
		wantErr     bool
		errContains string
	}{
		{
			name:       "success",
			campaignID: campaignID,
			setupMock: func() {
				campaign := &model.Campaign{
					ID:          campaignID,
					Name:        "Test Campaign",
					UTMCampaign: "test_campaign",
					StartAt:     time.Now(),
					EndAt:       time.Now().Add(24 * time.Hour),
					CreatedAt:   time.Now(),
					CampaignProducts: []model.CampaignProduct{
						{
							CampaignID: campaignID,
							ProductID:  productID,
						},
					},
				}
				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(campaign, nil).Once()
			},
			wantErr: false,
		},
		{
			name:       "error when campaign not found",
			campaignID: campaignID,
			setupMock: func() {
				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(nil, errors.New("not found")).Once()
			},
			wantErr:     true,
			errContains: "campaign not found",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset mocks
			suite.campaignRepo.ExpectedCalls = nil

			// Setup mocks
			tt.setupMock()

			// Execute
			result, err := suite.service.GetCampaignResponse(suite.ctx, tt.campaignID)

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
				assert.Equal(suite.T(), campaignID, result.ID)
			}
		})
	}
}

// TestCampaignService_GetAllCampaigns tests the GetAllCampaigns method
func (suite *CampaignServiceTestSuite) TestCampaignService_GetAllCampaigns() {
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
				campaigns := []*model.Campaign{
					{
						ID:          uuid.New(),
						Name:        "Campaign 1",
						UTMCampaign: "campaign_1",
						StartAt:     time.Now(),
						EndAt:       time.Now().Add(24 * time.Hour),
						CreatedAt:   time.Now(),
					},
					{
						ID:          uuid.New(),
						Name:        "Campaign 2",
						UTMCampaign: "campaign_2",
						StartAt:     time.Now(),
						EndAt:       time.Now().Add(24 * time.Hour),
						CreatedAt:   time.Now(),
					},
				}
				suite.campaignRepo.On("FindAll", suite.ctx, 10, 0).
					Return(campaigns, int64(2), nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "error when FindAll fails",
			limit:  10,
			offset: 0,
			setupMock: func() {
				suite.campaignRepo.On("FindAll", suite.ctx, 10, 0).
					Return(nil, int64(0), errors.New("database error")).Once()
			},
			wantErr:     true,
			errContains: "failed to get campaigns",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset mocks
			suite.campaignRepo.ExpectedCalls = nil

			// Setup mocks
			tt.setupMock()

			// Execute
			result, err := suite.service.GetAllCampaigns(suite.ctx, tt.limit, tt.offset)

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

// TestCampaignService_DeleteCampaign tests the DeleteCampaign method
func (suite *CampaignServiceTestSuite) TestCampaignService_DeleteCampaign() {
	campaignID := uuid.New()
	campaign := &model.Campaign{
		ID:   campaignID,
		Name: "Test Campaign",
	}

	tests := []struct {
		name        string
		campaignID  uuid.UUID
		setupMock   func()
		wantErr     bool
		errContains string
	}{
		{
			name:       "success",
			campaignID: campaignID,
			setupMock: func() {
				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(campaign, nil).Once()

				suite.campaignRepo.On("Delete", suite.ctx, campaignID).
					Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:       "error when campaign not found",
			campaignID: campaignID,
			setupMock: func() {
				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(nil, errors.New("not found")).Once()
			},
			wantErr:     true,
			errContains: "campaign not found",
		},
		{
			name:       "error when Delete fails",
			campaignID: campaignID,
			setupMock: func() {
				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(campaign, nil).Once()

				suite.campaignRepo.On("Delete", suite.ctx, campaignID).
					Return(errors.New("database error")).Once()
			},
			wantErr:     true,
			errContains: "failed to delete campaign",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset mocks
			suite.campaignRepo.ExpectedCalls = nil

			// Setup mocks
			tt.setupMock()

			// Execute
			err := suite.service.DeleteCampaign(suite.ctx, tt.campaignID)

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

// TestCampaignService_UpdateCampaign tests the UpdateCampaign method
func (suite *CampaignServiceTestSuite) TestCampaignService_UpdateCampaign() {
	campaignID := uuid.New()
	startAt := time.Now().Add(24 * time.Hour)
	endAt := time.Now().Add(30 * 24 * time.Hour)
	newName := "Updated Campaign"

	tests := []struct {
		name        string
		campaignID  uuid.UUID
		req         dto.UpdateCampaignRequest
		setupMock   func()
		wantErr     bool
		errContains string
	}{
		{
			name:       "success updating name",
			campaignID: campaignID,
			req: dto.UpdateCampaignRequest{
				Name: newName,
			},
			setupMock: func() {
				campaign := &model.Campaign{
					ID:          campaignID,
					Name:        "Original Campaign",
					UTMCampaign: "original_campaign",
					StartAt:     startAt,
					EndAt:       endAt,
				}
				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(campaign, nil).Once().
					Return(campaign, nil).Once() // Called again after update

				updatedCampaign := *campaign
				updatedCampaign.Name = newName
				suite.campaignRepo.On("Update", suite.ctx, mock.MatchedBy(func(c *model.Campaign) bool {
					return c.ID == campaignID && c.Name == newName
				})).Return(nil).Once()

				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(&updatedCampaign, nil).Once()
			},
			wantErr: false,
		},
		{
			name:       "error when campaign not found",
			campaignID: campaignID,
			req: dto.UpdateCampaignRequest{
				Name: newName,
			},
			setupMock: func() {
				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(nil, errors.New("not found")).Once()
			},
			wantErr:     true,
			errContains: "campaign not found",
		},
		{
			name:       "error when dates are invalid",
			campaignID: campaignID,
			req: dto.UpdateCampaignRequest{
				StartAt: &endAt,
				EndAt:   &startAt, // Reversed
			},
			setupMock: func() {
				campaign := &model.Campaign{
					ID:          campaignID,
					Name:        "Test Campaign",
					UTMCampaign: "test_campaign",
					StartAt:     startAt,
					EndAt:       endAt,
				}
				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(campaign, nil).Once()
			},
			wantErr:     true,
			errContains: "end_at must be after start_at",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset mocks
			suite.campaignRepo.ExpectedCalls = nil
			suite.linkRepo.ExpectedCalls = nil
			suite.offerRepo.ExpectedCalls = nil

			// Setup mocks
			tt.setupMock()

			// Execute
			result, err := suite.service.UpdateCampaign(suite.ctx, tt.campaignID, tt.req)

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
				assert.Equal(suite.T(), campaignID, result.ID)
			}
		})
	}
}

func TestCampaignServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CampaignServiceTestSuite))
}
