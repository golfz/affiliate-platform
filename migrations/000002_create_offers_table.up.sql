-- Offers (prices from marketplaces)
CREATE TABLE offers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    marketplace VARCHAR(20) NOT NULL CHECK (marketplace IN ('lazada', 'shopee')),
    store_name VARCHAR(200),
    price DECIMAL(10, 2) NOT NULL CHECK (price >= 0),
    marketplace_product_url TEXT NOT NULL,
    last_checked_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(product_id, marketplace)
);

CREATE INDEX idx_offers_product_id ON offers(product_id);
CREATE INDEX idx_offers_marketplace ON offers(marketplace);
CREATE INDEX idx_offers_last_checked_at ON offers(last_checked_at);
