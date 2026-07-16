ALTER TABLE affiliate_visit_events
    ADD COLUMN IF NOT EXISTS registered_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS registered_at TIMESTAMPTZ NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_affiliate_visit_events_registered_user_unique
    ON affiliate_visit_events (registered_user_id)
    WHERE registered_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_affiliate_visit_events_affiliate_registered
    ON affiliate_visit_events (affiliate_user_id, registered_at DESC)
    WHERE registered_user_id IS NOT NULL;
