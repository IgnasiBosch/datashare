package service

import (
	"reflect"
	"testing"
)

var testData = []struct {
	name    string
	key     string
	content []byte
	err     string
}{
	{
		name:    "EmptyKey",
		key:     "",
		content: []byte("SomeContent"),
		err:     "crypto/aes: invalid key size 0",
	},
	{
		name:    "EmptyKeyAndContent",
		key:     "",
		content: []byte("SomeContent"),
		err:     "crypto/aes: invalid key size 0",
	},
	{
		name:    "RegularUse",
		key:     "SomeKey",
		content: []byte("SomeContent"),
		err:     "",
	},
}

func TestEncryption(t *testing.T) {
	e := NewEncryption(1, 32, 16, "SomeHashSalt")

	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {

			ciphertext, err := e.Encrypt(td.key, td.content)
			if err != nil && err.Error() != td.err {
				t.Fatalf("Expected error '%s', but got '%s'", td.err, err.Error())
			}

			if td.err == "" {
				if ciphertext == nil {
					t.Fatalf("Expected ciphertext, but got nothing")
				}

				plaintext, perr := e.Decrypt(td.key, ciphertext)
				if perr != nil {
					t.Fatalf("Expected no error but got: '%s'", perr.Error())
				}

				if !reflect.DeepEqual(plaintext, td.content) {
					t.Fatalf("Expected content '%s' but got '%s'", td.content, plaintext)
				}

				hash := e.HashString(string(td.content))
				if hash == "" {
					t.Fatalf("Expected hash, but got nothing")
				}
			}
		})
	}
}
