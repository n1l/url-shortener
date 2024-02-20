package storage

import (
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/n1l/url-shortener/internal/models"
)

type FileStorage struct {
	lock    sync.Mutex
	cache   *InMemoryStorage
	file    *os.File
	encoder *json.Encoder
	decoder *json.Decoder
}

func NewFileStorage(filename string) (*FileStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	storage := &FileStorage{
		cache:   NewInMemoryStorage(),
		file:    file,
		encoder: json.NewEncoder(file),
		decoder: json.NewDecoder(file),
	}

	err = updateCache(storage)
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *FileStorage) Save(rec *models.URLRecord) error {
	s.cache.Save(rec)
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.encoder.Encode(rec)
}

func (s *FileStorage) Get(hash string) (string, bool) {
	return s.cache.Get(hash)
}

func (s *FileStorage) Close() error {
	return s.file.Close()
}

func updateCache(s *FileStorage) error {

	for {
		var rec models.URLRecord
		if err := readFromFile(s, &rec); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		saveInternal(s.cache, &rec)
	}

	return nil
}

func readFromFile(storage *FileStorage, rec *models.URLRecord) error {
	return storage.decoder.Decode(&rec)
}
