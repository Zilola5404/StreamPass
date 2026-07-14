package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	apperrors "streampass/shared/errors"
)

// This file implements the minimal subset of JWT (HS256, exp/sub claims)
// StreamPass needs. A dependency-free implementation was chosen over
// vendoring a third-party JWT library because the claim set is tiny and
// fixed (KISS/YAGNI) — see shared/config's yaml_minimal.go for the same
// rationale applied to YAML.

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type jwtClaims struct {
	Sub string `json:"sub"`           // user ID
	Jti string `json:"jti,omitempty"` // token id (refresh tokens only)
	Typ string `json:"typ"`           // "access" | "refresh"
	Exp int64  `json:"exp"`
	Iat int64  `json:"iat"`
}

var jwtHeaderEncoded = base64URLEncode(mustMarshal(jwtHeader{Alg: "HS256", Typ: "JWT"}))

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err) // only ever called with the fixed jwtHeader literal above
	}
	return b
}

func base64URLEncode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

// signJWT builds and signs a compact JWS with the given claims and secret.
func signJWT(claims jwtClaims, secret []byte) (string, error) {
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", apperrors.Wrap(apperrors.CodeInternal, "failed to marshal jwt claims", err)
	}
	payload := jwtHeaderEncoded + "." + base64URLEncode(claimsJSON)

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	signature := base64URLEncode(mac.Sum(nil))

	return payload + "." + signature, nil
}

// verifyJWT validates the signature and expiry of a compact JWS and
// returns its claims.
func verifyJWT(token string, secret []byte) (jwtClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return jwtClaims{}, apperrors.New(apperrors.CodeTokenInvalid, "malformed token")
	}

	payload := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	expectedSig := base64URLEncode(mac.Sum(nil))

	if subtle.ConstantTimeCompare([]byte(expectedSig), []byte(parts[2])) != 1 {
		return jwtClaims{}, apperrors.New(apperrors.CodeTokenInvalid, "invalid token signature")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return jwtClaims{}, apperrors.Wrap(apperrors.CodeTokenInvalid, "malformed token claims", err)
	}

	var claims jwtClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return jwtClaims{}, apperrors.Wrap(apperrors.CodeTokenInvalid, "malformed token claims", err)
	}

	if time.Now().UTC().Unix() > claims.Exp {
		return jwtClaims{}, apperrors.New(apperrors.CodeTokenExpired, "token expired")
	}

	return claims, nil
}
