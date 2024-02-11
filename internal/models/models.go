package models

type CreateShortenRequest struct {
	Url string `json:"url"`
}

type CreateShortenResponse struct {
	Url string `json:"result"`
}
