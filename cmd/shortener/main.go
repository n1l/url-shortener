package main

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var shortedUrls map[string]string = make(map[string]string)

func getHashOfUrl(url string) string {
	sum := md5.Sum([]byte(url))
	encoded := base64.StdEncoding.EncodeToString(sum[:])

	return encoded[:8]
}

func CreateShortedUrlHandler(w http.ResponseWriter, r *http.Request) {
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

	stringUri := string(body)
	if _, parseErr := url.ParseRequestURI(stringUri); parseErr != nil {
		http.Error(w, "Bad Request!", http.StatusBadRequest)
		return
	}

	hashId := getHashOfUrl(stringUri)
	shortedUrls[hashId] = stringUri
	resultStr := fmt.Sprintf("http://%s/%s", r.Host, hashId)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resultStr))
}

func GetUrlByHash(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Bad Request!", http.StatusBadRequest)
		return
	}

	hashId := chi.URLParam(r, "id")
	if url, ok := shortedUrls[hashId]; ok {
		http.Redirect(w, r, url, http.StatusPermanentRedirect)
		return
	}

	http.Error(w, "Bad Request!", http.StatusBadRequest)
}

func main() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Post("/", CreateShortedUrlHandler)
	router.Get("/{id}", GetUrlByHash)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		panic(err)
	}
}
