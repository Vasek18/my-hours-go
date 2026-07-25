package auth

import "testing"

func TestGenerateToken(t *testing.T) {
	token, hash, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if token == "" || hash == "" {
		t.Fatal("token and hash must be non-empty")
	}
	if token == hash {
		t.Fatal("token must not equal its hash")
	}
	if got := HashToken(token); got != hash {
		t.Fatalf("HashToken(token) = %q, want %q", got, hash)
	}
}

func TestGenerateTokenIsRandom(t *testing.T) {
	t1, _, _ := GenerateToken()
	t2, _, _ := GenerateToken()
	if t1 == t2 {
		t.Fatal("two generated tokens should differ")
	}
}
