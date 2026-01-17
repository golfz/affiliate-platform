package service

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// RedirectServiceTestSuite is the test suite for RedirectService
type RedirectServiceTestSuite struct {
	suite.Suite
	service   *RedirectService
	linkRepo  *MockLinkRepository
	clickSvc  *ClickService
	clickRepo *MockClickRepository
	logger    logger.Logger
	ctx       context.Context
}

func (suite *RedirectServiceTestSuite) SetupTest() {
	suite.linkRepo = new(MockLinkRepository)
	suite.clickRepo = new(MockClickRepository)
	log, err := logger.NewZapLogger("info")
	if err != nil {
		suite.T().Fatal("Failed to create logger:", err)
	}
	suite.logger = log
	suite.clickSvc = NewClickService(suite.clickRepo, suite.linkRepo, suite.logger)
	suite.service = NewRedirectService(suite.linkRepo, suite.clickSvc, suite.logger)
	suite.ctx = context.Background()
}

func (suite *RedirectServiceTestSuite) TearDownTest() {
	suite.linkRepo.AssertExpectations(suite.T())
	// Note: clickSvc.TrackClick is called in a goroutine, so we can't assert expectations
	// The click tracking is fire-and-forget
	// We also don't assert clickRepo expectations since TrackClick runs asynchronously
}

// TestRedirectService_Redirect tests the Redirect method
func (suite *RedirectServiceTestSuite) TestRedirectService_Redirect() {
	linkID := uuid.New()
	shortCode := "abc123xyz"
	ipAddress := net.ParseIP("192.168.1.1")
	userAgent := "Mozilla/5.0"
	referrer := "https://example.com"

	link := &model.Link{
		ID:          linkID,
		ShortCode:   shortCode,
		TargetURL:   "https://www.lazada.co.th/products/test.html",
		ProductID:   uuid.New(),
		CampaignID:  uuid.New(),
		Marketplace: model.MarketplaceLazada,
	}

	tests := []struct {
		name        string
		shortCode   string
		ipAddress   net.IP
		userAgent   string
		referrer    string
		setupMock   func()
		wantErr     bool
		errContains string
		wantURL     string
	}{
		{
			name:      "success with valid link",
			shortCode: shortCode,
			ipAddress: ipAddress,
			userAgent: userAgent,
			referrer:  referrer,
			setupMock: func() {
				suite.linkRepo.On("FindByShortCode", suite.ctx, shortCode).
					Return(link, nil).Once()

				// TrackClick is called in a goroutine, so we can't assert it
				// But we can verify the link was found
				// Mock clickRepo.Create to avoid panic (even though it's async)
				suite.clickRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Click")).
					Return(nil).Maybe() // Use Maybe() since it may or may not be called
			},
			wantErr: false,
			wantURL: "https://www.lazada.co.th/products/test.html",
		},
		{
			name:      "success with nil IP address",
			shortCode: shortCode,
			ipAddress: nil,
			userAgent: userAgent,
			referrer:  referrer,
			setupMock: func() {
				suite.linkRepo.On("FindByShortCode", suite.ctx, shortCode).
					Return(link, nil).Once()

				// Mock clickRepo.Create to avoid panic (even though it's async)
				suite.clickRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Click")).
					Return(nil).Maybe() // Use Maybe() since it may or may not be called
			},
			wantErr: false,
			wantURL: "https://www.lazada.co.th/products/test.html",
		},
		{
			name:      "error when link not found",
			shortCode: shortCode,
			ipAddress: ipAddress,
			userAgent: userAgent,
			referrer:  referrer,
			setupMock: func() {
				suite.linkRepo.On("FindByShortCode", suite.ctx, shortCode).
					Return(nil, errors.New("not found")).Once()
				// No clickRepo mock needed since link not found, so TrackClick won't be called
			},
			wantErr:     true,
			errContains: "link not found",
		},
		{
			name:      "error when redirect URL is invalid",
			shortCode: shortCode,
			ipAddress: ipAddress,
			userAgent: userAgent,
			referrer:  referrer,
			setupMock: func() {
				invalidLink := &model.Link{
					ID:          linkID,
					ShortCode:   shortCode,
					TargetURL:   "http://invalid-domain.com/product", // Not whitelisted
					ProductID:   uuid.New(),
					CampaignID:  uuid.New(),
					Marketplace: model.MarketplaceLazada,
				}
				suite.linkRepo.On("FindByShortCode", suite.ctx, shortCode).
					Return(invalidLink, nil).Once()
				// No clickRepo mock needed since URL is invalid, so TrackClick won't be called
			},
			wantErr:     true,
			errContains: "invalid redirect URL",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset mocks
			suite.linkRepo.ExpectedCalls = nil

			// Setup mocks
			tt.setupMock()

			// Execute
			result, err := suite.service.Redirect(suite.ctx, tt.shortCode, tt.ipAddress, tt.userAgent, tt.referrer)

			// Assert
			if tt.wantErr {
				assert.Error(suite.T(), err)
				if tt.errContains != "" {
					assert.Contains(suite.T(), err.Error(), tt.errContains)
				}
				assert.Empty(suite.T(), result)
			} else {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), tt.wantURL, result)

				// Give goroutine time to execute (for click tracking)
				time.Sleep(10 * time.Millisecond)
			}
		})
	}
}

func TestRedirectServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RedirectServiceTestSuite))
}
