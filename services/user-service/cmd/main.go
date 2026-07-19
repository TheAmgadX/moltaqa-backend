package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	user_grpc "github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/infrastructure/grpc"
	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/service"
	"github.com/TheAmgadX/moltaqa-backend/shared/env"
	pb "github.com/TheAmgadX/moltaqa-backend/shared/proto/users"
	"google.golang.org/grpc"
)

func createServer(port string) (*grpc.Server, *net.Listener, error) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Printf("failed to listen to tcp server in port %s : %v", ":"+port, err)
		return nil, nil, err
	}

	grpc_server := grpc.NewServer()

	service := service.NewService()

	pb.RegisterUsersServiceServer(grpc_server, user_grpc.NewUserGRPCServer(service))

	return grpc_server, &lis, nil
}

func gracefulShutdown(grpcServer *grpc.Server, shutdownTimeout time.Duration) {
	done := make(chan struct{})

	go func() {
		log.Println("Gracefully stopping gRPC server...")
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Println("gRPC server stopped gracefully.")

	case <-time.After(shutdownTimeout):
		log.Println("Graceful shutdown timed out, forcing stop.")
		grpcServer.Stop()
	}
}

func RunServer(grpcServer *grpc.Server, lis *net.Listener, ctx context.Context, shutdownTimeout time.Duration) error {
	serverErrChan := make(chan error, 1)

	go func() {
		log.Printf("gRPC server listening on %s\n", (*lis).Addr())

		if err := grpcServer.Serve(*lis); err != nil {
			serverErrChan <- err
		}

		close(serverErrChan)
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	select {

	case <-stopChan:
		log.Println("Received shutdown signal.")

	case err := <-serverErrChan:
		return err

	case <-ctx.Done():
		log.Println("Context cancelled.")
	}

	gracefulShutdown(grpcServer, shutdownTimeout)

	return nil
}

func main() {
	log.Println("Start user service.")

	port := env.GetString("GRPC_PORT", "")

	if port == "" {
		log.Println("Couldn't read the port from environment variables.")
		return
	}

	grpcServer, lis, err := createServer(port)
	if err != nil {
		log.Printf("failed to create server: %v", err)
		return
	}

	if err := RunServer(grpcServer, lis, context.Background(), 10*time.Second); err != nil {
		log.Printf("failed to run grpc server: %v", err)
		return
	}

	log.Println("Shutdown user service.")
}
