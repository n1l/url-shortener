package memcache

import (
	"io"

	"github.com/n1l/url-shortener/internal/filestorage"
	"github.com/n1l/url-shortener/internal/models"
)

type MemCache struct {
	cache map[string]string
}

func NewMemCache() *MemCache {
	return &MemCache{
		cache: make(map[string]string),
	}
}

func NewMemCacheFromFile(s *filestorage.FileStorage) (*MemCache, error) {
	cache := NewMemCache()

	for {
		var rec models.URLRecord
		if err := s.Read(&rec); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		cache.add(&rec)
	}

	return cache, nil
}

func (m *MemCache) add(rec *models.URLRecord) {
	m.cache[rec.ShortURL] = rec.OriginalURL
}

func (m *MemCache) Add(rec *models.URLRecord) {
	m.add(rec)
}

func (m *MemCache) Get(hash string) (string, bool) {
	val, ok := m.cache[hash]
	return val, ok
}
