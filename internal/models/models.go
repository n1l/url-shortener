package models

// Request описывает запрос пользователя.
// см. https://yandex.ru/dev/dialogs/alice/doc/request.html
type CreateShortenRequest struct {
	Url string `json:"url"`
}

type CreateShortenResponse struct {
	Url string `json:"result"`
}
