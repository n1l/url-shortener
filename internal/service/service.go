package service

import "github.com/n1l/url-shortener/internal/config"

type Service struct {
	URLSaver  URLSaver
	URLGetter URLGetter
	Options   *config.Options
}

func NewService(options *config.Options, urlSaver URLSaver, urlGetter URLGetter) *Service {
	return &Service{
		Options:   options,
		URLSaver:  urlSaver,
		URLGetter: urlGetter,
	}
}
