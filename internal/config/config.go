package config

import (
	"flag"
)

var Options struct {
	PrivateHost string
	PublicHost  string
}

func ParseOptions() {
	flag.StringVar(&Options.PrivateHost, "a", "localhost:8080", "The service address at start")
	flag.StringVar(&Options.PublicHost, "b", "localhost:8080", "The shortener result base address")
	flag.Parse()
}
