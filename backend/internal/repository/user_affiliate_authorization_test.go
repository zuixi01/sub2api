package repository

import (
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
)

func TestUserEntityToServiceMapsAffiliateAuthorization(t *testing.T) {
	t.Parallel()

	mapped := userEntityToService(&dbent.User{AffiliateAuthorized: true})
	if mapped == nil || !mapped.AffiliateAuthorized {
		t.Fatal("affiliate authorization must be mapped from persistence")
	}
}
