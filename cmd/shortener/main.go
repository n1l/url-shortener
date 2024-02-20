package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/n1l/url-shortener/internal/config"
	"github.com/n1l/url-shortener/internal/di"
	"github.com/n1l/url-shortener/internal/logger"
	"github.com/n1l/url-shortener/internal/storage"
	"github.com/n1l/url-shortener/internal/zipper"
)

func serverHandler(services *di.Services) http.Handler {
	router := chi.NewRouter()
	router.Use(logger.RequestLoggerMiddleware)
	router.Use(zipper.GzipMiddleware)

	router.Post("/api/shorten", services.CreateShortedURLfromJSONHandler)
	router.Post("/", services.CreateShortedURLHandler)
	router.Get("/{id}", services.GetURLByHashHandler)

	return router
}

func main() {
	var options config.Options
	config.ParseOptions(&options)
	logger.Initialize(options.LogLevel)

	fstorage, err := storage.NewFileStorage(options.StoragePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fstorage.Close()

	services := di.NewServices(&options, fstorage, fstorage)

	server := &http.Server{Addr: options.PrivateHost, Handler: serverHandler(services)}

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

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-serverCtx.Done()
}
