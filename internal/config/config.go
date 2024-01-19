package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type Options struct {
	PrivateHost string `env:"SERVER_ADDRESS"`
	PublicHost  string `env:"BASE_URL"`
}

func ParseOptions(ops *Options) {
	flag.StringVar(&ops.PrivateHost, "a", "localhost:8080", "The service address at start")
	flag.StringVar(&ops.PublicHost, "b", "http://localhost:8080", "The shortener result base address")
	flag.Parse()

	err := env.Parse(ops)
	if err != nil {
		log.Print("Failed to read environment variables")
	}
}
