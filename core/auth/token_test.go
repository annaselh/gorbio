package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTokenRoundTrip(t *testing.T) {
	svc := NewTokenService("rahasia-super", time.Hour)
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	tok, _ := svc.Issue(userID, "user@x.com", []string{"admin"})
	claims, err := svc.Verify(tok)
	if err != nil {
		t.Fatalf("valid token must pas: %v", err)
	}
	if claims.UserID != userID || claims.Roles[0] != "admin" {
		t.Fatalf("wrong claim: %+v", claims)
	}
}

func TestTokenTempered(t *testing.T) {
	svc := NewTokenService("rahasia-super", time.Hour)
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	tok, _ := svc.Issue(userID, "user@x.com", nil)
	if _, err := svc.Verify(tok + "x"); err == nil {
		t.Fatal("token has changed, must not valid")
	}
}

func TestTokenExpired(t *testing.T) {
	svc := NewTokenService("rahasia-super", -time.Minute)
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	tok, _ := svc.Issue(userID, "user@x.com", nil)
	if _, err := svc.Verify(tok); err == nil {
		t.Fatal("token has expired, must not valid")
	}
}
