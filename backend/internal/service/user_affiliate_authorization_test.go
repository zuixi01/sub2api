package service

import "testing"

func TestUserCanPromoteRequiresActiveAuthorizedUser(t *testing.T) {
	t.Parallel()

	if (&User{Status: StatusActive}).CanPromote() {
		t.Fatal("active user without authorization must not promote")
	}
	if (&User{Status: StatusDisabled, AffiliateAuthorized: true}).CanPromote() {
		t.Fatal("disabled authorized user must not promote")
	}
	if !(&User{Status: StatusActive, AffiliateAuthorized: true}).CanPromote() {
		t.Fatal("active authorized user must promote")
	}
}
