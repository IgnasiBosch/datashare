package service

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	mrand "math/rand"
	"time"
)

const (
	idLength  = 25
	idCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	keyLength  = 12
	keyCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-<>!$()=*:;"
)

var seededRand *mrand.Rand = mrand.New(
	mrand.NewSource(time.Now().UnixNano()),
)

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func NewKey() string {
	return StringWithCharset(keyLength, keyCharset)
}

func NewID(typePrefix string) string {
	return typePrefix + "_" + StringWithCharset(idLength, idCharset)
}

func generateAndEncodeKey(length int, encode func([]byte) string) (string, error) {
	key := make([]byte, length)

	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}

	return encode(key), nil
}

func GenerateAuthKey() (string, error) {
	return generateAndEncodeKey(64, base64.StdEncoding.EncodeToString)
}

func GenerateCSRFSecret() (string, error) {
	return generateAndEncodeKey(32, hex.EncodeToString)
}
