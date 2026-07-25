package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "correct horse battery staple" {
		t.Fatal("hash must not equal the plaintext password")
	}
	if !CheckPassword(hash, "correct horse battery staple") {
		t.Fatal("CheckPassword should accept the correct password")
	}
	if CheckPassword(hash, "wrong password") {
		t.Fatal("CheckPassword should reject an incorrect password")
	}
}
