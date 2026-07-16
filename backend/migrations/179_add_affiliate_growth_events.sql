CREATE TABLE IF NOT EXISTS affiliate_visit_events (
    id BIGSERIAL PRIMARY KEY,
    affiliate_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    aff_code VARCHAR(32) NOT NULL,
    visited_on DATE NOT NULL,
    visitor_hash CHAR(64) NOT NULL,
    utm_source VARCHAR(100) NOT NULL DEFAULT '',
    utm_medium VARCHAR(100) NOT NULL DEFAULT '',
    utm_campaign VARCHAR(100) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (affiliate_user_id, visited_on, visitor_hash)
);

CREATE INDEX IF NOT EXISTS idx_affiliate_visit_events_affiliate_day
    ON affiliate_visit_events (affiliate_user_id, visited_on DESC);
