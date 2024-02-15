package main

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/n1l/url-shortener/internal/config"
	"github.com/n1l/url-shortener/internal/database"
	"github.com/n1l/url-shortener/internal/logger"
	"github.com/n1l/url-shortener/internal/models"
)

var options config.Options

var shortedUrls map[string]string = make(map[string]string)
var dbProducer *database.Producer

func loadToMemory(fname string) error {
	recs, err := database.Load(fname)
	if err != nil {
		return err
	}

	for _, rec := range recs {
		shortedUrls[rec.ShortURL] = rec.OriginalURL
	}
	return nil
}

func getHashOfURLAndPersist(url string) string {
	sum := md5.Sum([]byte(url))
	encoded := base64.StdEncoding.EncodeToString(sum[:])
	hash := strings.Replace(encoded, "/", "", -1)[:8]
	saveInMemory(url, hash)
	saveOnDisk(url, hash)
	return hash
}

func saveInMemory(url string, hash string) {
	shortedUrls[hash] = url
}

func saveOnDisk(url string, hash string) error {
	urlReord := database.URLRecord{
		ShortURL:    hash,
		OriginalURL: url,
	}

	return dbProducer.Write(&urlReord)
}

func CreateShortedURLfromJSONHandler(w http.ResponseWriter, r *http.Request) {
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

	stringURI := req.Url
	hashID := getHashOfURLAndPersist(stringURI)
	resultStr := fmt.Sprintf("%s/%s", options.PublicHost, hashID)

	resp := models.CreateShortenResponse{
		Url: resultStr,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func CreateShortedURLHandler(w http.ResponseWriter, r *http.Request) {
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

	hashID := getHashOfURLAndPersist(stringURI)
	resultStr := fmt.Sprintf("%s/%s", options.PublicHost, hashID)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resultStr))
}

func GetURLByHashHandler(w http.ResponseWriter, r *http.Request) {
	const parameterName = "id"

	if r.Method != http.MethodGet {
		http.Error(w, "Bad Request!", http.StatusBadRequest)
		return
	}

	hashID := chi.URLParam(r, parameterName)
	if url, ok := shortedUrls[hashID]; ok {
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return
	}

	http.Error(w, fmt.Sprintf("Bad Request! id: '%s' not found", hashID), http.StatusBadRequest)
}

func gzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		h.ServeHTTP(ow, r)
	})
}

func serverHandler() http.Handler {
	router := chi.NewRouter()
	router.Use(logger.RequestLoggerMiddleware)
	router.Use(gzipMiddleware)

	router.Post("/api/shorten", CreateShortedURLfromJSONHandler)
	router.Post("/", CreateShortedURLHandler)
	router.Get("/{id}", GetURLByHashHandler)

	return router
}

func main() {
	config.ParseOptions(&options)
	logger.Initialize("Debug")

	server := &http.Server{Addr: options.PrivateHost, Handler: serverHandler()}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		shutdownCtx, cancelFunc := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancelFunc()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	err := loadToMemory(options.StoragePath)
	if err != nil {
		log.Fatal(err)
	}

	prod, err := database.NewProducer(options.StoragePath)
	if err != nil {
		log.Fatal(err)
	}
	dbProducer = prod

	// Run the server
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
