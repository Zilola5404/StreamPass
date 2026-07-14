package security

import (
	"time"

	"streampass/backend/internal/domain/user"
	apperrors "streampass/shared/errors"
	"streampass/shared/idgen"
)

const (
	claimTypeAccess  = "access"
	claimTypeRefresh = "refresh"
)

// JWTTokenIssuer implements auth.TokenIssuer using the minimal HS256 JWT
// in jwt_minimal.go. The secret is injected (never hardcoded — spec
// section "Безопасность").
type JWTTokenIssuer struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewJWTTokenIssuer builds an issuer with the given secret and TTLs.
func NewJWTTokenIssuer(secret string, accessTTL, refreshTTL time.Duration) *JWTTokenIssuer {
	return &JWTTokenIssuer{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

// IssueAccessToken mints a short-lived access token for userID.
func (i *JWTTokenIssuer) IssueAccessToken(userID user.ID) (string, time.Time, error) {
	now := time.Now().UTC()
	exp := now.Add(i.accessTTL)
	token, err := signJWT(jwtClaims{
		Sub: string(userID),
		Typ: claimTypeAccess,
		Iat: now.Unix(),
		Exp: exp.Unix(),
	}, i.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, exp, nil
}

// IssueRefreshToken mints a long-lived refresh token for userID, tagged
// with a random jti so it can be individually revoked (see
// user.SessionStore).
func (i *JWTTokenIssuer) IssueRefreshToken(userID user.ID) (string, user.RefreshTokenID, time.Time, error) {
	now := time.Now().UTC()
	exp := now.Add(i.refreshTTL)
	jti := idgen.New()
	token, err := signJWT(jwtClaims{
		Sub: string(userID),
		Jti: jti,
		Typ: claimTypeRefresh,
		Iat: now.Unix(),
		Exp: exp.Unix(),
	}, i.secret)
	if err != nil {
		return "", "", time.Time{}, err
	}
	return token, user.RefreshTokenID(jti), exp, nil
}

// ParseAccessToken verifies an access token and returns the subject user ID.
func (i *JWTTokenIssuer) ParseAccessToken(token string) (user.ID, error) {
	claims, err := verifyJWT(token, i.secret)
	if err != nil {
		return "", err
	}
	if claims.Typ != claimTypeAccess {
		return "", apperrors.New(apperrors.CodeTokenInvalid, "not an access token")
	}
	return user.ID(claims.Sub), nil
}

// ParseRefreshToken verifies a refresh token and returns the subject user
// ID and token ID (jti).
func (i *JWTTokenIssuer) ParseRefreshToken(token string) (user.ID, user.RefreshTokenID, error) {
	claims, err := verifyJWT(token, i.secret)
	if err != nil {
		return "", "", err
	}
	if claims.Typ != claimTypeRefresh {
		return "", "", apperrors.New(apperrors.CodeTokenInvalid, "not a refresh token")
	}
	return user.ID(claims.Sub), user.RefreshTokenID(claims.Jti), nil
}
