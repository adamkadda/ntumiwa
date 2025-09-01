package cookies

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
)

/*
	Thanks to Alex Edwards, I trust this dude.
	The point is to define these functions that will be called
	by other packages when working with cookies.

	https://www.alexedwards.net/blog/working-with-cookies-in-go
*/

var (
	ErrValueTooLong = errors.New("cookie value too long")
	ErrInvalidValue = errors.New("invalid cookie value")
)

func Write(w http.ResponseWriter, cookie http.Cookie) error {
	cookie.Value = base64.URLEncoding.EncodeToString([]byte(cookie.Value))

	if len(cookie.String()) > 4096 {
		return ErrValueTooLong
	}

	http.SetCookie(w, &cookie)

	return nil
}

func Read(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	value, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return "", ErrInvalidValue
	}

	return string(value), nil
}

/*
	Don't trust cookies. They're stored on the client and can be edited.
	We need a way to verify that a cookie hasn't been edited, and has
	the same name and value we've set.

	HMAC (Hashed based Message Authentication Code) is a type of
	MAC (Message Authentication Code) that is acquired by executing a
	cryptographic hash function on the data is to be authenticated.
*/

func WriteSigned(w http.ResponseWriter, cookie http.Cookie, secretKey []byte) error {
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(cookie.Name))
	mac.Write([]byte(cookie.Value))
	signature := mac.Sum(nil)

	// Prepend cookie.Value with HMAC signature
	cookie.Value = string(signature) + cookie.Value

	return Write(w, cookie)
}

func ReadSigned(r *http.Request, name string, secretKey []byte) (string, error) {
	signedValue, err := Read(r, name)
	if err != nil {
		return "", err
	}

	//
	if len(signedValue) < sha256.Size {
		return "", ErrInvalidValue
	}

	signature := signedValue[:sha256.Size]
	value := signedValue[sha256.Size:]

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(name))
	mac.Write([]byte(value))
	expectedSignature := mac.Sum(nil)

	// Compare two MACs
	if !hmac.Equal([]byte(signature), expectedSignature) {
		return "", ErrInvalidValue
	}

	return value, nil
}
