package hash

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

func encodeHash(salt, hash []byte) string {
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		p.memory,
		p.iterations,
		p.parallelism,
		b64Salt,
		b64Hash,
	)
}

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

func decodeHash(encodedHash string) (salt, hash []byte, err error) {
	values := strings.Split(encodedHash, "$")
	if len(values) != 6 {
		return nil, nil, ErrInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(values[2], "v=%d", &version)
	if err != nil {
		return nil, nil, err
	}

	if version != argon2.Version {
		return nil, nil, ErrIncompatibleVersion
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(values[4])
	if err != nil {
		return nil, nil, err
	}

	hash, err = base64.RawStdEncoding.Strict().DecodeString(values[5])
	if err != nil {
		return nil, nil, err
	}

	return salt, hash, nil
}
