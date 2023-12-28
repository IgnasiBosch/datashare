package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"golang.org/x/crypto/pbkdf2"
)

type Encryption struct {
	iterations int
	blockSize  int
	saltLength int
	ivLength   int
	hashSalt   string
}

func NewEncryption(iterations, blockSize, saltLength int, hashSalt string) *Encryption {
	return &Encryption{
		iterations: iterations,
		blockSize:  blockSize,
		saltLength: saltLength,
		ivLength:   12, // Must have same value as https://github.com/golang/go/blob/master/src/crypto/cipher/gcm.go#L157
		hashSalt:   hashSalt,
	}
}

func (e *Encryption) deriveKey(key string, salt []byte) ([]byte, []byte) {
	// http://www.ietf.org/rfc/rfc2898.txt
	if salt == nil {
		salt = make([]byte, e.saltLength)
		rand.Read(salt)
	}
	return pbkdf2.Key([]byte(key), salt, e.iterations, e.blockSize, sha256.New), salt
}

func (e *Encryption) Encrypt(key string, content []byte) ([]byte, error) {
	derivedKey, salt := e.deriveKey(key, nil)
	iv := make([]byte, e.ivLength)
	// http://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-38d.pdf
	// Section 8.2
	rand.Read(iv)
	b, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(b)
	if err != nil {
		return nil, err
	}
	data := aesgcm.Seal(nil, iv, content, nil)
	return append(append(salt, iv...), data...), nil
}

func (e *Encryption) Decrypt(key string, cipherContent []byte) ([]byte, error) {
	// The salt and iv are the first saltLength and ivLength bytes of the cipherContent
	salt, iv, data := cipherContent[:e.saltLength], cipherContent[e.saltLength:e.saltLength+e.ivLength], cipherContent[e.saltLength+e.ivLength:]
	derivedKey, _ := e.deriveKey(key, salt)
	b, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(b)
	if err != nil {
		return nil, err
	}
	data, err = aesgcm.Open(nil, iv, data, nil)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (e *Encryption) HashString(s string) string {
	saltedString := s + e.hashSalt
	hash := sha256.Sum256([]byte(saltedString))
	return hex.EncodeToString(hash[:])
}
