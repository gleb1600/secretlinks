package storage

import (
	"sync"
	"time"
)

type Link struct {
	Secret    string
	ExpiresAt time.Time
	MaxViews  int
	Views     int
}

type Storage interface {
	Create(key string, link Link, b bool) bool
	Update(key string, link Link)
	Get(key string) (Link, bool)
	Delete(key string)
}

type MemoryStorage struct {
	mu    sync.Mutex
	links map[string]Link
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		links: make(map[string]Link),
	}
}

func (s *MemoryStorage) Create(key string, link Link, b bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k := range s.links {
		if k == key {
			b = false
			return b
		}
	}
	s.links[key] = link
	return b
}

func (s *MemoryStorage) Update(key string, link Link) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.links[key] = link
}

func (s *MemoryStorage) Get(key string) (Link, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	link, exists := s.links[key]
	return link, exists
}

func (s *MemoryStorage) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.links, key)
}
