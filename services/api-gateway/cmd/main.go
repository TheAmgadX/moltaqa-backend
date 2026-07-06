package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TheAmgadX/moltaqa-backend/shared/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func addMiddlewares(router *chi.Mux) {
	router.Use(middleware.Logger)
	router.Use(middlewares.CORSMiddleware)
}

func defineRoutes(router *chi.Mux) {
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
}

func createServer() *http.Server {
	router := chi.NewRouter()

	addMiddlewares(router)

	defineRoutes(router)

	return &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
}

func gracefulShutdown(server *http.Server, shutdownTimeout time.Duration) error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		// server graceful shutdown has failed, close the server `Not Gracefully`
		log.Printf("failed during shutting down the server gracefully: %v\n", err)

		if closeErr := server.Close(); closeErr != nil {
			log.Printf("failed during closing the server: %v\n", closeErr)
			return errors.Join(err, closeErr)
		}

		return err
	}

	log.Println("Server Graceful Shutdown done successfully.")

	return nil
}

func runServer(server *http.Server, shutdownTimeout time.Duration, ctx context.Context) error {
	serverErrChan := make(chan error, 1)

	// run the server
	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}

		close(serverErrChan)
	}()

	stopChan := make(chan os.Signal, 1)

	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// wait for signals
	select {
	case <-stopChan:
		log.Println("received a stop signal")

	case <-serverErrChan:
		log.Println("server error:", <-serverErrChan)

	case <-ctx.Done():
		log.Println("context done:", <-ctx.Done())
	}

	// handle server shutdown
	return gracefulShutdown(server, shutdownTimeout)
}

func main() {
	log.Println("Start api-gateway service...")

	server := createServer()

	if err := runServer(server, 10*time.Second, context.Background()); err != nil {
		log.Printf("Failed while running the server: %v", err)
		return
	}
}
