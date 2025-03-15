package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"secretlinks/middleware"
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

		secret := middleware.EncryptText(r.FormValue("secret"))

		expiration, err := strconv.Atoi(r.FormValue("expiration"))

		if err != nil {
			if r.FormValue("expiration") == "" {
				expiration = 60
			} else {
				http.Error(w, "Expected int value", http.StatusNotAcceptable)
				return
			}
		}

		maxViews, err := strconv.Atoi(r.FormValue("maxviews"))

		if err != nil {
			if r.FormValue("maxviews") == "" {
				maxViews = 1
			} else {
				http.Error(w, "Expected int value", http.StatusNotAcceptable)
				return
			}
		}

		var resultKey string

		for keyIsUnique := false; !keyIsUnique; {
			key := generateKey(8)
			expiresAt := time.Now().Add(time.Duration(expiration) * time.Minute) // Пример: фиксированное время
			link := storage.Link{
				Secret:    secret,
				ExpiresAt: expiresAt,
				MaxViews:  maxViews,
			}

			keyIsUnique = s.Create(key, link, true)
			resultKey = key
		}

		fmt.Fprintf(w, "http://%s/%s", r.Host, resultKey)
	}
}
