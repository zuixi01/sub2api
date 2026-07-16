package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type affiliateAuthorizationRepoStub struct {
	AffiliateRepository
	authorized     map[int64]bool
	profiles       map[int64]*AffiliateSummary
	profilesByCode map[string]*AffiliateSummary
	bindCalls      int
	accrueCalls    int
	setCalls       int
	visitCalls     int
}

func (r *affiliateAuthorizationRepoStub) EnsureUserAffiliate(_ context.Context, userID int64) (*AffiliateSummary, error) {
	if profile := r.profiles[userID]; profile != nil {
		copy := *profile
		return &copy, nil
	}
	return &AffiliateSummary{UserID: userID, CreatedAt: time.Now()}, nil
}

func (r *affiliateAuthorizationRepoStub) GetAffiliateByCode(_ context.Context, code string) (*AffiliateSummary, error) {
	if profile := r.profilesByCode[code]; profile != nil {
		copy := *profile
		return &copy, nil
	}
	return nil, ErrAffiliateProfileNotFound
}

func (r *affiliateAuthorizationRepoStub) BindInviter(context.Context, int64, int64) (bool, error) {
	r.bindCalls++
	return true, nil
}

func (r *affiliateAuthorizationRepoStub) AccrueQuota(context.Context, int64, int64, float64, int, *int64) (bool, error) {
	r.accrueCalls++
	return true, nil
}

func (r *affiliateAuthorizationRepoStub) GetAccruedRebateFromInvitee(context.Context, int64, int64) (float64, error) {
	return 0, nil
}

func (r *affiliateAuthorizationRepoStub) IsAffiliateAuthorized(_ context.Context, userID int64) (bool, error) {
	return r.authorized[userID], nil
}

func (r *affiliateAuthorizationRepoStub) IsAffiliateSettlementEligible(context.Context, int64) (bool, error) {
	return false, nil
}

func (r *affiliateAuthorizationRepoStub) SetAffiliateAuthorized(_ context.Context, _ int64, userID int64, authorized bool) error {
	r.authorized[userID] = authorized
	r.setCalls++
	return nil
}

func (r *affiliateAuthorizationRepoStub) RecordAffiliateVisit(context.Context, AffiliateVisitInput) (int64, bool, error) {
	r.visitCalls++
	return 1, true, nil
}

func (r *affiliateAuthorizationRepoStub) GetAffiliateGrowthMetrics(context.Context, int64) (AffiliateGrowthMetrics, error) {
	return AffiliateGrowthMetrics{}, nil
}

type affiliateAuthorizationSettingRepoStub struct {
	SettingRepository
	values map[string]string
}

func (r *affiliateAuthorizationSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	value, ok := r.values[key]
	if !ok {
		return "", ErrSettingNotFound
	}
	return value, nil
}

func newAffiliateAuthorizationService(repo *affiliateAuthorizationRepoStub) *AffiliateService {
	settings := NewSettingService(&affiliateAuthorizationSettingRepoStub{values: map[string]string{
		SettingKeyAffiliateEnabled:    "true",
		SettingKeyAffiliateRebateRate: "3",
	}}, nil)
	return NewAffiliateService(repo, settings, nil, nil)
}

func TestAffiliateAuthorizationGuardsBindingAndAccrual(t *testing.T) {
	ctx := context.Background()
	inviterID := int64(100)
	repo := &affiliateAuthorizationRepoStub{
		authorized: map[int64]bool{inviterID: false},
		profiles: map[int64]*AffiliateSummary{
			200:       {UserID: 200, CreatedAt: time.Now()},
			inviterID: {UserID: inviterID, AffCode: "INVITER", CreatedAt: time.Now()},
		},
		profilesByCode: map[string]*AffiliateSummary{
			"INVITER": {UserID: inviterID, AffCode: "INVITER", CreatedAt: time.Now()},
		},
	}
	svc := newAffiliateAuthorizationService(repo)

	err := svc.BindInviterByCode(ctx, 200, "INVITER")
	if !errors.Is(err, ErrAffiliateCodeInvalid) {
		t.Fatalf("unauthorized inviter must be rejected with invalid-code error, got %v", err)
	}
	if repo.bindCalls != 0 {
		t.Fatal("unauthorized inviter must not receive a new binding")
	}

	repo.authorized[inviterID] = true
	if err := svc.BindInviterByCode(ctx, 200, "INVITER"); err != nil {
		t.Fatalf("authorized inviter must be bindable: %v", err)
	}
	if repo.bindCalls != 1 {
		t.Fatalf("expected one binding, got %d", repo.bindCalls)
	}

	inviteeID := int64(201)
	repo.profiles[inviteeID] = &AffiliateSummary{UserID: inviteeID, InviterID: &inviterID, CreatedAt: time.Now()}
	repo.authorized[inviterID] = false
	rebate, err := svc.AccrueInviteRebateForOrder(ctx, inviteeID, 100, nil)
	if err != nil {
		t.Fatalf("revoked inviter accrual should be ignored, got %v", err)
	}
	if rebate != 0 || repo.accrueCalls != 0 {
		t.Fatalf("revoked inviter must not accrue new rebate, got rebate=%v calls=%d", rebate, repo.accrueCalls)
	}
}

func TestAdminSetAffiliateAuthorization(t *testing.T) {
	repo := &affiliateAuthorizationRepoStub{authorized: map[int64]bool{}}
	svc := newAffiliateAuthorizationService(repo)

	if err := svc.AdminSetAffiliateAuthorization(context.Background(), 1, 42, true); err != nil {
		t.Fatalf("set authorization: %v", err)
	}
	if !repo.authorized[42] || repo.setCalls != 1 {
		t.Fatal("authorization must be delegated to the repository")
	}
}

func TestRecordAffiliateVisitRequiresAuthorizedAffiliate(t *testing.T) {
	ctx := context.Background()
	affiliateID := int64(101)
	repo := &affiliateAuthorizationRepoStub{
		authorized: map[int64]bool{affiliateID: false},
		profilesByCode: map[string]*AffiliateSummary{
			"INVITER": {UserID: affiliateID, AffCode: "INVITER", CreatedAt: time.Now()},
		},
	}
	svc := newAffiliateAuthorizationService(repo)
	input := AffiliateVisitInput{AffCode: "INVITER", VisitedOn: time.Now(), VisitorHash: "a" + strings.Repeat("1", 63)}

	_, recorded, err := svc.RecordAffiliateVisit(ctx, input)
	if !errors.Is(err, ErrAffiliateCodeInvalid) {
		t.Fatalf("unauthorized affiliate must reject visit, got %v", err)
	}
	if recorded || repo.visitCalls != 0 {
		t.Fatal("unauthorized affiliate must not record a visit")
	}

	repo.authorized[affiliateID] = true
	_, recorded, err = svc.RecordAffiliateVisit(ctx, input)
	if err != nil || !recorded || repo.visitCalls != 1 {
		t.Fatalf("authorized affiliate visit must be recorded, recorded=%v err=%v calls=%d", recorded, err, repo.visitCalls)
	}
}
