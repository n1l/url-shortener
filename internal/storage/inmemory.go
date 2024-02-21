package storage

import (
	"sync"

	"github.com/n1l/url-shortener/internal/models"
)

type InMemoryStorage struct {
	lock  sync.Mutex
	cache map[string]*models.URLRecord
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		cache: make(map[string]*models.URLRecord),
	}
}

func (s *InMemoryStorage) Save(rec *models.URLRecord) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cache[rec.ShortURL] = rec
	return nil
}

func (s *InMemoryStorage) Get(hash string) (string, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	val, ok := s.cache[hash]
	if ok {
		return val.OriginalURL, ok
	}
	return "", ok
}

func (s *InMemoryStorage) saveInternal(rec *models.URLRecord) {
	s.cache[rec.ShortURL] = rec
}
