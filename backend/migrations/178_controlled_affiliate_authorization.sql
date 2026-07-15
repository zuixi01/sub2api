ALTER TABLE users
    ADD COLUMN IF NOT EXISTS affiliate_authorized BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_users_affiliate_authorized
    ON users (affiliate_authorized)
    WHERE affiliate_authorized = TRUE AND deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS affiliate_authorization_audits (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_admin_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    authorized BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_affiliate_authorization_audits_user_created
    ON affiliate_authorization_audits (user_id, created_at DESC);

UPDATE user_affiliates
SET aff_rebate_rate_percent = NULL
WHERE aff_rebate_rate_percent IS NOT NULL;

INSERT INTO settings (key, value, updated_at)
VALUES
    ('affiliate_rebate_rate', '3', NOW()),
    ('affiliate_rebate_freeze_hours', '168', NOW()),
    ('affiliate_rebate_duration_days', '365', NOW()),
    ('affiliate_admin_recharge_enabled', 'false', NOW())
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    updated_at = NOW();
