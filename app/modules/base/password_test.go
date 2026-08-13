package base

import "testing"

func TestPasswordHashRoundTrip(t *testing.T) {
	hash, err := hashPassword("a secure passphrase with enough length")
	if err != nil {
		t.Fatal(err)
	}
	if !verifyPassword(hash, "a secure passphrase with enough length") {
		t.Fatal("correct password did not verify")
	}
	if verifyPassword(hash, "wrong password") {
		t.Fatal("wrong password verified")
	}
}

func TestPasswordRejectsTooShortValue(t *testing.T) {
	if _, err := hashPassword("too-short"); err == nil {
		t.Fatal("short password was accepted")
	}
}
