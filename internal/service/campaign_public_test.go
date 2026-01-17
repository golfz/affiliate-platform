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
					CampaignProducts: []model.CampaignProduct{
						{
							CampaignID: campaignID,
							ProductID:  productID,
							Product: model.Product{
								ID:       productID,
								Title:    "Test Product",
								ImageURL: "https://example.com/image.jpg",
							},
						},
					},
				}
				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(campaign, nil).Once()

				offers := []*model.Offer{
					{
						ID:                    uuid.New(),
						ProductID:             productID,
						Marketplace:           model.MarketplaceLazada,
						StoreName:             "Lazada Store",
						Price:                 299.00,
						MarketplaceProductURL: "https://www.lazada.co.th/products/test.html",
						LastCheckedAt:         time.Now(),
					},
					{
						ID:                    uuid.New(),
						ProductID:             productID,
						Marketplace:           model.MarketplaceShopee,
						StoreName:             "Shopee Store",
						Price:                 279.00,
						MarketplaceProductURL: "https://shopee.co.th/product/test",
						LastCheckedAt:         time.Now(),
					},
				}
				suite.offerRepo.On("FindByProductID", suite.ctx, productID).
					Return(offers, nil).Once()

				links := []*model.Link{
					{
						ID:          uuid.New(),
						ProductID:   productID,
						CampaignID:  campaignID,
						Marketplace: model.MarketplaceLazada,
						ShortCode:   "abc123xyz",
					},
					{
						ID:          uuid.New(),
						ProductID:   productID,
						CampaignID:  campaignID,
						Marketplace: model.MarketplaceShopee,
						ShortCode:   "def456uvw",
					},
				}
				suite.linkRepo.On("FindByProductIDAndCampaignID", suite.ctx, productID, campaignID).
					Return(links, nil).Once()
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
			name:       "error when campaign is not active (not started)",
			campaignID: campaignID,
			setupMock: func() {
				campaign := &model.Campaign{
					ID:          campaignID,
					Name:        "Test Campaign",
					UTMCampaign: "test_campaign",
					StartAt:     time.Now().Add(24 * time.Hour), // Starts tomorrow
					EndAt:       time.Now().Add(30 * 24 * time.Hour),
				}
				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(campaign, nil).Once()
			},
			wantErr:     true,
			errContains: "campaign is not active",
		},
		{
			name:       "error when campaign is not active (ended)",
			campaignID: campaignID,
			setupMock: func() {
				campaign := &model.Campaign{
					ID:          campaignID,
					Name:        "Test Campaign",
					UTMCampaign: "test_campaign",
					StartAt:     time.Now().Add(-30 * 24 * time.Hour), // Started 30 days ago
					EndAt:       time.Now().Add(-24 * time.Hour),      // Ended yesterday
				}
				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(campaign, nil).Once()
			},
			wantErr:     true,
			errContains: "campaign is not active",
		},
		{
			name:       "success when no links exist (auto-create)",
			campaignID: campaignID,
			setupMock: func() {
				campaign := &model.Campaign{
					ID:          campaignID,
					Name:        "Test Campaign",
					UTMCampaign: "test_campaign",
					StartAt:     startAt,
					EndAt:       endAt,
					CampaignProducts: []model.CampaignProduct{
						{
							CampaignID: campaignID,
							ProductID:  productID,
							Product: model.Product{
								ID:       productID,
								Title:    "Test Product",
								ImageURL: "https://example.com/image.jpg",
							},
						},
					},
				}
				suite.campaignRepo.On("FindByID", suite.ctx, campaignID).
					Return(campaign, nil).Once()

				offers := []*model.Offer{
					{
						ID:                    uuid.New(),
						ProductID:             productID,
						Marketplace:           model.MarketplaceLazada,
						StoreName:             "Lazada Store",
						Price:                 299.00,
						MarketplaceProductURL: "https://www.lazada.co.th/products/test.html",
						LastCheckedAt:         time.Now(),
					},
				}
				suite.offerRepo.On("FindByProductID", suite.ctx, productID).
					Return(offers, nil).Once()

				// No existing links
				suite.linkRepo.On("FindByProductIDAndCampaignID", suite.ctx, productID, campaignID).
					Return([]*model.Link{}, nil).Once()

				// Auto-create link calls
				suite.linkRepo.On("ShortCodeExists", suite.ctx, mock.AnythingOfType("string")).
					Return(false, nil).Once()

				suite.linkRepo.On("Create", suite.ctx, mock.MatchedBy(func(link *model.Link) bool {
					return link.ProductID == productID &&
						link.CampaignID == campaignID &&
						link.Marketplace == model.MarketplaceLazada &&
						link.ShortCode != ""
				})).Run(func(args mock.Arguments) {
					link := args.Get(1).(*model.Link)
					link.ID = uuid.New()
				}).Return(nil).Once()

				// Re-fetch links after creation
				createdLink := &model.Link{
					ID:          uuid.New(),
					ProductID:   productID,
					CampaignID:  campaignID,
					Marketplace: model.MarketplaceLazada,
					ShortCode:   "newcode123",
				}
				suite.linkRepo.On("FindByProductIDAndCampaignID", suite.ctx, productID, campaignID).
					Return([]*model.Link{createdLink}, nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset mocks
			suite.campaignRepo.ExpectedCalls = nil
			suite.productRepo.ExpectedCalls = nil
			suite.offerRepo.ExpectedCalls = nil
			suite.linkRepo.ExpectedCalls = nil

			// Setup mocks
			tt.setupMock()

			// Execute
			result, err := suite.service.GetPublicCampaign(suite.ctx, tt.campaignID)

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
				assert.Equal(suite.T(), "Test Campaign", result.Name)
			}
		})
	}
}

func TestCampaignPublicServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CampaignPublicServiceTestSuite))
}
