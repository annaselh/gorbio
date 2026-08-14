package dashboard

import (
	"testing"
	"time"
)

func TestMonthToDateCurrentPeriod(t *testing.T) {
	now := time.Date(2026, 8, 14, 15, 30, 0, 0, time.UTC)
	current, _ := MonthToDate(now)

	if !current.Start.Equal(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("current period should start at the first of the month, got %s", current.Start)
	}
	if !current.End.Equal(now) {
		t.Fatalf("current period should end now, got %s", current.End)
	}
}

// Comparing a partial month against a whole one would report a collapse in
// revenue every time someone opens the dashboard early in the month.
func TestMonthToDateComparesEqualSpans(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	current, previous := MonthToDate(now)

	currentSpan := current.End.Sub(current.Start)
	previousSpan := previous.End.Sub(previous.Start)

	if currentSpan != previousSpan {
		t.Fatalf("spans must match: current %s, previous %s", currentSpan, previousSpan)
	}
	if !previous.Start.Equal(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("previous period should start a month earlier, got %s", previous.Start)
	}
}

// A long month compared against a short one must not spill past the boundary.
func TestMonthToDatePreviousNeverOverlapsCurrent(t *testing.T) {
	now := time.Date(2026, 3, 31, 23, 0, 0, 0, time.UTC)
	current, previous := MonthToDate(now)

	if previous.End.After(current.Start) {
		t.Fatalf("previous period must not reach into the current month: %s > %s",
			previous.End, current.Start)
	}
}

func TestPercentChange(t *testing.T) {
	cases := []struct {
		name              string
		current, previous int64
		want              float64
	}{
		{"growth", 150, 100, 50},
		{"decline", 50, 100, -50},
		{"flat", 100, 100, 0},
		{"to zero", 0, 100, -100},
	}

	for _, tc := range cases {
		if got := percentChange(tc.current, tc.previous); got != tc.want {
			t.Fatalf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}

// No baseline means no measurable change. Reporting a jump from zero as growth
// would put a fabricated percentage on the dashboard in a tenant's first month.
func TestPercentChangeWithoutBaselineIsZero(t *testing.T) {
	if got := percentChange(500, 0); got != 0 {
		t.Fatalf("growth from a zero baseline should report 0, got %v", got)
	}
}

func TestNewMetricCarriesBothPeriods(t *testing.T) {
	metric := newMetric(120, 100)

	if metric.Value != 120 || metric.Previous != 100 {
		t.Fatalf("metric should carry both figures, got %+v", metric)
	}
	if metric.Delta != 20 {
		t.Fatalf("delta should be 20, got %v", metric.Delta)
	}
}
