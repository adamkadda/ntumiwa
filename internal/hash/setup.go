package hash

import (
	"github.com/adamkadda/ntumiwa-site/shared/config"
	"golang.org/x/crypto/argon2"
)

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

var p params

var dummy string

func Setup(cfg config.HashConfig) {
	p = params{
		memory:      cfg.Memory,
		iterations:  cfg.Iterations,
		parallelism: cfg.Parallelism,
		saltLength:  cfg.SaltLength,
		keyLength:   cfg.KeyLength,
	}

	salt := make([]byte, p.saltLength)
	password := []byte("foo")
	hash := argon2.IDKey(
		password,
		salt,
		p.iterations,
		p.memory,
		p.parallelism,
		p.keyLength,
	)

	dummy = encodeHash(salt, hash)
}

func DummyHash() string {
	return dummy
}
