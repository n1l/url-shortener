package filestorage

import (
	"encoding/json"
	"os"

	"github.com/n1l/url-shortener/internal/models"
)

type FileStorage struct {
	file    *os.File
	encoder *json.Encoder
	decoder *json.Decoder
}

func NewFileStorage(filename string) (*FileStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileStorage{
		file:    file,
		encoder: json.NewEncoder(file),
		decoder: json.NewDecoder(file),
	}, nil
}

func (storage *FileStorage) Write(rec *models.URLRecord) error {
	return storage.encoder.Encode(rec)
}

func (storage *FileStorage) Read(rec *models.URLRecord) error {
	return storage.decoder.Decode(&rec)
}

func (p *FileStorage) Close() error {
	return p.file.Close()
}
