package api

import (
	"github.com/jonosize/affiliate-platform/internal/api/handlers"
	"github.com/jonosize/affiliate-platform/internal/config"
	"github.com/jonosize/affiliate-platform/internal/database"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/service"
	"github.com/jonosize/affiliate-platform/internal/worker"
	"github.com/labstack/echo/v4"
)

// SetupRoutes configures all API routes
func SetupRoutes(e *echo.Echo, db *database.DB, cfg config.Config, log logger.Logger, priceRefreshWorker *worker.PriceRefreshWorker) {
	// Initialize services
	productService := service.NewProductService(db, log)
	campaignService := service.NewCampaignService(db, cfg, log)
	linkService := service.NewLinkService(db, cfg, log)
	redirectService := service.NewRedirectService(db, log)
	campaignPublicService := service.NewCampaignPublicService(db, cfg, log)
	dashboardService := service.NewDashboardService(db, log)

	// Initialize handlers
	productHandler := handlers.NewProductHandler(productService, log)
	campaignHandler := handlers.NewCampaignHandler(campaignService, log)
	linkHandler := handlers.NewLinkHandler(linkService, log)
	redirectHandler := handlers.NewRedirectHandler(redirectService, log)
	campaignPublicHandler := handlers.NewCampaignPublicHandler(campaignPublicService, log)
	workerHandler := handlers.NewWorkerHandler(priceRefreshWorker, log)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService, log)

	// Admin routes (no auth)
	adminGroup := e.Group("/api")
	{
		// Products
		adminGroup.GET("/products", productHandler.GetAllProducts)
		adminGroup.POST("/products", productHandler.CreateProduct)
		adminGroup.GET("/products/:id/offers", productHandler.GetProductOffers)
		adminGroup.DELETE("/products/:id", productHandler.DeleteProduct)

		// Campaigns
		adminGroup.GET("/campaigns", campaignHandler.GetAllCampaigns)
		adminGroup.GET("/campaigns/:id", campaignHandler.GetCampaign)
		adminGroup.POST("/campaigns", campaignHandler.CreateCampaign)
		adminGroup.PATCH("/campaigns/:id", campaignHandler.UpdateCampaign)
		adminGroup.PATCH("/campaigns/:id/products", campaignHandler.UpdateCampaignProducts)
		adminGroup.DELETE("/campaigns/:id", campaignHandler.DeleteCampaign)

		// Links
		adminGroup.POST("/links", linkHandler.CreateLink)

		// Worker
		adminGroup.POST("/worker/refresh-prices", workerHandler.TriggerPriceRefresh)

		// Dashboard
		adminGroup.GET("/dashboard", dashboardHandler.GetDashboardStats)
	}

	// Public routes (no auth)
	publicGroup := e.Group("/api")
	{
		// Public campaign endpoint
		publicGroup.GET("/campaigns/:id/public", campaignPublicHandler.GetPublicCampaign)
	}

	// Public redirect route (no group, direct route)
	e.GET("/go/:short_code", redirectHandler.Redirect)
}
