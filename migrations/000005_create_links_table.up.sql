-- Affiliate Links
CREATE TABLE links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    marketplace VARCHAR(20) NOT NULL CHECK (marketplace IN ('lazada', 'shopee')),
    short_code VARCHAR(20) NOT NULL UNIQUE,
    target_url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_links_short_code ON links(short_code);
CREATE INDEX idx_links_product_campaign ON links(product_id, campaign_id);
