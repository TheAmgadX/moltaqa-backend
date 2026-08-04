package grpc_test

import (
	"testing"
	"time"

	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/infrastructure/repository"
	grpcpkg "github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/infrastructure/grpc"
	servicepkg "github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/service"
	"github.com/twmb/franz-go/pkg/kgo"
)

// mustNewGRPCServerWithRepo is like mustNewGRPCServer but accepts any
// repository.UserRepository, allowing specialised fakes (e.g. lookupRecorderRepo)
// without casting.
func mustNewGRPCServerWithRepo(t *testing.T, repo repository.UserRepository) *grpcpkg.UserGRPCServer {
	t.Helper()
	client, err := kgo.NewClient(
		kgo.SeedBrokers("127.0.0.1:9092"),
		kgo.RecordDeliveryTimeout(1*time.Second),
		kgo.RequestTimeoutOverhead(100*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("failed to construct kafka client: %v", err)
	}
	t.Cleanup(client.Close)

	svc, err := servicepkg.NewService(repo, client)
	if err != nil {
		t.Fatalf("failed to construct service: %v", err)
	}
	return grpcpkg.NewUserGRPCServer(svc)
}
