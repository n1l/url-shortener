package main

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
)

var shortedUrls map[string]string = make(map[string]string)

func getHashOfUrl(url string) string {
	sum := md5.Sum([]byte(url))
	encoded := base64.StdEncoding.EncodeToString(sum[:])
	return strings.Replace(encoded, "/", "", -1)[:8]
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

func service() http.Handler {
	logger := httplog.NewLogger("url-shortener-logger", httplog.Options{
		LogLevel:         slog.LevelDebug,
		Concise:          true,
		RequestHeaders:   true,
		MessageFieldName: "message",
	})

	router := chi.NewRouter()
	router.Use(httplog.RequestLogger(logger))
	router.Post("/", CreateShortedUrlHandler)
	router.Get("/{id}", GetUrlByHash)

	return router
}

func main() {
	server := &http.Server{Addr: "0.0.0.0:8080", Handler: service()}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

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

	// Run the server
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
