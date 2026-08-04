package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TheAmgadX/moltaqa-backend/services/Auth-service/internal/auth"
	cache_redis "github.com/TheAmgadX/moltaqa-backend/services/Auth-service/internal/infrastructure/cache/redis"
	Auth_events "github.com/TheAmgadX/moltaqa-backend/services/Auth-service/internal/infrastructure/events"
	Auth_grpc "github.com/TheAmgadX/moltaqa-backend/services/Auth-service/internal/infrastructure/grpc"
	repository "github.com/TheAmgadX/moltaqa-backend/services/Auth-service/internal/infrastructure/repository/postgres"
	"github.com/TheAmgadX/moltaqa-backend/services/Auth-service/internal/service"
	"github.com/TheAmgadX/moltaqa-backend/shared/env"
	"github.com/TheAmgadX/moltaqa-backend/shared/kafka"
	pb "github.com/TheAmgadX/moltaqa-backend/shared/proto/auth"
	userspb "github.com/TheAmgadX/moltaqa-backend/shared/proto/users"
	goredis "github.com/redis/go-redis/v9"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func build_DB_DSN() string {
	host := env.GetString("DB_HOST", "localhost")
	port := env.GetString("DB_PORT", "5432")
	user := env.GetString("DB_USER", "postgres")
	pass := env.GetString("DB_PASSWORD", "postgres")
	dbName := env.GetString("DB_NAME", "postgres")

	// postgres://<user>:<password>@<host>:<port>/<dbname>?sslmode=disable
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, host, port, dbName,
	)
}

const (
	serviceId = "auth-service-kafka-client"
)

func createKafkaClient() (*kgo.Client, error) {
	cfg := kafka.NewConfig(serviceId, "")

	client, err := kafka.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func createRedisClient() (*goredis.Client, error) {
	addr := fmt.Sprintf(
		"%s:%s",
		env.GetString("REDIS_HOST", "localhost"),
		env.GetString("REDIS_PORT", "6379"),
	)

	client := goredis.NewClient(&goredis.Options{
		Addr: addr,
		DB:   env.GetInt("REDIS_DB", 0),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}

func createUsersClient() (userspb.UsersServiceClient, *grpc.ClientConn, error) {
	addr := env.GetString("USER_SERVICE_ADDR", "localhost:50052")

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	return userspb.NewUsersServiceClient(conn), conn, nil
}

func createServer(port string) (*grpc.Server, *net.Listener, func(), error) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Printf("failed to listen to tcp server in port %s : %v", ":"+port, err)
		return nil, nil, nil, err
	}

	grpc_server := grpc.NewServer()

	repo, err := repository.NewAuthPostgresRepository(build_DB_DSN())
	if err != nil {
		log.Printf("failed to create repository: %v\n", err)
		return nil, nil, nil, err
	}

	redisClient, err := createRedisClient()
	if err != nil {
		log.Printf("failed to create redis client: %v\n", err)
		return nil, nil, nil, err
	}

	kafkaClient, err := createKafkaClient()
	if err != nil {
		log.Printf("failed to create kafka client: %v\n", err)
		redisClient.Close()
		return nil, nil, nil, err
	}

	usersClient, usersConn, err := createUsersClient()
	if err != nil {
		log.Printf("failed to create users client: %v\n", err)
		kafkaClient.Close()
		redisClient.Close()
		return nil, nil, nil, err
	}

	otpStore := cache_redis.NewOTPTransactionStore(redisClient, "")
	publisher := Auth_events.NewPublisher(kafka.NewProducer(kafkaClient))
	signer := auth.NewJWTSigner()

	service, err := service.NewService(repo, otpStore, publisher, signer, usersClient)
	if err != nil {
		log.Printf("failed to create service: %v\n", err)
		usersConn.Close()
		kafkaClient.Close()
		redisClient.Close()
		return nil, nil, nil, err
	}

	pb.RegisterAuthServiceServer(grpc_server, Auth_grpc.NewAuthGRPCServer(service))

	cleanup := func() {
		usersConn.Close()
		kafkaClient.Close()
		redisClient.Close()
	}

	return grpc_server, &lis, cleanup, nil
}

func gracefulShutdown(grpcServer *grpc.Server, cleanup func(), shutdownTimeout time.Duration) {
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

	if cleanup != nil {
		cleanup()
	}
}

func RunServer(grpcServer *grpc.Server, lis *net.Listener, cleanup func(), ctx context.Context, shutdownTimeout time.Duration) error {
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
		if cleanup != nil {
			cleanup()
		}
		return err

	case <-ctx.Done():
		log.Println("Context cancelled.")
	}

	gracefulShutdown(grpcServer, cleanup, shutdownTimeout)

	return nil
}

func main() {
	log.Println("Start auth service.")

	port := env.GetString("GRPC_PORT", "")

	if port == "" {
		log.Println("Couldn't read the port from environment variables.")
		return
	}

	grpcServer, lis, cleanup, err := createServer(port)
	if err != nil {
		log.Printf("failed to create server: %v", err)
		return
	}

	if err := RunServer(grpcServer, lis, cleanup, context.Background(), 10*time.Second); err != nil {
		log.Printf("failed to run grpc server: %v", err)
		return
	}

	log.Println("Shutdown auth service.")
}
