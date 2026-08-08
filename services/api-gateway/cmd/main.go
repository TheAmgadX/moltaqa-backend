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

	"github.com/TheAmgadX/moltaqa-backend/services/api-gateway/internal/handlers"
	"github.com/TheAmgadX/moltaqa-backend/services/api-gateway/internal/middlewares"
	"github.com/TheAmgadX/moltaqa-backend/shared/env"
	shared_middlewares "github.com/TheAmgadX/moltaqa-backend/shared/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authpb "github.com/TheAmgadX/moltaqa-backend/shared/proto/auth"
	users "github.com/TheAmgadX/moltaqa-backend/shared/proto/users"
)

func addMiddlewares(router *chi.Mux) {
	router.Use(middleware.Logger)
	router.Use(shared_middlewares.CORSMiddleware)
}

func defineUsersRoutes(router *chi.Mux, handler *handlers.UserHandler) {
	// Global Middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.SetHeader("Content-Type", "application/json"))

	router.Route("/api/v1", func(r chi.Router) {

		// -------------------------------------------------------------
		// 1. AUTHENTICATION & ACCOUNT LIFECYCLE ROUTES
		// -------------------------------------------------------------
		r.Route("/auth", func(r chi.Router) {
			// Flow: User Creation / Login (POST /auth/login)
			r.Post("/login", handler.Login)

			// Flow: Verify OTP for Login, Register, Restore, or Delete
			r.Post("/verify-otp", handler.VerifyOTP)

			// Flow: Refresh Access Token
			r.Post("/refresh", handler.RefreshToken)
		})

		// -------------------------------------------------------------
		// 2. USER PROFILE & PRIVACY ROUTES
		// -------------------------------------------------------------
		r.Route("/users", func(r chi.Router) {

			// Protected Routes (Require Auth Middleware)
			r.Group(func(r chi.Router) {
				r.Use(middlewares.Authentication)

				// Flow: Register Contact
				r.Post("/contacts", handler.RegisterContact)

				// Flow: Update User Profile (Username, Display Name, Bio, etc.)
				r.Patch("/", handler.UpdateUserProfile)

				// Flow: Request User Account Deletion
				r.Delete("/", handler.DeleteUserAccount)

				// Flow: Profile Image Upload
				r.Post("/avatar", handler.UploadProfileImage)

				// Flow: Request Account Restoration
				r.Post("/restore", handler.Restore)

				// Privacy Settings Routes
				r.Route("/privacy-settings", func(r chi.Router) {
					r.Get("/", handler.GetPrivacySettings)

					// Flow: Update Privacy Settings (Search, Profile, Chat)
					r.Patch("/", handler.UpdatePrivacySettings)
				})
			})

			// Flow: Get User
			r.Get("/", handler.GetUser)

			// Flow: Check if User Exists
			r.Get("/exists", handler.UserExists)

			// Flow: Search For users
			r.Post("/search", handler.SearchUsers)

			// Flow: Get User Summary
			r.Get("/{id}/summary", handler.GetUserSummary)

			// Flow: Get Users Summaries
			r.Post("/summaries", handler.GetUsersSummary)
		})
	})
}

func createServer(userGrpcConn *grpc.ClientConn, authGrpcConn *grpc.ClientConn) *http.Server {
	router := chi.NewRouter()

	addMiddlewares(router)

	usersClient := users.NewUsersServiceClient(userGrpcConn)
	authClient := authpb.NewAuthServiceClient(authGrpcConn)

	usersHandler := handlers.NewUserHandler(usersClient, authClient)

	defineUsersRoutes(router, usersHandler)

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

func createGRPCConnection(serviceUrl string) (*grpc.ClientConn, error) {
	grpcConn, err := grpc.NewClient(serviceUrl,
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}
	return grpcConn, nil
}

func main() {
	log.Println("Start api-gateway service...")

	userService := env.GetString("USER_SERVICE_URL", "localhost")
	userGrpcConn, err := createGRPCConnection(userService)
	if err != nil {
		log.Fatalf("failed to connect to user grpc server: %v", err)
	}
	defer userGrpcConn.Close()

	authService := env.GetString("AUTH_SERVICE_URL", "localhost")
	authGrpcConn, err := createGRPCConnection(authService)
	if err != nil {
		log.Fatalf("failed to connect to auth grpc server: %v", err)
	}
	defer authGrpcConn.Close()

	server := createServer(userGrpcConn, authGrpcConn)

	if err := runServer(server, 10*time.Second, context.Background()); err != nil {
		log.Printf("Failed while running the server: %v", err)
		return
	}
}
