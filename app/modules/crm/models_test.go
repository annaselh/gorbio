package crm

import "testing"

func TestCustomerStatusValid(t *testing.T) {
	if !CustomerStatusActive.Valid() || !CustomerStatusInactive.Valid() {
		t.Fatal("both customer statuses should be valid")
	}
	for _, status := range []CustomerStatus{"", "Archived", "active"} {
		if status.Valid() {
			t.Fatalf("%q should not be accepted as a status", status)
		}
	}
}

func TestPageDefaultsAndClamps(t *testing.T) {
	if limit, offset := page(0, 0); limit != 50 || offset != 0 {
		t.Fatalf("defaults: got limit %d offset %d, want 50/0", limit, offset)
	}
	if limit, _ := page(500, 0); limit != 50 {
		t.Fatalf("an oversized limit should fall back to 50, got %d", limit)
	}
	if limit, _ := page(25, 0); limit != 25 {
		t.Fatalf("a reasonable limit should be honoured, got %d", limit)
	}
	if _, offset := page(10, -5); offset != 0 {
		t.Fatalf("a negative offset should clamp to 0, got %d", offset)
	}
}

func TestIsUniqueViolationRecognisesPostgresCodes(t *testing.T) {
	cases := []struct {
		message string
		want    bool
	}{
		{"ERROR: duplicate key value violates unique constraint", true},
		{"SQLSTATE 23505", true},
		{"pq: unique constraint failed", true},
		{"connection refused", false},
		{"record not found", false},
	}

	for _, tc := range cases {
		if got := isUniqueViolation(errString(tc.message)); got != tc.want {
			t.Fatalf("%q: got %v, want %v", tc.message, got, tc.want)
		}
	}
}

type errString string

func (e errString) Error() string { return string(e) }
