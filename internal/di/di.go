package di

import "github.com/n1l/url-shortener/internal/config"

type Services struct {
	URLSaver  URLSaver
	URLGetter URLGetter
	Options   *config.Options
}

func NewServices(options *config.Options, urlSaver URLSaver, urlGetter URLGetter) *Services {
	return &Services{
		Options:   options,
		URLSaver:  urlSaver,
		URLGetter: urlGetter,
	}
}
