package protocol

// Mock generation for generated HealthService clients (Connect-go and gRPC)
//
// Run via:
//   - make generate-mocks
//   - or: go generate ./...

//go:generate go run go.uber.org/mock/mockgen -destination=../../../mocks/mock_connect_go_health_service_client.go -package=mocks -mock_names=HealthServiceClient=MockConnectGoHealthServiceClient github.com/seventeenthearth/sudal/gen/go/health/v1/healthv1connect HealthServiceClient

//go:generate go run go.uber.org/mock/mockgen -destination=../../../mocks/mock_grpc_health_service_client.go -package=mocks -mock_names=HealthServiceClient=MockGRPCHealthServiceClient github.com/seventeenthearth/sudal/gen/go/health/v1 HealthServiceClient
