package helpers

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

// ConcurrentTestResult represents the result of a concurrent test operation
type ConcurrentTestResult struct {
	Success  bool
	Error    error
	Duration time.Duration
	Protocol string
	Metadata map[string]string
}

// ConcurrentTestHelper provides utilities for testing concurrent operations
type ConcurrentTestHelper struct {
	Results []ConcurrentTestResult
	mutex   sync.Mutex
}

// NewConcurrentTestHelper creates a new concurrent test helper
func NewConcurrentTestHelper() *ConcurrentTestHelper {
	return &ConcurrentTestHelper{
		Results: make([]ConcurrentTestResult, 0),
	}
}

// AddResult adds a test result (thread-safe)
func (h *ConcurrentTestHelper) AddResult(result ConcurrentTestResult) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.Results = append(h.Results, result)
}

// GetResults returns all test results (thread-safe)
func (h *ConcurrentTestHelper) GetResults() []ConcurrentTestResult {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	resultsCopy := make([]ConcurrentTestResult, len(h.Results))
	copy(resultsCopy, h.Results)
	return resultsCopy
}

// GetSuccessCount returns the number of successful operations
func (h *ConcurrentTestHelper) GetSuccessCount() int {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	count := 0
	for _, result := range h.Results {
		if result.Success {
			count++
		}
	}
	return count
}

// GetErrorCount returns the number of failed operations
func (h *ConcurrentTestHelper) GetErrorCount() int {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	count := 0
	for _, result := range h.Results {
		if !result.Success {
			count++
		}
	}
	return count
}

// Clear clears all results
func (h *ConcurrentTestHelper) Clear() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.Results = h.Results[:0]
}

// ProtocolTestScenarioHelper represents a test scenario for different protocols
type ProtocolTestScenarioHelper struct {
	Name           string
	Protocol       string
	ExpectedStatus healthv1.ServingStatus
	ShouldSucceed  bool
	Metadata       map[string]string
	Timeout        time.Duration
}

// NewProtocolTestScenarioHelper creates a new protocol test scenario helper
func NewProtocolTestScenarioHelper(name, protocol string, expectedStatus healthv1.ServingStatus, shouldSucceed bool) ProtocolTestScenarioHelper {
	return ProtocolTestScenarioHelper{
		Name:           name,
		Protocol:       protocol,
		ExpectedStatus: expectedStatus,
		ShouldSucceed:  shouldSucceed,
		Metadata:       make(map[string]string),
		Timeout:        5 * time.Second,
	}
}

// AddMetadata adds metadata to the test scenario
func (p *ProtocolTestScenarioHelper) AddMetadata(key, value string) {
	p.Metadata[key] = value
}

// SetTimeout sets the timeout for the test scenario
func (p *ProtocolTestScenarioHelper) SetTimeout(timeout time.Duration) {
	p.Timeout = timeout
}

// IntegrationTestContext holds the context for integration tests
type IntegrationTestContext struct {
	// Server components
	Server   *http.Server
	Listener net.Listener
	BaseURL  string

	// Mocks
	MockRepo          *internalMocks.MockHealthRepository
	MockHealthService *internalMocks.MockHealthService
	MockPostgres      *internalMocks.MockPostgresManager
	MockRedis         *internalMocks.MockRedisManager
	MockCache         *internalMocks.MockCacheUtil

	// Mock controller for gomock
	MockCtrl *gomock.Controller

	// Handlers
	HealthHandler        *healthInterface.Handler
	HealthConnectHandler *healthConnect.HealthServiceHandler

	// Clients
	ConnectGoClientHelper *ConnectGoMockHelper
	GRPCClientHelper      *GRPCMockHelper
	HTTPClient            *http.Client

	// Test utilities
	ConcurrentTester *ConcurrentTestHelper

	// Configuration
	TestTimeout time.Duration
	Protocol    string
}

// NewIntegrationTestContext creates a new integration test context
func NewIntegrationTestContext() *IntegrationTestContext {
	ctrl := gomock.NewController(nil) // Will be set properly in test setup
	mockRedis := internalMocks.NewMockRedisManager(ctrl)
	return &IntegrationTestContext{
		MockCtrl:          ctrl,
		MockRepo:          internalMocks.NewMockHealthRepository(ctrl),
		MockHealthService: internalMocks.NewMockHealthService(ctrl),
		MockPostgres:      internalMocks.NewMockPostgresManager(ctrl),
		MockRedis:         mockRedis,
		MockCache:         internalMocks.NewMockCacheUtil(ctrl),
		ConcurrentTester:  NewConcurrentTestHelper(),
		HTTPClient:        &http.Client{Timeout: 10 * time.Second},
		TestTimeout:       5 * time.Second,
		Protocol:          "http",
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
	ctx.ConnectGoClientHelper = NewConnectGoMockHelper(ctx.MockCtrl, protocol)
}

// SetupGRPCClient sets up a native gRPC client
func (ctx *IntegrationTestContext) SetupGRPCClient() {
	ctx.GRPCClientHelper = NewGRPCMockHelper(ctx.MockCtrl)
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

	ctx.MockHealthService.EXPECT().Ping(gomock.Any()).Return(entity.OkStatus(), nil).AnyTimes()
	ctx.MockHealthService.EXPECT().Check(gomock.Any()).Return(entity.HealthyStatus(), nil).AnyTimes()
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
	ctx.MockHealthService.EXPECT().CheckDatabase(gomock.Any()).Return(entity.HealthyDatabaseStatus("Database is healthy", stats), nil).AnyTimes()

	if ctx.ConnectGoClientHelper != nil {
		ctx.ConnectGoClientHelper.SetServingStatus()
	}

	if ctx.GRPCClientHelper != nil {
		ctx.GRPCClientHelper.SetServingStatus()
	}
}

// ConfigureMockForUnhealthyState configures all mocks for unhealthy state
func (ctx *IntegrationTestContext) ConfigureMockForUnhealthyState(err error) {
	// Configure repository mock to return unhealthy status
	ctx.MockRepo.EXPECT().GetStatus(gomock.Any()).Return(nil, err).AnyTimes()
	ctx.MockRepo.EXPECT().GetDatabaseStatus(gomock.Any()).Return(nil, err).AnyTimes()

	ctx.MockHealthService.EXPECT().Ping(gomock.Any()).Return(nil, err).AnyTimes()
	ctx.MockHealthService.EXPECT().Check(gomock.Any()).Return(nil, err).AnyTimes()
	ctx.MockHealthService.EXPECT().CheckDatabase(gomock.Any()).Return(nil, err).AnyTimes()

	if ctx.ConnectGoClientHelper != nil {
		ctx.ConnectGoClientHelper.SetError(err)
	}

	if ctx.GRPCClientHelper != nil {
		ctx.GRPCClientHelper.SetError(err)
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

	if ctx.ConnectGoClientHelper != nil {
		ctx.ConnectGoClientHelper.SetNotServingStatus()
	}

	if ctx.GRPCClientHelper != nil {
		ctx.GRPCClientHelper.SetNotServingStatus()
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

			// Create a unique client helper for this request to avoid race conditions
			clientHelper := NewConnectGoMockHelper(ctx.MockCtrl, ctx.Protocol)
			clientHelper.SetServingStatus()

			// Make request
			reqCtx, cancel := context.WithTimeout(context.Background(), ctx.TestTimeout)
			defer cancel()

			req := connect.NewRequest(&healthv1.CheckRequest{})
			resp, err := clientHelper.GetMock().Check(reqCtx, req)

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

			// Create a unique client helper for this request
			clientHelper := NewGRPCMockHelper(ctx.MockCtrl)
			clientHelper.SetServingStatus()

			// Make request
			reqCtx, cancel := context.WithTimeout(context.Background(), ctx.TestTimeout)
			defer cancel()

			req := &healthv1.CheckRequest{}
			resp, err := clientHelper.GetMock().Check(reqCtx, req)

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
