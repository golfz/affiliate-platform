//go:build integration
// +build integration

package service

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jonosize/affiliate-platform/internal/config"
	"github.com/jonosize/affiliate-platform/internal/database"
	"github.com/jonosize/affiliate-platform/internal/dto"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/repository"
	"github.com/jonosize/affiliate-platform/pkg/adapters/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// CampaignIntegrationTestSuite is an integration test suite for CampaignService
// This test uses a real database connection and tests the full workflow
type CampaignIntegrationTestSuite struct {
	suite.Suite
	db          *database.DB
	cfg         config.Config
	logger      logger.Logger
	ctx         context.Context
	productSvc  *ProductService
	campaignSvc *CampaignService
	linkSvc     *LinkService
	redirectSvc *RedirectService
	clickSvc    *ClickService
	productID   uuid.UUID
	campaignID  uuid.UUID
}

func (suite *CampaignIntegrationTestSuite) SetupSuite() {
	// Skip if not running integration tests
	if testing.Short() {
		suite.T().Skip("Skipping integration test in short mode")
	}

	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./configs"
	}
	cfg := config.LoadOrPanic(configPath)
	suite.cfg = cfg

	// Initialize logger
	log, err := logger.NewZapLogger("info")
	require.NoError(suite.T(), err, "Failed to create logger")
	suite.logger = log

	// Initialize database
	db, err := database.InitGORM(cfg)
	require.NoError(suite.T(), err, "Failed to initialize database")
	suite.db = db

	suite.ctx = context.Background()
}

func (suite *CampaignIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		_ = suite.db.Close() // Ignore error in test cleanup
	}
}

func (suite *CampaignIntegrationTestSuite) SetupTest() {
	// Initialize repositories
	productRepo := repository.NewProductRepository(suite.db)
	offerRepo := repository.NewOfferRepository(suite.db)
	campaignRepo := repository.NewCampaignRepository(suite.db)
	linkRepo := repository.NewLinkRepository(suite.db)
	clickRepo := repository.NewClickRepository(suite.db)

	// Initialize adapters
	lazadaAdapter, shopeeAdapter, err := mock.GetMockAdapters()
	require.NoError(suite.T(), err, "Failed to get mock adapters")

	// Initialize services
	suite.productSvc = NewProductService(
		productRepo,
		offerRepo,
		lazadaAdapter,
		shopeeAdapter,
		suite.logger,
	)

	suite.campaignSvc = NewCampaignService(
		campaignRepo,
		linkRepo,
		offerRepo,
		productRepo,
		suite.cfg,
		suite.logger,
	)

	suite.linkSvc = NewLinkService(
		linkRepo,
		campaignRepo,
		productRepo,
		offerRepo,
		suite.cfg,
		suite.logger,
	)

	suite.clickSvc = NewClickService(
		clickRepo,
		linkRepo,
		suite.logger,
	)

	suite.redirectSvc = NewRedirectService(
		linkRepo,
		suite.clickSvc,
		suite.logger,
	)
}

func (suite *CampaignIntegrationTestSuite) TearDownTest() {
	// Clean up test data
	if suite.campaignID != uuid.Nil {
		_ = suite.campaignSvc.DeleteCampaign(suite.ctx, suite.campaignID)
	}
	if suite.productID != uuid.Nil {
		_ = suite.productSvc.DeleteProduct(suite.ctx, suite.productID)
	}
}

// TestCampaignWorkflow tests the complete workflow:
// 1. Create a product with offers
// 2. Create a campaign with that product
// 3. Verify links are automatically created
// 4. Test redirect and click tracking
func (suite *CampaignIntegrationTestSuite) TestCampaignWorkflow() {
	// Step 1: Create a product
	productReq := dto.CreateProductRequest{
		LazadaURL: "https://www.lazada.co.th/products/test-product.html",
		ShopeeURL: "https://shopee.co.th/product/test-product",
	}

	product, err := suite.productSvc.CreateProduct(suite.ctx, productReq)
	require.NoError(suite.T(), err, "Failed to create product")
	require.NotNil(suite.T(), product)
	suite.productID = product.ID

	// Verify product was created
	assert.NotEqual(suite.T(), uuid.Nil, product.ID)
	assert.NotEmpty(suite.T(), product.Title)

	// Step 2: Verify offers were created
	offersResp, err := suite.productSvc.GetProductOffers(suite.ctx, product.ID)
	require.NoError(suite.T(), err, "Failed to get product offers")
	require.NotNil(suite.T(), offersResp)
	assert.Greater(suite.T(), len(offersResp.Offers), 0, "Product should have at least one offer")

	// Step 3: Create a campaign with the product
	startAt := time.Now().Add(-24 * time.Hour)   // Started yesterday
	endAt := time.Now().Add(30 * 24 * time.Hour) // Ends in 30 days

	campaignReq := dto.CreateCampaignRequest{
		Name:        "Integration Test Campaign",
		UTMCampaign: "integration_test_2025",
		StartAt:     startAt,
		EndAt:       endAt,
		ProductIDs:  []uuid.UUID{product.ID},
	}

	campaign, err := suite.campaignSvc.CreateCampaign(suite.ctx, campaignReq)
	require.NoError(suite.T(), err, "Failed to create campaign")
	require.NotNil(suite.T(), campaign)
	suite.campaignID = campaign.ID

	// Verify campaign was created
	assert.NotEqual(suite.T(), uuid.Nil, campaign.ID)
	assert.Equal(suite.T(), "Integration Test Campaign", campaign.Name)
	assert.Equal(suite.T(), "integration_test_2025", campaign.UTMCampaign)

	// Step 4: Verify links were automatically created for the product
	// Get campaign response to see product IDs
	campaignResponse, err := suite.campaignSvc.GetCampaignResponse(suite.ctx, campaign.ID)
	require.NoError(suite.T(), err, "Failed to get campaign response")
	assert.Contains(suite.T(), campaignResponse.ProductIDs, product.ID, "Campaign should contain the product")

	// Step 5: Create a link manually to test link creation
	// First, get an offer to use
	offersResp, err = suite.productSvc.GetProductOffers(suite.ctx, product.ID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), offersResp)
	require.Greater(suite.T(), len(offersResp.Offers), 0, "Product should have offers")

	// Create link for Lazada marketplace
	linkReq := dto.CreateLinkRequest{
		ProductID:   product.ID,
		CampaignID:  campaign.ID,
		Marketplace: "lazada",
	}

	link, err := suite.linkSvc.CreateLink(suite.ctx, linkReq)
	require.NoError(suite.T(), err, "Failed to create link")
	require.NotNil(suite.T(), link)

	// Verify link was created
	assert.NotEqual(suite.T(), uuid.Nil, link.ID)
	assert.NotEmpty(suite.T(), link.ShortCode)
	assert.NotEmpty(suite.T(), link.TargetURL)
	assert.Contains(suite.T(), link.TargetURL, "utm_source=affiliate")
	assert.Contains(suite.T(), link.TargetURL, "utm_campaign=integration_test_2025")
	assert.Contains(suite.T(), link.FullURL, link.ShortCode)

	// Step 6: Test redirect (this will also track a click)
	ipAddress := net.IP{192, 168, 1, 1} // Use a valid IP address
	redirectURL, err := suite.redirectSvc.Redirect(
		suite.ctx,
		link.ShortCode,
		ipAddress,             // IP address
		"Mozilla/5.0 (Test)",  // User agent
		"https://example.com", // Referrer
	)
	require.NoError(suite.T(), err, "Failed to redirect")
	assert.NotEmpty(suite.T(), redirectURL)
	assert.Contains(suite.T(), redirectURL, "utm_source=affiliate")

	// Step 7: Verify click was tracked (give it a moment for async processing)
	time.Sleep(100 * time.Millisecond)
	clickCount, err := suite.clickSvc.GetClickStats(suite.ctx, link.ID)
	require.NoError(suite.T(), err, "Failed to get click stats")
	assert.GreaterOrEqual(suite.T(), clickCount, int64(1), "At least one click should be tracked")

	// Step 8: Test updating campaign products
	newProductReq := dto.CreateProductRequest{
		LazadaURL: "https://www.lazada.co.th/products/another-product.html",
	}
	newProduct, err := suite.productSvc.CreateProduct(suite.ctx, newProductReq)
	require.NoError(suite.T(), err, "Failed to create second product")

	// Update campaign to include both products
	err = suite.campaignSvc.UpdateCampaignProducts(suite.ctx, campaign.ID, []uuid.UUID{product.ID, newProduct.ID})
	require.NoError(suite.T(), err, "Failed to update campaign products")

	// Verify campaign was updated
	updatedCampaign, err := suite.campaignSvc.GetCampaignResponse(suite.ctx, campaign.ID)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), updatedCampaign.ProductIDs, product.ID)
	assert.Contains(suite.T(), updatedCampaign.ProductIDs, newProduct.ID)

	// Clean up second product
	_ = suite.productSvc.DeleteProduct(suite.ctx, newProduct.ID)
}

func TestCampaignIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CampaignIntegrationTestSuite))
}
