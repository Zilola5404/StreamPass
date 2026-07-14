package auth

import (
	"regexp"

	apperrors "streampass/shared/errors"
)

// minPasswordLength is the only password policy the MVP spec requires
// implicitly (Argon2id hashing assumes non-trivial input). Kept as a named
// constant rather than a magic number per the "no magic numbers" rule.
const minPasswordLength = 8

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// validateCredentials applies the minimal validation the Auth use cases
// need before touching the repository or hasher.
func validateCredentials(email, password string) error {
	if !emailPattern.MatchString(email) {
		return apperrors.New(apperrors.CodeInvalidInput, "invalid email format").
			WithDetails(map[string]any{"field": "email"})
	}
	if len(password) < minPasswordLength {
		return apperrors.New(apperrors.CodeInvalidInput, "password too short").
			WithDetails(map[string]any{"field": "password", "min_length": minPasswordLength})
	}
	return nil
}
