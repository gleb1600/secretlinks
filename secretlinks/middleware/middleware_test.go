package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptionDecryption(t *testing.T) {
	secret := "test"
	encrypted := EncryptText(secret)
	decrypted := DecryptText(encrypted)
	assert.Equal(t, secret, decrypted)
}
