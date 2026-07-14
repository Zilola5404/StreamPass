// Package idgen generates unique, unpredictable identifiers using
// crypto/rand. A UUID library was deliberately not vendored: StreamPass
// only needs opaque unique strings (not RFC 4122 compliance), so a
// dependency-free generator satisfies KISS/YAGNI.
package idgen

import (
	"crypto/rand"
	"encoding/hex"
)

// idByteLength gives 256 bits of entropy — comfortably collision-free at
// any scale this project will reach.
const idByteLength = 32

// New returns a random 64-character hex string suitable as a primary key
// or token identifier.
func New() string {
	b := make([]byte, idByteLength)
	// crypto/rand.Read only fails if the OS entropy source is unavailable,
	// which is an unrecoverable environment fault, not a business error —
	// panicking here is intentional (fail fast rather than silently
	// returning a low-entropy or empty ID).
	if _, err := rand.Read(b); err != nil {
		panic("idgen: failed to read random bytes: " + err.Error())
	}
	return hex.EncodeToString(b)
}
