package order

import "testing"

func TestRenewalConfirmationPolicy(t *testing.T) {
	if !CanConfirmRenewal(StatusPending, TypeRenewal) {
		t.Fatal("pending renewal order should be confirmable")
	}
	if CanConfirmRenewal(StatusFulfilled, TypeRenewal) {
		t.Fatal("fulfilled renewal order must not be confirmed again")
	}
	if CanConfirmRenewal(StatusPending, TypePurchase) {
		t.Fatal("purchase order must not use renewal confirmation")
	}
}

func TestBillingCycleMonths(t *testing.T) {
	cases := map[string]int{"monthly": 1, "quarterly": 3, "semi_yearly": 6, "yearly": 12}
	for cycle, want := range cases {
		got, ok := BillingCycleMonths(cycle)
		if !ok || got != want {
			t.Fatalf("cycle %q got (%d, %v), want (%d, true)", cycle, got, ok, want)
		}
	}
	if got, ok := BillingCycleMonths("weekly"); ok || got != 0 {
		t.Fatalf("unsupported cycle got (%d, %v), want (0, false)", got, ok)
	}
}
