-- Campaign Products (many-to-many)
CREATE TABLE campaign_products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(campaign_id, product_id)
);

CREATE INDEX idx_campaign_products_campaign ON campaign_products(campaign_id);
CREATE INDEX idx_campaign_products_product ON campaign_products(product_id);
