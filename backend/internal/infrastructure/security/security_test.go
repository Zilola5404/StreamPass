package security

import (
	"testing"
	"time"

	"streampass/backend/internal/domain/user"
)

func TestArgon2Hasher_HashAndVerify(t *testing.T) {
	h := NewArgon2Hasher()

	hash, err := h.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if !h.Verify(hash, "correct horse battery staple") {
		t.Error("Verify() = false for correct password, want true")
	}
	if h.Verify(hash, "wrong password") {
		t.Error("Verify() = true for wrong password, want false")
	}
}

func TestArgon2Hasher_DistinctSaltsProduceDistinctHashes(t *testing.T) {
	h := NewArgon2Hasher()
	h1, _ := h.Hash("same-password")
	h2, _ := h.Hash("same-password")
	if h1 == h2 {
		t.Error("two hashes of the same password should differ due to random salt")
	}
}

func TestJWTTokenIssuer_AccessTokenRoundTrip(t *testing.T) {
	issuer := NewJWTTokenIssuer("test-secret", time.Minute, time.Hour)

	token, exp, err := issuer.IssueAccessToken(user.ID("user-123"))
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}
	if exp.Before(time.Now()) {
		t.Error("access token expiry should be in the future")
	}

	gotID, err := issuer.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("ParseAccessToken() error = %v", err)
	}
	if gotID != "user-123" {
		t.Errorf("ParseAccessToken() = %q, want user-123", gotID)
	}
}

func TestJWTTokenIssuer_RefreshTokenRoundTrip(t *testing.T) {
	issuer := NewJWTTokenIssuer("test-secret", time.Minute, time.Hour)

	token, tokenID, _, err := issuer.IssueRefreshToken(user.ID("user-456"))
	if err != nil {
		t.Fatalf("IssueRefreshToken() error = %v", err)
	}

	gotUserID, gotTokenID, err := issuer.ParseRefreshToken(token)
	if err != nil {
		t.Fatalf("ParseRefreshToken() error = %v", err)
	}
	if gotUserID != "user-456" || gotTokenID != tokenID {
		t.Errorf("ParseRefreshToken() = (%q, %q), want (user-456, %q)", gotUserID, gotTokenID, tokenID)
	}
}

func TestJWTTokenIssuer_RejectsTamperedToken(t *testing.T) {
	issuer := NewJWTTokenIssuer("test-secret", time.Minute, time.Hour)
	token, _, err := issuer.IssueAccessToken(user.ID("user-789"))
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}

	tampered := token[:len(token)-1] + "x"
	if _, err := issuer.ParseAccessToken(tampered); err == nil {
		t.Error("expected error for tampered token, got nil")
	}
}

func TestJWTTokenIssuer_RejectsExpiredToken(t *testing.T) {
	issuer := NewJWTTokenIssuer("test-secret", -time.Minute, time.Hour)
	token, _, err := issuer.IssueAccessToken(user.ID("user-999"))
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}
	if _, err := issuer.ParseAccessToken(token); err == nil {
		t.Error("expected error for expired token, got nil")
	}
}

func TestJWTTokenIssuer_RejectsWrongTokenTypeForSlot(t *testing.T) {
	issuer := NewJWTTokenIssuer("test-secret", time.Minute, time.Hour)
	refreshToken, _, _, err := issuer.IssueRefreshToken(user.ID("user-1"))
	if err != nil {
		t.Fatalf("IssueRefreshToken() error = %v", err)
	}
	if _, err := issuer.ParseAccessToken(refreshToken); err == nil {
		t.Error("expected error when parsing a refresh token as an access token, got nil")
	}
}
