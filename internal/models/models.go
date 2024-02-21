package models

type CreateShortenRequest struct {
	URL string `json:"url"`
}

type CreateShortenResponse struct {
	URL string `json:"result"`
}

type URLRecord struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
