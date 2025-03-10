package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"secretlinks/storage"
	"strconv"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateKey(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func CreateHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		secret := r.FormValue("secret")
		expiration, _ := strconv.Atoi(r.FormValue("expiration"))
		if expiration == 0 {
			expiration = 60
		}
		maxViews, _ := strconv.Atoi(r.FormValue("maxviews"))
		if maxViews == 0 {
			maxViews = 1
		}
		// Валидация тут

		key := generateKey(8)
		expiresAt := time.Now().Add(time.Duration(expiration) * time.Minute) // Пример: фиксированное время
		link := storage.Link{
			Secret:    secret,
			ExpiresAt: expiresAt,
			MaxViews:  maxViews,
		}

		s.Create(key, link)

		fmt.Fprintf(w, "http://%s/%s", r.Host, key)
	}
}
