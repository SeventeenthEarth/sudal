package mocks

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/onsi/gomega"
	healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
	"github.com/seventeenthearth/sudal/gen/go/health/v1/healthv1connect"
	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
	healthInterface "github.com/seventeenthearth/sudal/internal/feature/health/interface"
	healthConnect "github.com/seventeenthearth/sudal/internal/feature/health/interface/connect"
	internalMocks "github.com/seventeenthearth/sudal/internal/mocks"
	"go.uber.org/mock/gomock"
)

// IntegrationTestContext holds the context for integration tests
type IntegrationTestContext struct {
	// Server components
	Server   *http.Server
	Listener net.Listener
	BaseURL  string

	// Mocks
	MockRepo     *internalMocks.MockHealthRepository
	MockService  *MockService
	MockPostgres *internalMocks.MockPostgresManager
	MockRedis    *MockRedisManager
	MockCache    *MockCacheUtil

	// Mock controller for gomock
	MockCtrl *gomock.Controller

	// Handlers
	HealthHandler        *healthInterface.Handler
	HealthConnectHandler *healthConnect.HealthServiceHandler

	// Clients
	ConnectGoClient *MockConnectGoClient
	GRPCClient      *MockGRPCClient
	HTTPClient      *http.Client

	// Test utilities
	ConcurrentTester *MockConcurrentTester

	// Configuration
	TestTimeout time.Duration
	Protocol    string
}

// NewIntegrationTestContext creates a new integration test context
func NewIntegrationTestContext() *IntegrationTestContext {
	ctrl := gomock.NewController(nil) // Will be set properly in test setup
	mockRedis := NewMockRedisManager()
	return &IntegrationTestContext{
		MockCtrl:         ctrl,
		MockRepo:         internalMocks.NewMockHealthRepository(ctrl),
		MockService:      NewMockService(),
		MockPostgres:     internalMocks.NewMockPostgresManager(ctrl),
		MockRedis:        mockRedis,
		MockCache:        NewMockCacheUtilWithRedis(mockRedis),
		ConcurrentTester: NewMockConcurrentTester(),
		HTTPClient:       &http.Client{Timeout: 10 * time.Second},
		TestTimeout:      5 * time.Second,
		Protocol:         "http",
	}
}

// SetupTestServer sets up a test server with mocked dependencies
func (ctx *IntegrationTestContext) SetupTestServer() error {
	// Create service with mock repository
	service := application.NewService(ctx.MockRepo)

	// Create handlers
	ctx.HealthHandler = healthInterface.NewHandler(service)
	ctx.HealthConnectHandler = healthConnect.NewHealthServiceHandler(service)

	// Create router
	mux := http.NewServeMux()

	// Register REST routes
	ctx.HealthHandler.RegisterRoutes(mux)

	// Register Connect-Go routes
	path, handler := healthv1connect.NewHealthServiceHandler(ctx.HealthConnectHandler)
	mux.Handle(path, handler)

	// Start test server
	var err error
	ctx.Listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	addr := ctx.Listener.Addr().String()
	ctx.BaseURL = "http://" + addr

	ctx.Server = &http.Server{
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		_ = ctx.Server.Serve(ctx.Listener)
	}()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	return nil
}

// TeardownTestServer tears down the test server
func (ctx *IntegrationTestContext) TeardownTestServer() error {
	if ctx.Server != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := ctx.Server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
	}

	if ctx.Listener != nil {
		if err := ctx.Listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	return nil
}

// SetupConnectGoClient sets up a Connect-Go client with specified protocol
func (ctx *IntegrationTestContext) SetupConnectGoClient(protocol string) {
	ctx.Protocol = protocol
	ctx.ConnectGoClient = NewMockConnectGoClient(protocol)
}

// SetupGRPCClient sets up a native gRPC client
func (ctx *IntegrationTestContext) SetupGRPCClient() {
	ctx.GRPCClient = NewMockGRPCClient()
}

// ConfigureMockForHealthyState configures all mocks for healthy state
func (ctx *IntegrationTestContext) ConfigureMockForHealthyState() {
	// Configure repository mock to return healthy status
	ctx.MockRepo.EXPECT().GetStatus(gomock.Any()).Return(entity.OkStatus(), nil).AnyTimes()
	ctx.MockRepo.EXPECT().GetDatabaseStatus(gomock.Any()).Return(&entity.DatabaseStatus{
		Status:  "healthy",
		Message: "Database connection is healthy",
		Stats: &entity.ConnectionStats{
			MaxOpenConnections: 25,
			OpenConnections:    5,
			InUse:              2,
			Idle:               3,
			WaitCount:          0,
			WaitDuration:       0,
			MaxIdleClosed:      0,
			MaxLifetimeClosed:  0,
		},
	}, nil).AnyTimes()

	ctx.MockService.PingFunc = func(context.Context) (*entity.HealthStatus, error) {
		return entity.OkStatus(), nil
	}
	ctx.MockService.CheckFunc = func(context.Context) (*entity.HealthStatus, error) {
		return entity.HealthyStatus(), nil
	}
	ctx.MockService.CheckDatabaseFunc = func(context.Context) (*entity.DatabaseStatus, error) {
		stats := &entity.ConnectionStats{
			MaxOpenConnections: 25,
			OpenConnections:    5,
			InUse:              2,
			Idle:               3,
			WaitCount:          0,
			WaitDuration:       0,
			MaxIdleClosed:      0,
			MaxLifetimeClosed:  0,
		}
		return entity.HealthyDatabaseStatus("Database is healthy", stats), nil
	}

	if ctx.ConnectGoClient != nil {
		ctx.ConnectGoClient.SetServingStatus()
	}

	if ctx.GRPCClient != nil {
		ctx.GRPCClient.SetServingStatus()
	}
}

// ConfigureMockForUnhealthyState configures all mocks for unhealthy state
func (ctx *IntegrationTestContext) ConfigureMockForUnhealthyState(err error) {
	// Configure repository mock to return unhealthy status
	ctx.MockRepo.EXPECT().GetStatus(gomock.Any()).Return(nil, err).AnyTimes()
	ctx.MockRepo.EXPECT().GetDatabaseStatus(gomock.Any()).Return(nil, err).AnyTimes()

	ctx.MockService.PingFunc = func(context.Context) (*entity.HealthStatus, error) {
		return nil, err
	}
	ctx.MockService.CheckFunc = func(context.Context) (*entity.HealthStatus, error) {
		return nil, err
	}
	ctx.MockService.CheckDatabaseFunc = func(context.Context) (*entity.DatabaseStatus, error) {
		return nil, err
	}

	if ctx.ConnectGoClient != nil {
		ctx.ConnectGoClient.SetError(err)
	}

	if ctx.GRPCClient != nil {
		ctx.GRPCClient.SetError(err)
	}
}

// ConfigureMockForDegradedState configures mocks for degraded state
func (ctx *IntegrationTestContext) ConfigureMockForDegradedState() {
	// Configure repository mock to return degraded status
	ctx.MockRepo.EXPECT().GetStatus(gomock.Any()).Return(entity.NewHealthStatus("degraded"), nil).AnyTimes()
	ctx.MockRepo.EXPECT().GetDatabaseStatus(gomock.Any()).Return(&entity.DatabaseStatus{
		Status:  "degraded",
		Message: "Database connection is degraded",
	}, nil).AnyTimes()

	if ctx.ConnectGoClient != nil {
		ctx.ConnectGoClient.SetNotServingStatus()
	}

	if ctx.GRPCClient != nil {
		ctx.GRPCClient.SetNotServingStatus()
	}
}

// RunConcurrentConnectGoRequests runs concurrent Connect-Go requests for testing
func (ctx *IntegrationTestContext) RunConcurrentConnectGoRequests(numRequests int) {
	ctx.ConcurrentTester.Clear()

	var wg sync.WaitGroup
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func(requestID int) {
			defer wg.Done()

			start := time.Now()

			// Create a unique client for this request to avoid race conditions
			client := NewMockConnectGoClient(ctx.Protocol)
			client.SetServingStatus()

			// Make request
			reqCtx, cancel := context.WithTimeout(context.Background(), ctx.TestTimeout)
			defer cancel()

			req := connect.NewRequest(&healthv1.CheckRequest{})
			resp, err := client.Check(reqCtx, req)

			duration := time.Since(start)

			result := ConcurrentTestResult{
				Success:  err == nil && resp != nil,
				Error:    err,
				Duration: duration,
				Protocol: ctx.Protocol,
				Metadata: make(map[string]string),
			}

			if resp != nil {
				for key, values := range resp.Header() {
					if len(values) > 0 {
						result.Metadata[key] = values[0]
					}
				}
			}

			ctx.ConcurrentTester.AddResult(result)
		}(i)
	}

	wg.Wait()
}

// RunConcurrentGRPCRequests runs concurrent native gRPC requests for testing
func (ctx *IntegrationTestContext) RunConcurrentGRPCRequests(numRequests int) {
	ctx.ConcurrentTester.Clear()

	var wg sync.WaitGroup
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func(requestID int) {
			defer wg.Done()

			start := time.Now()

			// Create a unique client for this request
			client := NewMockGRPCClient()
			client.SetServingStatus()

			// Make request
			reqCtx, cancel := context.WithTimeout(context.Background(), ctx.TestTimeout)
			defer cancel()

			req := &healthv1.CheckRequest{}
			resp, err := client.Check(reqCtx, req)

			duration := time.Since(start)

			result := ConcurrentTestResult{
				Success:  err == nil && resp != nil,
				Error:    err,
				Duration: duration,
				Protocol: "grpc",
				Metadata: make(map[string]string),
			}

			ctx.ConcurrentTester.AddResult(result)
		}(i)
	}

	wg.Wait()
}

// AssertAllRequestsSucceeded asserts that all concurrent requests succeeded
func (ctx *IntegrationTestContext) AssertAllRequestsSucceeded() {
	results := ctx.ConcurrentTester.GetResults()
	successCount := ctx.ConcurrentTester.GetSuccessCount()
	errorCount := ctx.ConcurrentTester.GetErrorCount()

	gomega.Expect(errorCount).To(gomega.Equal(0),
		fmt.Sprintf("Expected no errors, but got %d errors out of %d requests", errorCount, len(results)))
	gomega.Expect(successCount).To(gomega.Equal(len(results)),
		fmt.Sprintf("Expected all %d requests to succeed, but only %d succeeded", len(results), successCount))
}

// AssertResponseConsistency asserts that responses are consistent across protocols
func (ctx *IntegrationTestContext) AssertResponseConsistency() {
	results := ctx.ConcurrentTester.GetResults()

	if len(results) == 0 {
		return
	}

	// Group results by protocol
	protocolResults := make(map[string][]ConcurrentTestResult)
	for _, result := range results {
		protocolResults[result.Protocol] = append(protocolResults[result.Protocol], result)
	}

	// Assert that all protocols have consistent success rates
	for protocol, protocolSpecificResults := range protocolResults {
		successCount := 0
		for _, result := range protocolSpecificResults {
			if result.Success {
				successCount++
			}
		}

		successRate := float64(successCount) / float64(len(protocolSpecificResults))
		gomega.Expect(successRate).To(gomega.BeNumerically(">=", 0.95),
			fmt.Sprintf("Protocol %s should have at least 95%% success rate, got %.2f%%", protocol, successRate*100))
	}
}
