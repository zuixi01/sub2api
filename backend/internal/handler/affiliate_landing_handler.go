package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type AffiliateLandingHandler struct {
	affiliateService *service.AffiliateService
	trackingSecret   string
}

func NewAffiliateLandingHandler(affiliateService *service.AffiliateService, cfg *config.Config) *AffiliateLandingHandler {
	secret := ""
	if cfg != nil {
		secret = cfg.JWT.Secret
	}
	return &AffiliateLandingHandler{affiliateService: affiliateService, trackingSecret: secret}
}

func (h *AffiliateLandingHandler) Redirect(c *gin.Context) {
	code := strings.ToUpper(strings.TrimSpace(c.Param("affCode")))
	if h == nil || h.affiliateService == nil || strings.TrimSpace(h.trackingSecret) == "" {
		c.Status(http.StatusNotFound)
		return
	}
	day := time.Now().UTC()
	_, err := h.affiliateService.RecordAffiliateVisit(c.Request.Context(), service.AffiliateVisitInput{
		AffCode: code, VisitedOn: day, VisitorHash: affiliateVisitorHash(h.trackingSecret, day, clientIP(c.Request), c.Request.UserAgent()),
		UTMSource: safeUTM(c.Query("utm_source")), UTMMedium: safeUTM(c.Query("utm_medium")), UTMCampaign: safeUTM(c.Query("utm_campaign")),
	})
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	query := url.Values{"aff_code": []string{code}}
	for _, key := range []string{"utm_source", "utm_medium", "utm_campaign"} {
		if value := safeUTM(c.Query(key)); value != "" {
			query.Set(key, value)
		}
	}
	c.Redirect(http.StatusFound, "/register?"+query.Encode())
}

func affiliateVisitorHash(secret string, day time.Time, ip, agent string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(day.UTC().Format("2006-01-02") + "\n" + ip + "\n" + agent))
	return hex.EncodeToString(mac.Sum(nil))
}

func clientIP(req *http.Request) string {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err == nil {
		return host
	}
	return req.RemoteAddr
}
func safeUTM(value string) string {
	value = strings.TrimSpace(value)
	if len(value) == 0 || len(value) > 100 {
		return ""
	}
	for i := 0; i < len(value); i++ {
		c := value[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '.' || c == '_' || c == '-') {
			return ""
		}
	}
	return value
}
