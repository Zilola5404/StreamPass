module streampass

go 1.22.2

require (
	golang.org/x/crypto v0.24.0
	golang.org/x/sys v0.21.0
)

replace golang.org/x/crypto => ./vendor-src/crypto
replace golang.org/x/sys => ./vendor-src/sys

require github.com/lib/pq v1.10.9
replace github.com/lib/pq => ./vendor-src/pq
