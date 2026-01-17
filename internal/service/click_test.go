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

// MockClickRepository is a mock implementation of ClickRepositoryInterface
type MockClickRepository struct {
	mock.Mock
}

func (m *MockClickRepository) Create(ctx context.Context, click *model.Click) error {
	args := m.Called(ctx, click)
	return args.Error(0)
}

func (m *MockClickRepository) CountByLinkID(ctx context.Context, linkID uuid.UUID) (int64, error) {
	args := m.Called(ctx, linkID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockClickRepository) CountByLinkIDAndTimeRange(ctx context.Context, linkID uuid.UUID, startAt, endAt time.Time) (int64, error) {
	args := m.Called(ctx, linkID, startAt, endAt)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockClickRepository) CountByCampaignID(ctx context.Context, campaignID uuid.UUID) (int64, error) {
	args := m.Called(ctx, campaignID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockClickRepository) FindRecentClicks(ctx context.Context, limit int) ([]model.Click, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Click), args.Error(1)
}

func (m *MockClickRepository) CountWithFilters(ctx context.Context, campaignID *uuid.UUID, marketplace *string, startDate, endDate time.Time) (int64, error) {
	args := m.Called(ctx, campaignID, marketplace, startDate, endDate)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockClickRepository) CountByCampaignWithFilters(ctx context.Context, campaignID *uuid.UUID, marketplace *string, startDate, endDate time.Time) ([]model.CampaignStatResult, error) {
	args := m.Called(ctx, campaignID, marketplace, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.CampaignStatResult), args.Error(1)
}

func (m *MockClickRepository) CountByMarketplaceWithFilters(ctx context.Context, campaignID *uuid.UUID, marketplace *string, startDate, endDate time.Time) ([]model.MarketplaceStatResult, error) {
	args := m.Called(ctx, campaignID, marketplace, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.MarketplaceStatResult), args.Error(1)
}

func (m *MockClickRepository) FindTopProductsWithFilters(ctx context.Context, campaignID *uuid.UUID, marketplace *string, startDate, endDate time.Time, limit int) ([]model.TopProductResult, error) {
	args := m.Called(ctx, campaignID, marketplace, startDate, endDate, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.TopProductResult), args.Error(1)
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

func (m *MockLinkRepository) ShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	args := m.Called(ctx, shortCode)
	return args.Bool(0), args.Error(1)
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

func (m *MockLinkRepository) CountWithFilters(ctx context.Context, campaignID *uuid.UUID, marketplace *string) (int64, error) {
	args := m.Called(ctx, campaignID, marketplace)
	return args.Get(0).(int64), args.Error(1)
}

// ClickServiceTestSuite is the test suite for ClickService
type ClickServiceTestSuite struct {
	suite.Suite
	service   *ClickService
	clickRepo *MockClickRepository
	linkRepo  *MockLinkRepository
	logger    logger.Logger
	ctx       context.Context
}

func (suite *ClickServiceTestSuite) SetupTest() {
	suite.clickRepo = new(MockClickRepository)
	suite.linkRepo = new(MockLinkRepository)
	log, err := logger.NewZapLogger("info")
	if err != nil {
		suite.T().Fatal("Failed to create logger:", err)
	}
	suite.logger = log
	suite.ctx = context.Background()
	suite.service = NewClickService(suite.clickRepo, suite.linkRepo, suite.logger)
}

func (suite *ClickServiceTestSuite) TearDownTest() {
	suite.clickRepo.AssertExpectations(suite.T())
	suite.linkRepo.AssertExpectations(suite.T())
}

// TestClickService_TrackClick tests the TrackClick method
func (suite *ClickServiceTestSuite) TestClickService_TrackClick() {
	linkID := uuid.New()
	ipAddress := net.ParseIP("192.168.1.1")
	userAgent := "Mozilla/5.0"
	referrer := "https://example.com"

	tests := []struct {
		name        string
		linkID      uuid.UUID
		ipAddress   net.IP
		userAgent   string
		referrer    string
		setupMock   func()
		wantErr     bool
		errContains string
	}{
		{
			name:      "success with IP address",
			linkID:    linkID,
			ipAddress: ipAddress,
			userAgent: userAgent,
			referrer:  referrer,
			setupMock: func() {
				suite.clickRepo.On("Create", suite.ctx, mock.MatchedBy(func(click *model.Click) bool {
					return click.LinkID == linkID &&
						click.IPAddress == ipAddress.String() &&
						click.UserAgent == userAgent &&
						click.Referrer == referrer
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:      "success with nil IP address",
			linkID:    linkID,
			ipAddress: nil,
			userAgent: userAgent,
			referrer:  referrer,
			setupMock: func() {
				suite.clickRepo.On("Create", suite.ctx, mock.MatchedBy(func(click *model.Click) bool {
					return click.LinkID == linkID &&
						click.IPAddress == "" &&
						click.UserAgent == userAgent &&
						click.Referrer == referrer
				})).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:      "error when Create fails",
			linkID:    linkID,
			ipAddress: ipAddress,
			userAgent: userAgent,
			referrer:  referrer,
			setupMock: func() {
				suite.clickRepo.On("Create", suite.ctx, mock.AnythingOfType("*model.Click")).
					Return(errors.New("database error")).Once()
			},
			wantErr:     true,
			errContains: "failed to track click",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset mocks
			suite.clickRepo.ExpectedCalls = nil

			// Setup mocks
			tt.setupMock()

			// Execute
			err := suite.service.TrackClick(suite.ctx, tt.linkID, tt.ipAddress, tt.userAgent, tt.referrer)

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

// TestClickService_GetClickStats tests the GetClickStats method
func (suite *ClickServiceTestSuite) TestClickService_GetClickStats() {
	linkID := uuid.New()

	tests := []struct {
		name        string
		linkID      uuid.UUID
		setupMock   func()
		wantCount   int64
		wantErr     bool
		errContains string
	}{
		{
			name:   "success",
			linkID: linkID,
			setupMock: func() {
				suite.clickRepo.On("CountByLinkID", suite.ctx, linkID).
					Return(int64(42), nil).Once()
			},
			wantCount: 42,
			wantErr:   false,
		},
		{
			name:   "error when CountByLinkID fails",
			linkID: linkID,
			setupMock: func() {
				suite.clickRepo.On("CountByLinkID", suite.ctx, linkID).
					Return(int64(0), errors.New("database error")).Once()
			},
			wantCount:   0,
			wantErr:     true,
			errContains: "database error",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset mocks
			suite.clickRepo.ExpectedCalls = nil

			// Setup mocks
			tt.setupMock()

			// Execute
			count, err := suite.service.GetClickStats(suite.ctx, tt.linkID)

			// Assert
			if tt.wantErr {
				assert.Error(suite.T(), err)
				if tt.errContains != "" {
					assert.Contains(suite.T(), err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(suite.T(), err)
				assert.Equal(suite.T(), tt.wantCount, count)
			}
		})
	}
}

func TestClickServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ClickServiceTestSuite))
}
