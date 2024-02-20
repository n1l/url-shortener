package di

import "github.com/n1l/url-shortener/internal/models"

type URLSaver interface {
	Save(rec *models.URLRecord) error
}

type URLGetter interface {
	Get(hash string) (string, bool)
}
