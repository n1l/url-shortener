package models

type CreateShortenRequest struct {
	URL string `json:"url"`
}

type CreateShortenResponse struct {
	URL string `json:"result"`
}
