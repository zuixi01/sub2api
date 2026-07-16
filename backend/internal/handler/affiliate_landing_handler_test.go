package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type landingAffiliateRepoStub struct {
	service.AffiliateRepository
	recorded service.AffiliateVisitInput
}

func (r *landingAffiliateRepoStub) GetAffiliateByCode(context.Context, string) (*service.AffiliateSummary, error) {
	return &service.AffiliateSummary{UserID: 77, AffCode: "INVITER"}, nil
}

func (r *landingAffiliateRepoStub) IsAffiliateAuthorized(context.Context, int64) (bool, error) {
	return true, nil
}

func (r *landingAffiliateRepoStub) RecordAffiliateVisit(_ context.Context, input service.AffiliateVisitInput) (int64, bool, error) {
	r.recorded = input
	return 901, true, nil
}

type landingSettingRepoStub struct{ service.SettingRepository }

func (landingSettingRepoStub) GetValue(context.Context, string) (string, error) { return "true", nil }

func TestAffiliateVisitorHashIsDailyAndNotRawIdentifier(t *testing.T) {
	day := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	first := affiliateVisitorHash("tracking-secret", day, "203.0.113.9", "example-agent")
	second := affiliateVisitorHash("tracking-secret", day, "203.0.113.9", "example-agent")
	nextDay := affiliateVisitorHash("tracking-secret", day.AddDate(0, 0, 1), "203.0.113.9", "example-agent")

	if len(first) != 64 || first != second || first == nextDay {
		t.Fatalf("expected a deterministic daily SHA-256 hash, got %q, %q, %q", first, second, nextDay)
	}
}

func TestSafeUTMAllowsOnlyCampaignSafeCharacters(t *testing.T) {
	if got := safeUTM("summer_2026-A.1"); got != "summer_2026-A.1" {
		t.Fatalf("safe campaign was changed: %q", got)
	}
	if got := safeUTM("<script>alert(1)</script>"); got != "" {
		t.Fatalf("unsafe campaign must be discarded, got %q", got)
	}
}

func TestAffiliateLandingRedirectsAndPersistsOnlyHashedVisitor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &landingAffiliateRepoStub{}
	settings := service.NewSettingService(landingSettingRepoStub{}, nil)
	serviceUnderTest := service.NewAffiliateService(repo, settings, nil, nil)
	h := &AffiliateLandingHandler{affiliateService: serviceUnderTest, trackingSecret: "tracking-secret"}
	router := gin.New()
	require.NoError(t, router.SetTrustedProxies(nil))
	router.GET("/r/:affCode", h.Redirect)

	req := httptest.NewRequest(http.MethodGet, "/r/inviter?utm_campaign=summer_2026", nil)
	req.RemoteAddr = "203.0.113.9:4321"
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Forwarded-For", "198.51.100.99")
	req.Header.Set("CF-Connecting-IP", "198.51.100.99")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	if response.Code != http.StatusFound {
		t.Fatalf("expected redirect, got %d", response.Code)
	}
	if location := response.Header().Get("Location"); location != "/register?aff_code=INVITER&utm_campaign=summer_2026" {
		t.Fatalf("unexpected redirect location: %q", location)
	}
	if repo.recorded.AffiliateUserID != 77 || len(repo.recorded.VisitorHash) != 64 || repo.recorded.VisitorHash == "203.0.113.9" {
		t.Fatalf("expected only a privacy-safe visitor hash, got %#v", repo.recorded)
	}
	require.Equal(t,
		affiliateVisitorHash("tracking-secret", repo.recorded.VisitedOn, "203.0.113.9", "test-agent"),
		repo.recorded.VisitorHash,
	)
	cookies := response.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, affiliateAttributionCookieName, cookies[0].Name)
	require.True(t, cookies[0].HttpOnly)
	require.Equal(t, http.SameSiteLaxMode, cookies[0].SameSite)
}

func TestAffiliateAttributionTokenRejectsTamperingAndExpiry(t *testing.T) {
	now := time.Date(2026, 7, 16, 8, 0, 0, 0, time.UTC)
	token := signAffiliateAttribution("tracking-secret", affiliateAttribution{
		VisitID:   901,
		AffCode:   "INVITER",
		ExpiresAt: now.Add(24 * time.Hour),
	})

	got, ok := verifyAffiliateAttribution("tracking-secret", token, now)
	require.True(t, ok)
	require.Equal(t, int64(901), got.VisitID)
	require.Equal(t, "INVITER", got.AffCode)

	_, ok = verifyAffiliateAttribution("tracking-secret", token+"tampered", now)
	require.False(t, ok)
	_, ok = verifyAffiliateAttribution("tracking-secret", token, now.Add(25*time.Hour))
	require.False(t, ok)
}

func TestAffiliateAttributionMiddlewareAddsVerifiedVisitToRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	token := signAffiliateAttribution("tracking-secret", affiliateAttribution{
		VisitID: 901, AffCode: "INVITER", ExpiresAt: time.Now().UTC().Add(time.Hour),
	})
	router := gin.New()
	router.Use(AffiliateAttributionMiddleware("tracking-secret"))
	router.GET("/register", func(c *gin.Context) {
		attribution, ok := service.AffiliateAttributionFromContext(c.Request.Context())
		require.True(t, ok)
		require.Equal(t, int64(901), attribution.VisitID)
		require.Equal(t, "INVITER", attribution.AffCode)
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/register", nil)
	request.AddCookie(&http.Cookie{Name: affiliateAttributionCookieName, Value: token})
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusNoContent, recorder.Code)
}
