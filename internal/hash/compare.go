package hash

import (
	"crypto/subtle"

	"golang.org/x/crypto/argon2"
)

func CompareHashAndPassword(encodedHash, password string) bool {
	decodedOk := true

	// NOTE: Does this introduce meaningful timing differences?
	salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		decodedOk = false
		salt, hash, _ = decodeHash(Dummy)
	}

	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		p.iterations,
		p.memory,
		p.parallelism,
		p.keyLength,
	)

	match := subtle.ConstantTimeCompare(hash, otherHash) == 1

	if !(decodedOk && match) {
		return false
	}

	return true
}
