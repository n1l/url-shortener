package di

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/n1l/url-shortener/internal/hasher"
	"github.com/n1l/url-shortener/internal/models"
)

func (s *Services) GetURLByHashHandler(w http.ResponseWriter, r *http.Request) {
	const parameterName = "id"

	if r.Method != http.MethodGet {
		http.Error(w, "Bad Request!", http.StatusBadRequest)
		return
	}

	urlGetter := s.URLGetter
	hashID := chi.URLParam(r, parameterName)

	if url, ok := urlGetter.Get(hashID); ok {
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return
	}

	http.Error(w, fmt.Sprintf("Bad Request! id: '%s' not found", hashID), http.StatusBadRequest)
}

func (s *Services) CreateShortedURLHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Bad Request!", http.StatusBadRequest)
		return
	}

	body, readErr := io.ReadAll(r.Body)
	if readErr != nil {
		http.Error(w, "Bad Request!", http.StatusBadRequest)
		return
	}

	stringURI := string(body)
	if _, parseErr := url.ParseRequestURI(stringURI); parseErr != nil {
		http.Error(w, "Bad Request!"+" "+parseErr.Error(), http.StatusBadRequest)
		return
	}

	ops := s.Options
	urlSaver := s.URLSaver
	hashID := hasher.GetHashOfURL(stringURI)

	rec := &models.URLRecord{
		ShortURL:    hashID,
		OriginalURL: stringURI,
	}
	urlSaver.Save(rec)

	resultStr := fmt.Sprintf("%s/%s", ops.PublicHost, hashID)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resultStr))
}

func (s *Services) CreateShortedURLfromJSONHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Bad Request!", http.StatusBadRequest)
		return
	}

	var req models.CreateShortenRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ops := s.Options
	urlSaver := s.URLSaver
	hashID := hasher.GetHashOfURL(req.URL)

	rec := &models.URLRecord{
		ShortURL:    hashID,
		OriginalURL: req.URL,
	}
	urlSaver.Save(rec)

	resultStr := fmt.Sprintf("%s/%s", ops.PublicHost, hashID)

	resp := models.CreateShortenResponse{
		URL: resultStr,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
