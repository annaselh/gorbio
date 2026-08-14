package procurement

import "testing"

func TestArrivesAtReceivedOnlyOnTheTransitionIn(t *testing.T) {
	if !arrivesAtReceived(PurchaseStatusConfirmed, PurchaseStatusReceived) {
		t.Fatal("confirmed -> received is the delivery landing and must move stock")
	}
	if !arrivesAtReceived(PurchaseStatusDraft, PurchaseStatusReceived) {
		t.Fatal("any transition into received must move stock")
	}
}

func TestArrivesAtReceivedIsNotRepeatedByAResend(t *testing.T) {
	// validatePurchaseTransition permits from == to as a no-op, so a client
	// resending Received must not book the same delivery into stock twice.
	if arrivesAtReceived(PurchaseStatusReceived, PurchaseStatusReceived) {
		t.Fatal("re-sending Received must not count as a second arrival")
	}
}

func TestArrivesAtReceivedIgnoresOtherTransitions(t *testing.T) {
	for _, transition := range []struct{ from, to PurchaseStatus }{
		{PurchaseStatusDraft, PurchaseStatusConfirmed},
		{PurchaseStatusDraft, PurchaseStatusCancelled},
		{PurchaseStatusConfirmed, PurchaseStatusCancelled},
		{PurchaseStatusReceived, PurchaseStatusConfirmed},
	} {
		if arrivesAtReceived(transition.from, transition.to) {
			t.Fatalf("%s -> %s must not move stock", transition.from, transition.to)
		}
	}
}
