package urlstorage

import (
	"sync"

	"github.com/n1l/url-shortener/internal/filestorage"
	"github.com/n1l/url-shortener/internal/memcache"
	"github.com/n1l/url-shortener/internal/models"
)

type URLStorage struct {
	lock    sync.Mutex
	cache   *memcache.MemCache
	storage *filestorage.FileStorage
}

func NewURLStorage(c *memcache.MemCache, s *filestorage.FileStorage) *URLStorage {
	return &URLStorage{
		cache:   c,
		storage: s,
	}
}

func (s *URLStorage) Get(hash string) (string, bool) {
	return s.cache.Get(hash)
}

func (s *URLStorage) Store(rec *models.URLRecord) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cache.Add(rec)
	s.storage.Write(rec)
}
