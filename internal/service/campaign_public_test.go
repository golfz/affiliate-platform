package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jonosize/affiliate-platform/internal/config"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// CampaignPublicServiceTestSuite is the test suite for CampaignPublicService
type CampaignPublicServiceTestSuite struct {
	suite.Suite
	service      *CampaignPublicService
	campaignRepo *MockCampaignRepository
	productRepo  *MockProductRepository
	offerRepo    *MockOfferRepository
	linkRepo     *MockLinkRepository
	cfg          config.Config
	logger       logger.Logger
	ctx          context.Context
}

func (suite *CampaignPublicServiceTestSuite) SetupTest() {
	suite.campaignRepo = new(MockCampaignRepository)
	suite.productRepo = new(MockProductRepository)
	suite.offerRepo = new(MockOfferRepository)
	suite.linkRepo = new(MockLinkRepository)
	suite.cfg = &MockConfig{apiBaseURL: "https://api.example.com"}
	log, err := logger.NewZapLogger("info")
	if err != nil {
		suite.T().Fatal("Failed to create logger:", err)
	}
	suite.logger = log
	suite.ctx = context.Background()
	suite.service = NewCampaignPublicService(
		suite.campaignRepo,
		suite.productRepo,
		suite.offerRepo,
		suite.linkRepo,
		suite.cfg,
		suite.logger,
	)
}

func (suite *CampaignPublicServiceTestSuite) TearDownTest() {
	suite.campaignRepo.AssertExpectations(suite.T())
	suite.productRepo.AssertExpectations(suite.T())
	suite.offerRepo.AssertExpectations(suite.T())
	suite.linkRepo.AssertExpectations(suite.T())
}

// TestCampaignPublicService_GetPublicCampaign tests the GetPublicCampaign method
func (suite *CampaignPublicServiceTestSuite) TestCampaignPublicService_GetPublicCampaign() {
	campaignID := uuid.New()
	productID := uuid.New()
	startAt := time.Now().Add(-24 * time.Hour) // Started yesterday
	endAt := time.Now().Add(24 * time.Hour)    // Ends tomorrow

	tests := []struct {
		name        string
		campaignID  uuid.UUID
		setupMock   func()
		wantErr     bool
		errContains string
	}{
		{
			name:       "success with active campaign",
			campaignID: campaignID,
			setupMock: func() {
				campaign := &model.Campaign{
					ID:          campaignID,
					Name:        "Test Campaign",
					UTMCampaign: "test_campaign",
					StartAt:     startAt,
					EndAt:       endAt,
				}
				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).Return(campaign, nil)

				product := &model.Product{
					ID:    productID,
					Title: "Test Product",
				}
				suite.productRepo.On("FindByID", suite.ctx, productID).Return(product, nil)

				offers := []*model.Offer{
					{
						ID:                    uuid.New(),
						ProductID:             productID,
						Marketplace:           model.MarketplaceLazada,
						Price:                 100.0,
						MarketplaceProductURL: "https://lazada.com/product",
					},
					{
						ID:                    uuid.New(),
						ProductID:             productID,
						Marketplace:           model.MarketplaceShopee,
						Price:                 95.0,
						MarketplaceProductURL: "https://shopee.com/product",
					},
				}
				suite.offerRepo.On("FindByProductID", suite.ctx, productID).Return(offers, nil)

				links := []*model.Link{
					{
						ID:          uuid.New(),
						ProductID:   productID,
						CampaignID:  campaignID,
						ShortCode:   "abc123",
						Marketplace: model.MarketplaceLazada,
						TargetURL:   "https://lazada.com/product?utm_source=test&utm_medium=affiliate&utm_campaign=test_campaign",
					},
					{
						ID:          uuid.New(),
						ProductID:   productID,
						CampaignID:  campaignID,
						ShortCode:   "def456",
						Marketplace: model.MarketplaceShopee,
						TargetURL:   "https://shopee.com/product?utm_source=test&utm_medium=affiliate&utm_campaign=test_campaign",
					},
				}
				suite.linkRepo.On("FindByProductIDAndCampaignID", suite.ctx, productID, campaignID).Return(links, nil)
			},
			wantErr: false,
		},
		{
			name:       "campaign not found",
			campaignID: campaignID,
			setupMock: func() {
				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).Return(nil, errors.New("campaign not found"))
			},
			wantErr:     true,
			errContains: "campaign not found",
		},
		{
			name:       "campaign not active - not started",
			campaignID: campaignID,
			setupMock: func() {
				campaign := &model.Campaign{
					ID:          campaignID,
					Name:        "Test Campaign",
					UTMCampaign: "test_campaign",
					StartAt:     time.Now().Add(24 * time.Hour), // Starts tomorrow
					EndAt:       time.Now().Add(48 * time.Hour),
				}
				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).Return(campaign, nil)
			},
			wantErr:     true,
			errContains: "campaign is not currently active",
		},
		{
			name:       "campaign not active - ended",
			campaignID: campaignID,
			setupMock: func() {
				campaign := &model.Campaign{
					ID:          campaignID,
					Name:        "Test Campaign",
					UTMCampaign: "test_campaign",
					StartAt:     time.Now().Add(-48 * time.Hour), // Started 2 days ago
					EndAt:       time.Now().Add(-24 * time.Hour), // Ended yesterday
				}
				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).Return(campaign, nil)
			},
			wantErr:     true,
			errContains: "campaign is not currently active",
		},
		{
			name:       "success with auto-created links",
			campaignID: campaignID,
			setupMock: func() {
				campaign := &model.Campaign{
					ID:          campaignID,
					Name:        "Test Campaign",
					UTMCampaign: "test_campaign",
					StartAt:     startAt,
					EndAt:       endAt,
				}
				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).Return(campaign, nil)

				product := &model.Product{
					ID:    productID,
					Title: "Test Product",
				}
				suite.productRepo.On("FindByID", suite.ctx, productID).Return(product, nil)

				offers := []*model.Offer{
					{
						ID:                    uuid.New(),
						ProductID:             productID,
						Marketplace:           model.MarketplaceLazada,
						Price:                 100.0,
						MarketplaceProductURL: "https://lazada.com/product",
					},
					{
						ID:                    uuid.New(),
						ProductID:             productID,
						Marketplace:           model.MarketplaceShopee,
						Price:                 95.0,
						MarketplaceProductURL: "https://shopee.com/product",
					},
				}
				suite.offerRepo.On("FindByProductID", suite.ctx, productID).Return(offers, nil)

				// No links exist yet
				suite.linkRepo.On("FindByProductIDAndCampaignID", suite.ctx, productID, campaignID).Return([]*model.Link{}, nil)

				// Mock ShortCodeExists to return false (short code doesn't exist)
				suite.linkRepo.On("ShortCodeExists", suite.ctx, mock.AnythingOfType("string")).Return(false, nil).Maybe()

				// Mock Create for auto-created links
				suite.linkRepo.On("Create", suite.ctx, mock.AnythingOfType("*model.Link")).Return(nil).Maybe()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.setupMock()
			result, err := suite.service.GetPublicCampaign(suite.ctx, tt.campaignID)

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

func TestCampaignPublicServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CampaignPublicServiceTestSuite))
}
