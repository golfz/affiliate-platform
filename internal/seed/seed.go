package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jonosize/affiliate-platform/internal/config"
	"github.com/jonosize/affiliate-platform/internal/database"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/model"
)

// SeedDatabase seeds the database with development/demo data
func SeedDatabase(db *database.DB, cfg config.Config, log logger.Logger) error {
	ctx := context.Background()

	log.Info("Starting database seeding...")

	// Seed products
	products, err := seedProducts(ctx, db, log)
	if err != nil {
		return fmt.Errorf("failed to seed products: %w", err)
	}
	log.Info("Products seeded", logger.Int("count", len(products)))

	// Seed offers
	offers, err := seedOffers(ctx, db, products, log)
	if err != nil {
		return fmt.Errorf("failed to seed offers: %w", err)
	}
	log.Info("Offers seeded", logger.Int("count", len(offers)))

	// Seed campaigns
	campaigns, err := seedCampaigns(ctx, db, log)
	if err != nil {
		return fmt.Errorf("failed to seed campaigns: %w", err)
	}
	log.Info("Campaigns seeded", logger.Int("count", len(campaigns)))

	// Seed campaign products
	if err := seedCampaignProducts(ctx, db, campaigns, products, log); err != nil {
		return fmt.Errorf("failed to seed campaign products: %w", err)
	}
	log.Info("Campaign products seeded")

	// Seed links
	links, err := seedLinks(ctx, db, cfg, campaigns, products, offers, log)
	if err != nil {
		return fmt.Errorf("failed to seed links: %w", err)
	}
	log.Info("Links seeded", logger.Int("count", len(links)))

	log.Info("Database seeding completed successfully")
	return nil
}

// seedProducts creates sample products
func seedProducts(ctx context.Context, db *database.DB, log logger.Logger) ([]*model.Product, error) {
	products := []*model.Product{
		{
			ID:       uuid.New(),
			Title:    "Premium Matcha Powder 100g",
			ImageURL: "https://via.placeholder.com/400",
		},
		{
			ID:       uuid.New(),
			Title:    "Organic Green Tea Leaves 200g",
			ImageURL: "https://via.placeholder.com/400",
		},
		{
			ID:       uuid.New(),
			Title:    "Jasmine Tea 150g",
			ImageURL: "https://via.placeholder.com/400",
		},
		{
			ID:       uuid.New(),
			Title:    "Oolong Tea 250g",
			ImageURL: "https://via.placeholder.com/400",
		},
		{
			ID:       uuid.New(),
			Title:    "White Tea 100g",
			ImageURL: "https://via.placeholder.com/400",
		},
	}

	for _, product := range products {
		if err := db.Write.WithContext(ctx).Create(product).Error; err != nil {
			// If product already exists (by ID), skip
			if err.Error() != "pq: duplicate key value violates unique constraint \"products_pkey\"" {
				return nil, err
			}
			log.Info("Product already exists, skipping", logger.String("id", product.ID.String()))
		}
	}

	return products, nil
}

// seedOffers creates sample offers for products
func seedOffers(ctx context.Context, db *database.DB, products []*model.Product, log logger.Logger) ([]*model.Offer, error) {
	if len(products) == 0 {
		return nil, fmt.Errorf("no products to create offers for")
	}

	offers := []*model.Offer{}

	// Create offers for each product (both Lazada and Shopee)
	for _, product := range products {
		// Lazada offer
		lazadaOffer := &model.Offer{
			ID:                    uuid.New(),
			ProductID:             product.ID,
			Marketplace:           model.MarketplaceLazada,
			StoreName:             "Lazada Store",
			Price:                 299.00 + float64(len(offers)*10), // Varying prices
			MarketplaceProductURL: fmt.Sprintf("https://www.lazada.co.th/products/product-%s", product.ID.String()[:8]),
			LastCheckedAt:         time.Now(),
		}
		offers = append(offers, lazadaOffer)

		// Shopee offer (slightly cheaper for demo)
		shopeeOffer := &model.Offer{
			ID:                    uuid.New(),
			ProductID:             product.ID,
			Marketplace:           model.MarketplaceShopee,
			StoreName:             "Shopee Store",
			Price:                 279.00 + float64(len(offers)*10), // Varying prices
			MarketplaceProductURL: fmt.Sprintf("https://shopee.co.th/product-%s", product.ID.String()[:8]),
			LastCheckedAt:         time.Now(),
		}
		offers = append(offers, shopeeOffer)
	}

	for _, offer := range offers {
		// Use Upsert to avoid duplicate key errors
		err := db.Write.WithContext(ctx).
			Where("product_id = ? AND marketplace = ?", offer.ProductID, offer.Marketplace).
			Assign(*offer).
			FirstOrCreate(offer).Error
		if err != nil {
			return nil, err
		}
	}

	return offers, nil
}

// seedCampaigns creates sample campaigns
func seedCampaigns(ctx context.Context, db *database.DB, log logger.Logger) ([]*model.Campaign, error) {
	now := time.Now()
	campaigns := []*model.Campaign{
		{
			ID:          uuid.New(),
			Name:        "Summer Sale 2025",
			UTMCampaign: "summer_2025",
			StartAt:     now.AddDate(0, 0, -7), // Started 7 days ago
			EndAt:       now.AddDate(0, 1, 0),  // Ends in 1 month
		},
		{
			ID:          uuid.New(),
			Name:        "Winter Special",
			UTMCampaign: "winter_2025",
			StartAt:     now.AddDate(0, 2, 0), // Starts in 2 months
			EndAt:       now.AddDate(0, 3, 0), // Ends in 3 months
		},
		{
			ID:          uuid.New(),
			Name:        "Flash Sale - Limited Time",
			UTMCampaign: "flash_sale_2025",
			StartAt:     now.AddDate(0, 0, -1), // Started 1 day ago
			EndAt:       now.AddDate(0, 0, 2),  // Ends in 2 days
		},
	}

	for _, campaign := range campaigns {
		if err := db.Write.WithContext(ctx).Create(campaign).Error; err != nil {
			if err.Error() != "pq: duplicate key value violates unique constraint \"campaigns_pkey\"" {
				return nil, err
			}
			log.Info("Campaign already exists, skipping", logger.String("id", campaign.ID.String()))
		}
	}

	return campaigns, nil
}

// seedCampaignProducts associates products with campaigns
func seedCampaignProducts(ctx context.Context, db *database.DB, campaigns []*model.Campaign, products []*model.Product, log logger.Logger) error {
	if len(campaigns) == 0 || len(products) == 0 {
		return nil
	}

	// First campaign: All products
	campaign1 := campaigns[0]
	for _, product := range products {
		campaignProduct := &model.CampaignProduct{
			ID:         uuid.New(),
			CampaignID: campaign1.ID,
			ProductID:  product.ID,
		}
		err := db.Write.WithContext(ctx).
			Where("campaign_id = ? AND product_id = ?", campaignProduct.CampaignID, campaignProduct.ProductID).
			FirstOrCreate(campaignProduct).Error
		if err != nil {
			return err
		}
	}

	// Second campaign: First 3 products
	if len(campaigns) > 1 && len(products) >= 3 {
		campaign2 := campaigns[1]
		for i := 0; i < 3 && i < len(products); i++ {
			campaignProduct := &model.CampaignProduct{
				ID:         uuid.New(),
				CampaignID: campaign2.ID,
				ProductID:  products[i].ID,
			}
			err := db.Write.WithContext(ctx).
				Where("campaign_id = ? AND product_id = ?", campaignProduct.CampaignID, campaignProduct.ProductID).
				FirstOrCreate(campaignProduct).Error
			if err != nil {
				return err
			}
		}
	}

	// Third campaign: Last 2 products
	if len(campaigns) > 2 && len(products) >= 2 {
		campaign3 := campaigns[2]
		for i := len(products) - 2; i < len(products); i++ {
			campaignProduct := &model.CampaignProduct{
				ID:         uuid.New(),
				CampaignID: campaign3.ID,
				ProductID:  products[i].ID,
			}
			err := db.Write.WithContext(ctx).
				Where("campaign_id = ? AND product_id = ?", campaignProduct.CampaignID, campaignProduct.ProductID).
				FirstOrCreate(campaignProduct).Error
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// seedLinks creates sample affiliate links
func seedLinks(ctx context.Context, db *database.DB, cfg config.Config, campaigns []*model.Campaign, products []*model.Product, offers []*model.Offer, log logger.Logger) ([]*model.Link, error) {
	if len(campaigns) == 0 || len(products) == 0 || len(offers) == 0 {
		return nil, fmt.Errorf("missing data to create links")
	}

	links := []*model.Link{}

	// Create links for first active campaign
	activeCampaign := campaigns[0]
	for i, product := range products {
		if i >= len(offers) {
			break
		}

		// Get offers for this product
		productOffers := []*model.Offer{}
		for _, offer := range offers {
			if offer.ProductID == product.ID {
				productOffers = append(productOffers, offer)
			}
		}

		// Create link for each marketplace
		for _, offer := range productOffers {
			// Generate short code (simple for seeding)
			shortCode := fmt.Sprintf("demo%02d%s", i, offer.Marketplace[:2])

			// Build target URL with UTM
			baseURL := offer.MarketplaceProductURL
			utmSource := "affiliate"
			utmMedium := "affiliate"
			utmCampaign := activeCampaign.UTMCampaign
			targetURL := fmt.Sprintf("%s?utm_source=%s&utm_medium=%s&utm_campaign=%s",
				baseURL, utmSource, utmMedium, utmCampaign)

			link := &model.Link{
				ID:          uuid.New(),
				ProductID:   product.ID,
				CampaignID:  activeCampaign.ID,
				Marketplace: offer.Marketplace,
				ShortCode:   shortCode,
				TargetURL:   targetURL,
			}

			// Use Upsert to avoid duplicate key errors
			err := db.Write.WithContext(ctx).
				Where("product_id = ? AND campaign_id = ? AND marketplace = ?", link.ProductID, link.CampaignID, link.Marketplace).
				Assign(*link).
				FirstOrCreate(link).Error
			if err != nil {
				return nil, err
			}

			links = append(links, link)
		}
	}

	return links, nil
}
