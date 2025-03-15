package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/boseji/auth/aesgcm"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func keyIs() []byte {
	return []byte("hello_this_is_32_symbols_string!")
}

func EncryptText(text string) string {
	iNonce := make([]byte, aesgcm.NonceSize)
	ciphertext, _, err := aesgcm.Encrypt([]byte(text), keyIs(), iNonce)
	if err != nil {
		panic(err)
	}
	return string(ciphertext)
}

func DecryptText(ciphertext string) string {
	iNonce := make([]byte, aesgcm.NonceSize)
	text, err := aesgcm.Decrypt([]byte(ciphertext), iNonce, keyIs())
	if err != nil {
		panic(err)
	}
	return string(text)
}
