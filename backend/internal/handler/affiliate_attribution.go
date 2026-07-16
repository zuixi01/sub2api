package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	affiliateAttributionCookieName = "sub2api_aff_ref"
	affiliateAttributionTTL        = 30 * 24 * time.Hour
)

type affiliateAttribution struct {
	VisitID   int64     `json:"visit_id"`
	AffCode   string    `json:"aff_code"`
	ExpiresAt time.Time `json:"expires_at"`
}

func signAffiliateAttribution(secret string, attribution affiliateAttribution) string {
	payload, err := json.Marshal(attribution)
	if err != nil {
		return ""
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(encoded))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return encoded + "." + signature
}

func verifyAffiliateAttribution(secret, token string, now time.Time) (affiliateAttribution, bool) {
	var attribution affiliateAttribution
	parts := strings.Split(token, ".")
	if strings.TrimSpace(secret) == "" || len(parts) != 2 {
		return attribution, false
	}
	providedSignature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return attribution, false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(parts[0]))
	if !hmac.Equal(providedSignature, mac.Sum(nil)) {
		return attribution, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || json.Unmarshal(payload, &attribution) != nil {
		return affiliateAttribution{}, false
	}
	attribution.AffCode = strings.ToUpper(strings.TrimSpace(attribution.AffCode))
	if attribution.VisitID <= 0 || attribution.AffCode == "" || !attribution.ExpiresAt.After(now) {
		return affiliateAttribution{}, false
	}
	return attribution, true
}

func AffiliateAttributionMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cookie, err := c.Cookie(affiliateAttributionCookieName); err == nil {
			if attribution, ok := verifyAffiliateAttribution(secret, cookie, time.Now().UTC()); ok {
				ctx := service.WithAffiliateAttribution(c.Request.Context(), attribution.VisitID, attribution.AffCode)
				c.Request = c.Request.WithContext(ctx)
			}
		}
		c.Next()
	}
}

func setAffiliateAttributionCookie(c *gin.Context, secret string, visitID int64, affCode string, now time.Time) {
	expiresAt := now.Add(affiliateAttributionTTL)
	token := signAffiliateAttribution(secret, affiliateAttribution{
		VisitID:   visitID,
		AffCode:   strings.ToUpper(strings.TrimSpace(affCode)),
		ExpiresAt: expiresAt,
	})
	if token == "" {
		return
	}
	secure := c.Request.TLS != nil || strings.EqualFold(strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")), "https")
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     affiliateAttributionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(affiliateAttributionTTL.Seconds()),
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}
