package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAffiliateGrowthMigrationCreatesPrivacySafeVisitEvents(t *testing.T) {
	content, err := FS.ReadFile("179_add_affiliate_growth_events.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS affiliate_visit_events")
	require.Contains(t, sql, "affiliate_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE")
	require.Contains(t, sql, "visited_on DATE NOT NULL")
	require.Contains(t, sql, "visitor_hash CHAR(64) NOT NULL")
	require.Contains(t, sql, "UNIQUE (affiliate_user_id, visited_on, visitor_hash)")
	require.NotContains(t, strings.ToLower(sql), "ip_address")
	require.NotContains(t, strings.ToLower(sql), "user_agent")
}

func TestAffiliateVerifiedAttributionMigrationLinksRegistrationWithoutRawIdentifiers(t *testing.T) {
	content, err := FS.ReadFile("180_affiliate_verified_attribution.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "registered_user_id BIGINT")
	require.Contains(t, sql, "REFERENCES users(id) ON DELETE SET NULL")
	require.Contains(t, sql, "registered_at TIMESTAMPTZ")
	require.Contains(t, sql, "UNIQUE INDEX")
	require.NotContains(t, strings.ToLower(sql), "ip_address")
	require.NotContains(t, strings.ToLower(sql), "user_agent")
}
