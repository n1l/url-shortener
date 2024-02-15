package database

import (
	"bufio"
	"encoding/json"
	"os"
)

type URLRecord struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *Producer) Write(event *URLRecord) error {
	return p.encoder.Encode(event)
}

func (p *Producer) Close() error {
	return p.file.Close()
}

func Load(fname string) ([]URLRecord, error) {
	var slice []URLRecord

	file, err := os.OpenFile(fname, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		rec := URLRecord{}
		json.Unmarshal(fileScanner.Bytes(), &rec)
		slice = append(slice, rec)
	}

	return slice, nil
}
