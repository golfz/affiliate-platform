package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/jonosize/affiliate-platform/internal/config"
	"github.com/jonosize/affiliate-platform/internal/database"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/model"
	"github.com/jonosize/affiliate-platform/internal/repository"
	"github.com/jonosize/affiliate-platform/pkg/adapters"
	"github.com/jonosize/affiliate-platform/pkg/adapters/mock"
	"github.com/robfig/cron/v3"
)

// PriceRefreshWorker handles periodic price refresh
type PriceRefreshWorker struct {
	cron        *cron.Cron
	db          *database.DB
	cfg         config.Config
	logger      logger.Logger
	offerRepo   *repository.OfferRepository
	productRepo *repository.ProductRepository
}

// NewPriceRefreshWorker creates a new price refresh worker
func NewPriceRefreshWorker(db *database.DB, cfg config.Config, log logger.Logger) *PriceRefreshWorker {
	// Create cron with seconds precision for local timezone
	// Using WithSeconds() means cron expression needs 6 fields: second minute hour day month weekday
	c := cron.New(cron.WithSeconds(), cron.WithLocation(time.Local))

	return &PriceRefreshWorker{
		cron:        c,
		db:          db,
		cfg:         cfg,
		logger:      log,
		offerRepo:   repository.NewOfferRepository(db),
		productRepo: repository.NewProductRepository(db),
	}
}

// Start starts the cron scheduler
func (w *PriceRefreshWorker) Start() error {
	cronExpr := w.cfg.GetPriceRefreshCron()
	if cronExpr == "" {
		cronExpr = "0 0 */6 * * *" // Default: every 6 hours (6-field format: second minute hour day month weekday)
	}

	_, err := w.cron.AddFunc(cronExpr, w.refreshPrices)
	if err != nil {
		return fmt.Errorf("failed to schedule price refresh job: %w", err)
	}

	w.cron.Start()
	w.logger.Info("Price refresh worker started", logger.String("cron", cronExpr))
	return nil
}

// Stop stops the cron scheduler gracefully
func (w *PriceRefreshWorker) Stop() {
	ctx := w.cron.Stop()
	w.logger.Info("Stopping price refresh worker...")
	<-ctx.Done()
	w.logger.Info("Price refresh worker stopped")
}

// refreshPrices refreshes prices for all offers
func (w *PriceRefreshWorker) refreshPrices() {
	ctx := context.Background()
	w.logger.Info("Starting price refresh job...")

	// Get all products with offers
	products, _, err := w.productRepo.FindAll(ctx, 1000, 0) // Get up to 1000 products
	if err != nil {
		w.logger.Error("Failed to fetch products for price refresh", logger.Error(err))
		return
	}

	w.logger.Info("Refreshing prices", logger.Int("product_count", len(products)))

	// Get adapters
	lazadaAdapter, shopeeAdapter, err := mock.GetMockAdapters()
	if err != nil {
		w.logger.Error("Failed to get adapters", logger.Error(err))
		return
	}

	refreshedCount := 0
	errorCount := 0

	for _, product := range products {
		// Get offers for this product
		offers, err := w.offerRepo.FindByProductID(ctx, product.ID)
		if err != nil {
			w.logger.Error("Failed to get offers for product", logger.Error(err), logger.String("product_id", product.ID.String()))
			errorCount++
			continue
		}

		for _, offer := range offers {
			// Select adapter based on marketplace
			var adapter adapters.MarketplaceAdapter
			if offer.Marketplace == model.MarketplaceLazada {
				adapter = lazadaAdapter
			} else if offer.Marketplace == model.MarketplaceShopee {
				adapter = shopeeAdapter
			} else {
				continue
			}

			// Fetch current offer/price
			offerData, err := adapter.FetchOffer(ctx, offer.MarketplaceProductURL)
			if err != nil {
				w.logger.Error("Failed to fetch offer", logger.Error(err),
					logger.String("product_id", product.ID.String()),
					logger.String("marketplace", string(offer.Marketplace)))
				errorCount++
				continue
			}

			// Update offer with new price
			offer.Price = offerData.Price
			offer.StoreName = offerData.StoreName
			offer.LastCheckedAt = time.Now()
			offer.MarketplaceProductURL = offerData.MarketplaceProductURL

			// Save updated offer
			if err := w.offerRepo.Upsert(ctx, offer); err != nil {
				w.logger.Error("Failed to update offer", logger.Error(err),
					logger.String("product_id", product.ID.String()),
					logger.String("marketplace", string(offer.Marketplace)))
				errorCount++
				continue
			}

			refreshedCount++
		}
	}

	w.logger.Info("Price refresh job completed",
		logger.Int("refreshed", refreshedCount),
		logger.Int("errors", errorCount))
}

// TriggerManualRefresh manually triggers a price refresh (for testing/admin)
func (w *PriceRefreshWorker) TriggerManualRefresh() error {
	w.logger.Info("Manual price refresh triggered")
	w.refreshPrices()
	return nil
}
