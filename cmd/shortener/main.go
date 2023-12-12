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

func getHashOfURL(url string) string {
	sum := md5.Sum([]byte(url))
	encoded := base64.StdEncoding.EncodeToString(sum[:])
	return strings.Replace(encoded, "/", "", -1)[:8]
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
		http.Error(w, "Bad Request!", http.StatusBadRequest)
		return
	}

	hashID := getHashOfURL(stringURI)
	shortedUrls[hashID] = stringURI
	resultStr := fmt.Sprintf("http://%s/%s", r.Host, hashID)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resultStr))
}

func GetURLByHash(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Bad Request!", http.StatusBadRequest)
		return
	}

	hashID := chi.URLParam(r, "id")
	if url, ok := shortedUrls[hashID]; ok {
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
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
	router.Post("/", CreateShortedURLHandler)
	router.Get("/{id}", GetURLByHash)

	return router
}

func main() {
	server := &http.Server{Addr: "0.0.0.0:8080", Handler: service()}

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

	// Run the server
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
