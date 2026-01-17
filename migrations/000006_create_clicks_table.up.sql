-- Clicks (tracking)
CREATE TABLE clicks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    timestamp TIMESTAMP DEFAULT NOW(),
    referrer TEXT,
    user_agent TEXT,
    ip_address INET,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_clicks_link_id ON clicks(link_id);
CREATE INDEX idx_clicks_timestamp ON clicks(timestamp DESC);
CREATE INDEX idx_clicks_link_timestamp ON clicks(link_id, timestamp DESC);
