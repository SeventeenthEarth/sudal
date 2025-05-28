package mocks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"connectrpc.com/connect"
	healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
)

// MockConnectGoClient is a mock implementation of the Connect-Go health service client
type MockConnectGoClient struct {
	CheckFunc func(ctx context.Context, req *connect.Request[healthv1.CheckRequest]) (*connect.Response[healthv1.CheckResponse], error)

	// Configuration for mock behavior
	ShouldFailCheck bool
	CustomError     error
	CustomResponse  *healthv1.CheckResponse
	Protocol        string
	Metadata        map[string]string
}

// NewMockConnectGoClient creates a new mock Connect-Go client
func NewMockConnectGoClient(protocol string) *MockConnectGoClient {
	return &MockConnectGoClient{
		Protocol: protocol,
		CustomResponse: &healthv1.CheckResponse{
			Status: healthv1.ServingStatus_SERVING_STATUS_SERVING,
		},
		Metadata: make(map[string]string),
	}
}

// NewMockConnectGoClientWithError creates a mock that returns errors
func NewMockConnectGoClientWithError(protocol string, err error) *MockConnectGoClient {
	return &MockConnectGoClient{
		Protocol:        protocol,
		ShouldFailCheck: true,
		CustomError:     err,
		Metadata:        make(map[string]string),
	}
}

// Check performs a mock health check request
func (m *MockConnectGoClient) Check(ctx context.Context, req *connect.Request[healthv1.CheckRequest]) (*connect.Response[healthv1.CheckResponse], error) {
	if m.CheckFunc != nil {
		return m.CheckFunc(ctx, req)
	}

	if m.ShouldFailCheck {
		if m.CustomError != nil {
			return nil, m.CustomError
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("mock grpc check failed"))
	}

	// Create response with metadata
	resp := connect.NewResponse(m.CustomResponse)

	// Add mock metadata based on protocol
	switch m.Protocol {
	case "grpc-web":
		resp.Header().Set("Content-Type", "application/grpc-web+proto")
		resp.Header().Set("Grpc-Accept-Encoding", "gzip")
	case "http":
		resp.Header().Set("Content-Type", "application/json")
	default:
		resp.Header().Set("Content-Type", "application/grpc+proto")
	}

	// Add custom metadata
	for key, value := range m.Metadata {
		resp.Header().Set(key, value)
	}

	return resp, nil
}

// SetServingStatus configures the mock to return SERVING status
func (m *MockConnectGoClient) SetServingStatus() {
	m.ShouldFailCheck = false
	m.CustomResponse = &healthv1.CheckResponse{
		Status: healthv1.ServingStatus_SERVING_STATUS_SERVING,
	}
}

// SetNotServingStatus configures the mock to return NOT_SERVING status
func (m *MockConnectGoClient) SetNotServingStatus() {
	m.ShouldFailCheck = false
	m.CustomResponse = &healthv1.CheckResponse{
		Status: healthv1.ServingStatus_SERVING_STATUS_NOT_SERVING,
	}
}

// SetUnknownStatus configures the mock to return UNKNOWN status
func (m *MockConnectGoClient) SetUnknownStatus() {
	m.ShouldFailCheck = false
	m.CustomResponse = &healthv1.CheckResponse{
		Status: healthv1.ServingStatus_SERVING_STATUS_UNKNOWN_UNSPECIFIED,
	}
}

// SetError configures the mock to return an error
func (m *MockConnectGoClient) SetError(err error) {
	m.ShouldFailCheck = true
	m.CustomError = err
}

// AddMetadata adds custom metadata to responses
func (m *MockConnectGoClient) AddMetadata(key, value string) {
	m.Metadata[key] = value
}

// GetProtocol returns the protocol this mock client is configured for
func (m *MockConnectGoClient) GetProtocol() string {
	return m.Protocol
}

// MockGRPCClient is a mock implementation for native gRPC client
type MockGRPCClient struct {
	CheckFunc func(ctx context.Context, req *healthv1.CheckRequest) (*healthv1.CheckResponse, error)

	// Configuration for mock behavior
	ShouldFailCheck bool
	CustomError     error
	CustomResponse  *healthv1.CheckResponse
	Metadata        map[string]string
}

// NewMockGRPCClient creates a new mock native gRPC client
func NewMockGRPCClient() *MockGRPCClient {
	return &MockGRPCClient{
		CustomResponse: &healthv1.CheckResponse{
			Status: healthv1.ServingStatus_SERVING_STATUS_SERVING,
		},
		Metadata: make(map[string]string),
	}
}

// NewMockGRPCClientWithError creates a mock that returns errors
func NewMockGRPCClientWithError(err error) *MockGRPCClient {
	return &MockGRPCClient{
		ShouldFailCheck: true,
		CustomError:     err,
		Metadata:        make(map[string]string),
	}
}

// Check performs a mock health check request
func (m *MockGRPCClient) Check(ctx context.Context, req *healthv1.CheckRequest) (*healthv1.CheckResponse, error) {
	if m.CheckFunc != nil {
		return m.CheckFunc(ctx, req)
	}

	if m.ShouldFailCheck {
		if m.CustomError != nil {
			return nil, m.CustomError
		}
		return nil, fmt.Errorf("mock grpc check failed")
	}

	return m.CustomResponse, nil
}

// SetServingStatus configures the mock to return SERVING status
func (m *MockGRPCClient) SetServingStatus() {
	m.ShouldFailCheck = false
	m.CustomResponse = &healthv1.CheckResponse{
		Status: healthv1.ServingStatus_SERVING_STATUS_SERVING,
	}
}

// SetNotServingStatus configures the mock to return NOT_SERVING status
func (m *MockGRPCClient) SetNotServingStatus() {
	m.ShouldFailCheck = false
	m.CustomResponse = &healthv1.CheckResponse{
		Status: healthv1.ServingStatus_SERVING_STATUS_NOT_SERVING,
	}
}

// SetError configures the mock to return an error
func (m *MockGRPCClient) SetError(err error) {
	m.ShouldFailCheck = true
	m.CustomError = err
}

// ConcurrentTestResult represents the result of a concurrent test operation
type ConcurrentTestResult struct {
	Success  bool
	Error    error
	Duration time.Duration
	Protocol string
	Metadata map[string]string
}

// MockConcurrentTester provides utilities for testing concurrent operations
type MockConcurrentTester struct {
	Results []ConcurrentTestResult
	mutex   sync.Mutex
}

// NewMockConcurrentTester creates a new concurrent tester
func NewMockConcurrentTester() *MockConcurrentTester {
	return &MockConcurrentTester{
		Results: make([]ConcurrentTestResult, 0),
	}
}

// AddResult adds a test result (thread-safe)
func (m *MockConcurrentTester) AddResult(result ConcurrentTestResult) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Results = append(m.Results, result)
}

// GetResults returns all test results (thread-safe)
func (m *MockConcurrentTester) GetResults() []ConcurrentTestResult {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	resultsCopy := make([]ConcurrentTestResult, len(m.Results))
	copy(resultsCopy, m.Results)
	return resultsCopy
}

// GetSuccessCount returns the number of successful operations
func (m *MockConcurrentTester) GetSuccessCount() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	count := 0
	for _, result := range m.Results {
		if result.Success {
			count++
		}
	}
	return count
}

// GetErrorCount returns the number of failed operations
func (m *MockConcurrentTester) GetErrorCount() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	count := 0
	for _, result := range m.Results {
		if !result.Success {
			count++
		}
	}
	return count
}

// Clear clears all results
func (m *MockConcurrentTester) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Results = m.Results[:0]
}

// ProtocolTestScenario represents a test scenario for different protocols
type ProtocolTestScenario struct {
	Name           string
	Protocol       string
	ExpectedStatus healthv1.ServingStatus
	ShouldSucceed  bool
	Metadata       map[string]string
	Timeout        time.Duration
}

// NewProtocolTestScenario creates a new protocol test scenario
func NewProtocolTestScenario(name, protocol string, expectedStatus healthv1.ServingStatus, shouldSucceed bool) ProtocolTestScenario {
	return ProtocolTestScenario{
		Name:           name,
		Protocol:       protocol,
		ExpectedStatus: expectedStatus,
		ShouldSucceed:  shouldSucceed,
		Metadata:       make(map[string]string),
		Timeout:        5 * time.Second,
	}
}

// AddMetadata adds metadata to the test scenario
func (p *ProtocolTestScenario) AddMetadata(key, value string) {
	p.Metadata[key] = value
}

// SetTimeout sets the timeout for the test scenario
func (p *ProtocolTestScenario) SetTimeout(timeout time.Duration) {
	p.Timeout = timeout
}
