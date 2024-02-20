package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type Options struct {
	PrivateHost string `env:"SERVER_ADDRESS"`
	PublicHost  string `env:"BASE_URL"`
	StoragePath string `env:"FILE_STORAGE_PATH"`
	LogLevel    string `env:"LOG_LEVEL"`
}

func ParseOptions(ops *Options) {
	flag.StringVar(&ops.PrivateHost, "a", "localhost:8080", "The service address at start")
	flag.StringVar(&ops.PublicHost, "b", "http://localhost:8080", "The shortener result base address")
	flag.StringVar(&ops.StoragePath, "f", "/tmp/short-url-db.json", "The shortener file storage")
	flag.StringVar(&ops.LogLevel, "l", "Debug", "Logger level")
	flag.Parse()

	err := env.Parse(ops)
	if err != nil {
		log.Print("Failed to read environment variables")
	}
}
