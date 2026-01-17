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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// DashboardServiceTestSuite is the test suite for DashboardService
type DashboardServiceTestSuite struct {
	suite.Suite
	service      *DashboardService
	clickRepo    *MockClickRepository
	linkRepo     *MockLinkRepository
	campaignRepo *MockCampaignRepository
	productRepo  *MockProductRepository
	logger       logger.Logger
	ctx          context.Context
}

func (suite *DashboardServiceTestSuite) SetupTest() {
	suite.clickRepo = new(MockClickRepository)
	suite.linkRepo = new(MockLinkRepository)
	suite.campaignRepo = new(MockCampaignRepository)
	suite.productRepo = new(MockProductRepository)
	log, err := logger.NewZapLogger("info")
	if err != nil {
		suite.T().Fatal("Failed to create logger:", err)
	}
	suite.logger = log
	suite.ctx = context.Background()
	suite.service = NewDashboardService(
		suite.clickRepo,
		suite.linkRepo,
		suite.campaignRepo,
		suite.productRepo,
		suite.logger,
	)
}

func (suite *DashboardServiceTestSuite) TearDownTest() {
	suite.clickRepo.AssertExpectations(suite.T())
	suite.linkRepo.AssertExpectations(suite.T())
	suite.campaignRepo.AssertExpectations(suite.T())
	suite.productRepo.AssertExpectations(suite.T())
}

// TestDashboardService_GetDashboardStats tests the GetDashboardStats method
func (suite *DashboardServiceTestSuite) TestDashboardService_GetDashboardStats() {
	campaignID := uuid.New()
	marketplace := "lazada"
	startDate := time.Now().Add(-7 * 24 * time.Hour)
	endDate := time.Now()

	tests := []struct {
		name        string
		params      dto.DashboardQueryParams
		setupMock   func()
		wantErr     bool
		errContains string
	}{
		{
			name: "success with all stats",
			params: dto.DashboardQueryParams{
				CampaignID:  &campaignID,
				Marketplace: &marketplace,
				StartDate:   &startDate,
				EndDate:     &endDate,
			},
			setupMock: func() {
				// Total clicks
				suite.clickRepo.On("CountWithFilters", suite.ctx, &campaignID, &marketplace, startDate, endDate).
					Return(int64(100), nil).Once()

				// Total links
				suite.linkRepo.On("CountWithFilters", suite.ctx, &campaignID, &marketplace).
					Return(int64(50), nil).Once()

				// Campaign stats
				suite.clickRepo.On("CountByCampaignWithFilters", suite.ctx, &campaignID, &marketplace, startDate, endDate).
					Return([]model.CampaignStatResult{
						{
							CampaignID:   campaignID,
							CampaignName: "Test Campaign",
							Clicks:       100,
						},
					}, nil).Once()

				// Marketplace stats
				suite.clickRepo.On("CountByMarketplaceWithFilters", suite.ctx, &campaignID, &marketplace, startDate, endDate).
					Return([]model.MarketplaceStatResult{
						{
							Marketplace: "lazada",
							Clicks:      60,
						},
						{
							Marketplace: "shopee",
							Clicks:      40,
						},
					}, nil).Once()

				// Top products
				suite.clickRepo.On("FindTopProductsWithFilters", suite.ctx, &campaignID, &marketplace, startDate, endDate, 10).
					Return([]model.TopProductResult{
						{
							ProductID:   uuid.New(),
							ProductName: "Product 1",
							Marketplace: "lazada",
							Clicks:      30,
						},
					}, nil).Once()

				// Recent clicks
				suite.clickRepo.On("FindRecentClicks", suite.ctx, 10).
					Return([]model.Click{
						{
							ID:        uuid.New(),
							Timestamp: time.Now(),
							Link: model.Link{
								ID:          uuid.New(),
								ProductID:   uuid.New(),
								CampaignID:  campaignID,
								Marketplace: model.MarketplaceLazada,
								Product: model.Product{
									ID:    uuid.New(),
									Title: "Product 1",
								},
								Campaign: model.Campaign{
									ID:   campaignID,
									Name: "Test Campaign",
								},
							},
						},
					}, nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "success with no filters",
			params: dto.DashboardQueryParams{},
			setupMock: func() {
				// Total clicks
				suite.clickRepo.On("CountWithFilters", suite.ctx, (*uuid.UUID)(nil), (*string)(nil), time.Time{}, mock.AnythingOfType("time.Time")).
					Return(int64(200), nil).Once()

				// Total links
				suite.linkRepo.On("CountWithFilters", suite.ctx, (*uuid.UUID)(nil), (*string)(nil)).
					Return(int64(100), nil).Once()

				// Campaign stats
				suite.clickRepo.On("CountByCampaignWithFilters", suite.ctx, (*uuid.UUID)(nil), (*string)(nil), time.Time{}, mock.AnythingOfType("time.Time")).
					Return([]model.CampaignStatResult{}, nil).Once()

				// Marketplace stats
				suite.clickRepo.On("CountByMarketplaceWithFilters", suite.ctx, (*uuid.UUID)(nil), (*string)(nil), time.Time{}, mock.AnythingOfType("time.Time")).
					Return([]model.MarketplaceStatResult{}, nil).Once()

				// Top products
				suite.clickRepo.On("FindTopProductsWithFilters", suite.ctx, (*uuid.UUID)(nil), (*string)(nil), time.Time{}, mock.AnythingOfType("time.Time"), 10).
					Return([]model.TopProductResult{}, nil).Once()

				// Recent clicks
				suite.clickRepo.On("FindRecentClicks", suite.ctx, 10).
					Return([]model.Click{}, nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "error when getTotalClicks fails",
			params: dto.DashboardQueryParams{},
			setupMock: func() {
				suite.clickRepo.On("CountWithFilters", suite.ctx, (*uuid.UUID)(nil), (*string)(nil), time.Time{}, mock.AnythingOfType("time.Time")).
					Return(int64(0), errors.New("database error")).Once()
			},
			wantErr:     true,
			errContains: "failed to get total clicks",
		},
		{
			name:   "error when getTotalLinks fails",
			params: dto.DashboardQueryParams{},
			setupMock: func() {
				suite.clickRepo.On("CountWithFilters", suite.ctx, (*uuid.UUID)(nil), (*string)(nil), time.Time{}, mock.AnythingOfType("time.Time")).
					Return(int64(100), nil).Once()

				suite.linkRepo.On("CountWithFilters", suite.ctx, (*uuid.UUID)(nil), (*string)(nil)).
					Return(int64(0), errors.New("database error")).Once()
			},
			wantErr:     true,
			errContains: "failed to get total links",
		},
		{
			name:   "error when getCampaignStats fails",
			params: dto.DashboardQueryParams{},
			setupMock: func() {
				suite.clickRepo.On("CountWithFilters", suite.ctx, (*uuid.UUID)(nil), (*string)(nil), time.Time{}, mock.AnythingOfType("time.Time")).
					Return(int64(100), nil).Once()

				suite.linkRepo.On("CountWithFilters", suite.ctx, (*uuid.UUID)(nil), (*string)(nil)).
					Return(int64(50), nil).Once()

				suite.clickRepo.On("CountByCampaignWithFilters", suite.ctx, (*uuid.UUID)(nil), (*string)(nil), time.Time{}, mock.AnythingOfType("time.Time")).
					Return(nil, errors.New("database error")).Once()
			},
			wantErr:     true,
			errContains: "failed to get campaign stats",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset mocks
			suite.clickRepo.ExpectedCalls = nil
			suite.linkRepo.ExpectedCalls = nil
			suite.campaignRepo.ExpectedCalls = nil
			suite.productRepo.ExpectedCalls = nil

			// Setup mocks
			tt.setupMock()

			// Execute
			result, err := suite.service.GetDashboardStats(suite.ctx, tt.params)

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
				assert.GreaterOrEqual(suite.T(), result.TotalClicks, int64(0))
				assert.GreaterOrEqual(suite.T(), result.TotalLinks, int64(0))
				assert.GreaterOrEqual(suite.T(), result.CTR, 0.0)
			}
		})
	}
}

func TestDashboardServiceTestSuite(t *testing.T) {
	suite.Run(t, new(DashboardServiceTestSuite))
}
