package handlers

import (
	"net/http"
	"secretlinks/storage"
	"time"
)

func RedirectHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path[1:]
		link, exists := s.Get(key)

		if !exists {
			http.NotFound(w, r)
			return
		}

		if link.Views >= link.MaxViews {
			http.Error(w, "Link expired", http.StatusGone)
			s.Delete(key)
			return
		}

		if time.Now().After(link.ExpiresAt) {
			http.Error(w, "Link expired", http.StatusGone)
			s.Delete(key)
			return
		}

		link.Views++
		s.Update(key, link)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(link.Secret))
	}
}
