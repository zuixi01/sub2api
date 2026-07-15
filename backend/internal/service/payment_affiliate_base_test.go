package service

import (
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
)

func TestAffiliateRebateBaseAmountUsesActualPayment(t *testing.T) {
	t.Parallel()

	order := &dbent.PaymentOrder{
		OrderType: payment.OrderTypeBalance,
		Amount:    120,
		PayAmount: 100,
	}
	if got := affiliateRebateBaseAmount(order); got != 100 {
		t.Fatalf("affiliate base must use pay amount, got %v", got)
	}

	order.PayAmount = 0
	if got := affiliateRebateBaseAmount(order); got != 0 {
		t.Fatalf("unpaid order must not have an affiliate base, got %v", got)
	}
}
