package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

var Options struct {
	PrivateHost string `env:"SERVER_ADDRESS"`
	PublicHost  string `env:"BASE_URL"`
}

func ParseOptions() {
	flag.StringVar(&Options.PrivateHost, "a", "localhost:8080", "The service address at start")
	flag.StringVar(&Options.PublicHost, "b", "localhost:8080", "The shortener result base address")
	flag.Parse()

	err := env.Parse(&Options)
	if err != nil {
		log.Print("Failed to read environment variables")
	}
}
