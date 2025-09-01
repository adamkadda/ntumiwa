package hash

import (
	"crypto/rand"

	"golang.org/x/crypto/argon2"
)

func GenerateFromPassword(password string) (hash []byte, err error) {
	salt, err := generateRandomBytes(p.saltLength)
	if err != nil {
		return nil, err
	}

	hash = argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	return hash, nil
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
