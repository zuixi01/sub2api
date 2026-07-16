package service

import (
	"context"
	"strings"
)

type affiliateAttributionContextKey struct{}

// AffiliateAttributionContext carries a verified landing visit through the
// existing registration flow without exposing the signed cookie to services.
type AffiliateAttributionContext struct {
	VisitID int64
	AffCode string
}

func WithAffiliateAttribution(ctx context.Context, visitID int64, affCode string) context.Context {
	if ctx == nil || visitID <= 0 {
		return ctx
	}
	return context.WithValue(ctx, affiliateAttributionContextKey{}, AffiliateAttributionContext{
		VisitID: visitID,
		AffCode: strings.ToUpper(strings.TrimSpace(affCode)),
	})
}

func AffiliateAttributionFromContext(ctx context.Context) (AffiliateAttributionContext, bool) {
	if ctx == nil {
		return AffiliateAttributionContext{}, false
	}
	attribution, ok := ctx.Value(affiliateAttributionContextKey{}).(AffiliateAttributionContext)
	return attribution, ok && attribution.VisitID > 0 && attribution.AffCode != ""
}
