package auth

import (
	"testing"
)

func TestHashVerify(t *testing.T) {
	hash, err := HashPassword("passwordRahasia123")
	if err != nil {
		t.Fatal(err)
	}
	ok, err := VerifyPassword("passwordRahasia123", hash)
	if err != nil || !ok {
		t.Fatalf("password tidak cocok, ok=%v err=%v", ok, err)
	}
	bad, _ := VerifyPassword("salah", hash)
	if bad {
		t.Fatal("password salah tidak boleh terverifikasi")
	}
}

func TestHashUnique(t *testing.T) {
	h1, _ := HashPassword("passwordRahasia123")
	h2, _ := HashPassword("passwordRahasia123")
	if h1 == h2 {
		t.Fatal("hash harus unik walaupun password sama")
	}
}
