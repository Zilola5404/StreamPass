// Package security holds infrastructure-layer implementations of security
// ports (password hashing today; nothing else yet — YAGNI).
package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	apperrors "streampass/shared/errors"
)

// argon2Params holds the tuning parameters for Argon2id. Values follow the
// OWASP-recommended baseline for interactive login (spec section 17
// mandates Argon2id but leaves parameters to implementation).
type argon2Params struct {
	memoryKiB   uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

var defaultParams = argon2Params{
	memoryKiB:   64 * 1024, // 64 MiB
	iterations:  3,
	parallelism: 2,
	saltLength:  16,
	keyLength:   32,
}

// Argon2Hasher implements auth.PasswordHasher using Argon2id. The encoded
// hash format is the PHC-style string used by the reference Argon2 CLI, so
// parameters travel with the hash and can be tuned over time without
// breaking verification of older hashes.
type Argon2Hasher struct {
	params argon2Params
}

// NewArgon2Hasher builds a hasher with the default, vetted parameters.
func NewArgon2Hasher() *Argon2Hasher {
	return &Argon2Hasher{params: defaultParams}
}

// Hash produces a PHC-encoded Argon2id hash of plaintext with a fresh
// random salt.
func (h *Argon2Hasher) Hash(plaintext string) (string, error) {
	salt := make([]byte, h.params.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", apperrors.Wrap(apperrors.CodeInternal, "failed to generate salt", err)
	}

	key := argon2.IDKey([]byte(plaintext), salt, h.params.iterations, h.params.memoryKiB, h.params.parallelism, h.params.keyLength)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.params.memoryKiB,
		h.params.iterations,
		h.params.parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)
	return encoded, nil
}

// Verify checks plaintext against a PHC-encoded Argon2id hash using a
// constant-time comparison to avoid timing side channels.
func (h *Argon2Hasher) Verify(encodedHash, plaintext string) bool {
	params, salt, key, err := decodeArgon2Hash(encodedHash)
	if err != nil {
		return false
	}

	candidate := argon2.IDKey([]byte(plaintext), salt, params.iterations, params.memoryKiB, params.parallelism, uint32(len(key)))
	return subtle.ConstantTimeCompare(candidate, key) == 1
}

func decodeArgon2Hash(encoded string) (argon2Params, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return argon2Params{}, nil, nil, apperrors.New(apperrors.CodeInvalidInput, "malformed argon2id hash")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return argon2Params{}, nil, nil, apperrors.Wrap(apperrors.CodeInvalidInput, "malformed argon2id version", err)
	}

	var p argon2Params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.memoryKiB, &p.iterations, &p.parallelism); err != nil {
		return argon2Params{}, nil, nil, apperrors.Wrap(apperrors.CodeInvalidInput, "malformed argon2id params", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return argon2Params{}, nil, nil, apperrors.Wrap(apperrors.CodeInvalidInput, "malformed argon2id salt", err)
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return argon2Params{}, nil, nil, apperrors.Wrap(apperrors.CodeInvalidInput, "malformed argon2id key", err)
	}

	return p, salt, key, nil
}
