package steps

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/cucumber/godog"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
	"github.com/seventeenthearth/sudal/gen/go/health/v1/healthv1connect"
)

// HealthCtx holds the context for health-related test scenarios
type HealthCtx struct {
	baseURL      string
	grpcEndpoint string

	// HTTP-related fields
	httpClient   *http.Client
	lastResponse *http.Response
	lastBody     []byte
	lastError    error

	// gRPC-related fields
	grpcConn     *grpc.ClientConn
	grpcClient   healthv1.HealthServiceClient
	grpcResponse *healthv1.CheckResponse
	grpcMetadata metadata.MD
	grpcTimeout  time.Duration

	// Concurrent test results
	concurrentResults     []ConcurrentResult
	grpcConcurrentResults []GRPCResult

	// Shared response for cross-context communication
	sharedResponse *SharedHTTPResponse

	// Connect-Go related fields
	connectClient            healthv1connect.HealthServiceClient
	connectResponse          *connect.Response[healthv1.CheckResponse]
	connectError             error
	connectConcurrentResults []ConnectResult
}

// ConcurrentResult holds the result of a concurrent HTTP request
type ConcurrentResult struct {
	Response *http.Response
	Body     []byte
	Error    error
}

// GRPCResult holds the result of a concurrent gRPC request
type GRPCResult struct {
	Response *healthv1.CheckResponse
	Error    error
	Metadata metadata.MD
}

// ConnectResult holds the result of a concurrent Connect-Go request
type ConnectResult struct {
	Response *connect.Response[healthv1.CheckResponse]
	Error    error
}

// NewHealthCtx creates a new health context
func NewHealthCtx() *HealthCtx {
	return &HealthCtx{
		baseURL:      getEnvOrDefault("BASE_URL", "http://localhost:8080"),
		grpcEndpoint: getEnvOrDefault("GRPC_ADDR", "localhost:8080"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		grpcTimeout: 5 * time.Second,
	}
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Given Steps

func (h *HealthCtx) theServerIsRunning() error {
	// Check if server is accessible via ping endpoint
	// Try multiple endpoints to verify server is running
	endpoints := []string{"/api/ping", "/api/healthz"}

	for _, endpoint := range endpoints {
		resp, err := h.httpClient.Get(fmt.Sprintf("%s%s", h.baseURL, endpoint))
		if err != nil {
			continue // Try next endpoint
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil // Server is running
		}
	}

	return fmt.Errorf("server is not running - none of the health endpoints are accessible")
}

func (h *HealthCtx) theGRPCClientIsConnected() error {
	conn, err := grpc.NewClient(h.grpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	h.grpcConn = conn
	h.grpcClient = healthv1.NewHealthServiceClient(conn)
	return nil
}

func (h *HealthCtx) theGRPCClientIsConnectedWithTimeout(timeoutStr string) error {
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return fmt.Errorf("invalid timeout format: %w", err)
	}

	h.grpcTimeout = timeout
	return h.theGRPCClientIsConnected()
}

// When Steps

func (h *HealthCtx) iCallTheHealthEndpoint(protocol string) error {
	switch strings.ToUpper(protocol) {
	case "REST":
		return h.makeRESTHealthRequest()
	case "GRPC":
		return h.makeGRPCHealthRequest()
	default:
		return fmt.Errorf("unknown protocol: %s", protocol)
	}
}

func (h *HealthCtx) iMakeAGETRequestTo(endpoint string) error {
	url := fmt.Sprintf("%s%s", h.baseURL, endpoint)
	resp, err := h.httpClient.Get(url)

	h.lastResponse = resp
	h.lastError = err

	if resp != nil {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			h.lastError = readErr
		} else {
			h.lastBody = body
		}
		resp.Body.Close()
	}

	return h.lastError
}

func (h *HealthCtx) iMakeAPOSTRequestToWithContentTypeAndBody(endpoint, contentType, body string) error {
	url := fmt.Sprintf("%s%s", h.baseURL, endpoint)

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		h.lastError = err
		return err
	}

	req.Header.Set("Content-Type", contentType)

	resp, err := h.httpClient.Do(req)
	h.lastResponse = resp
	h.lastError = err

	if resp != nil {
		respBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			h.lastError = readErr
		} else {
			h.lastBody = respBody
		}
		resp.Body.Close()
	}

	return h.lastError
}

func (h *HealthCtx) iMakeAConnectGoHealthRequest() error {
	url := fmt.Sprintf("%s/health.v1.HealthService/Check", h.baseURL)

	req, err := http.NewRequest("POST", url, strings.NewReader("{}"))
	if err != nil {
		h.lastError = err
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	h.lastResponse = resp
	h.lastError = err

	if resp != nil {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			h.lastError = readErr
		} else {
			h.lastBody = body
		}
		resp.Body.Close()
	}

	return h.lastError
}

func (h *HealthCtx) iMakeAGRPCHealthCheckRequest() error {
	if h.grpcClient == nil {
		return fmt.Errorf("gRPC client is not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.grpcTimeout)
	defer cancel()

	resp, err := h.grpcClient.Check(ctx, &healthv1.CheckRequest{})
	h.grpcResponse = resp
	h.lastError = err

	return err
}

func (h *HealthCtx) iMakeAGRPCHealthCheckRequestWithMetadata() error {
	if h.grpcClient == nil {
		return fmt.Errorf("gRPC client is not connected")
	}

	md := metadata.New(map[string]string{
		"test-header": "test-value",
	})

	ctx := metadata.NewOutgoingContext(context.Background(), md)
	ctx, cancel := context.WithTimeout(ctx, h.grpcTimeout)
	defer cancel()

	var header metadata.MD
	resp, err := h.grpcClient.Check(ctx, &healthv1.CheckRequest{}, grpc.Header(&header))
	h.grpcResponse = resp
	h.grpcMetadata = header
	h.lastError = err

	return err
}

func (h *HealthCtx) iMakeAGRPCHealthCheckRequestConditional(withMetadata string) error {
	if strings.TrimSpace(withMetadata) == "" {
		// No metadata
		return h.iMakeAGRPCHealthCheckRequest()
	} else {
		// With metadata
		return h.iMakeAGRPCHealthCheckRequestWithMetadata()
	}
}

func (h *HealthCtx) iMakeConcurrentHealthRequests(numRequestsStr string, protocol string) error {
	var numRequests int
	if _, err := fmt.Sscanf(numRequestsStr, "%d", &numRequests); err != nil {
		return fmt.Errorf("invalid number of requests: %w", err)
	}

	switch strings.ToUpper(protocol) {
	case "REST":
		return h.makeConcurrentRESTRequests(numRequests)
	case "GRPC":
		return h.makeConcurrentGRPCRequests(numRequests)
	default:
		return fmt.Errorf("unknown protocol: %s", protocol)
	}
}

// Helper methods for making requests

func (h *HealthCtx) makeRESTHealthRequest() error {
	return h.iMakeAGETRequestTo("/api/healthz")
}

func (h *HealthCtx) makeGRPCHealthRequest() error {
	return h.iMakeAGRPCHealthCheckRequest()
}

func (h *HealthCtx) makeConcurrentRESTRequests(numRequests int) error {
	var wg sync.WaitGroup
	results := make([]ConcurrentResult, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get(fmt.Sprintf("%s/api/healthz", h.baseURL))

			result := ConcurrentResult{
				Response: resp,
				Error:    err,
			}

			if resp != nil {
				body, readErr := io.ReadAll(resp.Body)
				if readErr != nil {
					result.Error = readErr
				} else {
					result.Body = body
				}
				resp.Body.Close()
			}

			results[index] = result
		}(i)
	}

	wg.Wait()
	h.concurrentResults = results
	return nil
}

func (h *HealthCtx) makeConcurrentGRPCRequests(numRequests int) error {
	if h.grpcClient == nil {
		return fmt.Errorf("gRPC client is not connected")
	}

	var wg sync.WaitGroup
	results := make([]GRPCResult, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), h.grpcTimeout)
			defer cancel()

			var header metadata.MD
			resp, err := h.grpcClient.Check(ctx, &healthv1.CheckRequest{}, grpc.Header(&header))

			results[index] = GRPCResult{
				Response: resp,
				Error:    err,
				Metadata: header,
			}
		}(i)
	}

	wg.Wait()
	h.grpcConcurrentResults = results
	return nil
}

// Then Steps

func (h *HealthCtx) theStatusShouldBeHealthy() error {
	if h.lastError != nil {
		return fmt.Errorf("request failed: %w", h.lastError)
	}

	if h.lastResponse != nil {
		if h.lastResponse.StatusCode != http.StatusOK {
			return fmt.Errorf("expected status 200, got %d", h.lastResponse.StatusCode)
		}

		var jsonData map[string]interface{}
		if err := json.Unmarshal(h.lastBody, &jsonData); err != nil {
			return fmt.Errorf("response is not valid JSON: %w", err)
		}

		status, exists := jsonData["status"]
		if !exists {
			return fmt.Errorf("response does not contain 'status' field")
		}

		if status != "healthy" {
			return fmt.Errorf("expected status 'healthy', got %v", status)
		}
	}

	if h.grpcResponse != nil {
		if h.grpcResponse.Status != healthv1.ServingStatus_SERVING_STATUS_SERVING {
			return fmt.Errorf("expected gRPC status SERVING, got %v", h.grpcResponse.Status)
		}
	}

	return nil
}

func (h *HealthCtx) theHTTPStatusShouldBe(expectedStatusStr string) error {
	var expectedStatus int
	if _, err := fmt.Sscanf(expectedStatusStr, "%d", &expectedStatus); err != nil {
		return fmt.Errorf("invalid status code: %w", err)
	}

	// Check local response first, then shared response
	var response *http.Response
	if h.lastResponse != nil {
		response = h.lastResponse
	} else if h.sharedResponse != nil && h.sharedResponse.Response != nil {
		response = h.sharedResponse.Response
	}

	if response == nil {
		return fmt.Errorf("no HTTP response received")
	}

	if response.StatusCode != expectedStatus {
		return fmt.Errorf("expected status %d, got %d", expectedStatus, response.StatusCode)
	}

	return nil
}

func (h *HealthCtx) theResponseShouldContainStatus(expectedStatus string) error {
	if h.lastBody == nil {
		return fmt.Errorf("no response body received")
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal(h.lastBody, &jsonData); err != nil {
		return fmt.Errorf("response is not valid JSON: %w", err)
	}

	status, exists := jsonData["status"]
	if !exists {
		return fmt.Errorf("response does not contain 'status' field")
	}

	if status != expectedStatus {
		return fmt.Errorf("expected status '%s', got %v", expectedStatus, status)
	}

	return nil
}

func (h *HealthCtx) theResponseShouldContainField(fieldName, expectedValue string) error {
	if h.lastBody == nil {
		return fmt.Errorf("no response body received")
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal(h.lastBody, &jsonData); err != nil {
		return fmt.Errorf("response is not valid JSON: %w", err)
	}

	value, exists := jsonData[fieldName]
	if !exists {
		return fmt.Errorf("response does not contain '%s' field", fieldName)
	}

	if fmt.Sprintf("%v", value) != expectedValue {
		return fmt.Errorf("expected %s '%s', got %v", fieldName, expectedValue, value)
	}

	return nil
}

func (h *HealthCtx) theGRPCResponseShouldIndicateServingStatus() error {
	if h.lastError != nil {
		return fmt.Errorf("gRPC request failed: %w", h.lastError)
	}

	if h.grpcResponse == nil {
		return fmt.Errorf("no gRPC response received")
	}

	if h.grpcResponse.Status != healthv1.ServingStatus_SERVING_STATUS_SERVING {
		return fmt.Errorf("expected gRPC status SERVING, got %v", h.grpcResponse.Status)
	}

	return nil
}

func (h *HealthCtx) theGRPCResponseShouldNotBeEmpty() error {
	if h.grpcResponse == nil {
		return fmt.Errorf("gRPC response is empty")
	}

	return nil
}

func (h *HealthCtx) theGRPCResponseShouldContainMetadata() error {
	if h.grpcMetadata == nil || len(h.grpcMetadata) == 0 {
		return fmt.Errorf("gRPC response does not contain metadata")
	}

	return nil
}

func (h *HealthCtx) allConcurrentRequestsShouldSucceed(protocol string) error {
	switch strings.ToUpper(protocol) {
	case "REST":
		return h.validateConcurrentRESTResults()
	case "GRPC":
		return h.validateConcurrentGRPCResults()
	default:
		return fmt.Errorf("unknown protocol: %s", protocol)
	}
}

func (h *HealthCtx) validateConcurrentRESTResults() error {
	if len(h.concurrentResults) == 0 {
		return fmt.Errorf("no concurrent results found")
	}

	failedCount := 0
	for _, result := range h.concurrentResults {
		if result.Error != nil {
			failedCount++
			continue
		}
		if result.Response == nil {
			failedCount++
			continue
		}
		if result.Response.StatusCode != http.StatusOK {
			failedCount++
			continue
		}

		var jsonData map[string]interface{}
		if err := json.Unmarshal(result.Body, &jsonData); err != nil {
			failedCount++
			continue
		}

		status, exists := jsonData["status"]
		if !exists || status != "healthy" {
			failedCount++
		}
	}

	if failedCount > 0 {
		return fmt.Errorf("%d out of %d concurrent requests failed", failedCount, len(h.concurrentResults))
	}

	return nil
}

func (h *HealthCtx) validateConcurrentGRPCResults() error {
	if len(h.grpcConcurrentResults) == 0 {
		return fmt.Errorf("no concurrent gRPC results found")
	}

	failedCount := 0
	for _, result := range h.grpcConcurrentResults {
		if result.Error != nil {
			failedCount++
			continue
		}
		if result.Response == nil {
			failedCount++
			continue
		}
		if result.Response.Status != healthv1.ServingStatus_SERVING_STATUS_SERVING {
			failedCount++
		}
	}

	if failedCount > 0 {
		return fmt.Errorf("%d out of %d concurrent gRPC requests failed", failedCount, len(h.grpcConcurrentResults))
	}

	return nil
}

// Register registers all health-related step definitions
func (h *HealthCtx) Register(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^the server is running$`, h.theServerIsRunning)
	sc.Step(`^the gRPC client is connected$`, h.theGRPCClientIsConnected)
	sc.Step(`^the gRPC client is connected with timeout "([^"]*)"$`, h.theGRPCClientIsConnectedWithTimeout)

	// When steps
	sc.Step(`^I call the "([^"]*)" health endpoint$`, h.iCallTheHealthEndpoint)
	sc.Step(`^I make a GET request to "([^"]*)"$`, h.iMakeAGETRequestTo)
	sc.Step(`^I make a POST request to "([^"]*)" with content type "([^"]*)" and body "([^"]*)"$`, h.iMakeAPOSTRequestToWithContentTypeAndBody)
	sc.Step(`^I make a Connect-Go health request$`, h.iMakeAConnectGoHealthRequest)
	sc.Step(`^I make a gRPC health check request$`, h.iMakeAGRPCHealthCheckRequest)
	sc.Step(`^I make a gRPC health check request with metadata$`, h.iMakeAGRPCHealthCheckRequestWithMetadata)
	sc.Step(`^I make a gRPC health check request( with metadata)?$`, h.iMakeAGRPCHealthCheckRequestConditional)
	sc.Step(`^I make (\d+) concurrent "([^"]*)" health requests$`, h.iMakeConcurrentHealthRequests)

	// Connect-Go specific steps
	sc.Step(`^the Connect-Go client is configured with "([^"]*)" protocol$`, h.theConnectGoClientIsConfiguredWithProtocol)
	sc.Step(`^the Connect-Go client is configured with "([^"]*)" protocol and (\d+)ms timeout$`, h.theConnectGoClientIsConfiguredWithProtocolAndTimeout)
	sc.Step(`^I make a Connect-Go health check request$`, h.iMakeAConnectGoHealthCheckRequest)
	sc.Step(`^I make (\d+) concurrent Connect-Go health check requests$`, h.iMakeConcurrentConnectGoHealthCheckRequests)
	sc.Step(`^I make (\d+) concurrent GET requests to "([^"]*)"$`, h.iMakeConcurrentGETRequestsTo)

	// Then steps
	sc.Step(`^the status should be healthy$`, h.theStatusShouldBeHealthy)
	sc.Step(`^the HTTP status should be (\d+)$`, h.theHTTPStatusShouldBe)
	sc.Step(`^the response should contain status "([^"]*)"$`, h.theResponseShouldContainStatus)
	sc.Step(`^the response should contain field "([^"]*)" with value "([^"]*)"$`, h.theResponseShouldContainField)
	sc.Step(`^the gRPC response should indicate serving status$`, h.theGRPCResponseShouldIndicateServingStatus)
	sc.Step(`^the gRPC response should not be empty$`, h.theGRPCResponseShouldNotBeEmpty)
	sc.Step(`^the gRPC response should contain metadata$`, h.theGRPCResponseShouldContainMetadata)
	sc.Step(`^all concurrent "([^"]*)" requests should succeed$`, h.allConcurrentRequestsShouldSucceed)

	// Connect-Go specific Then steps
	sc.Step(`^the Connect-Go response should indicate serving status$`, h.theConnectGoResponseShouldIndicateServingStatus)
	sc.Step(`^the Connect-Go response should not be empty$`, h.theConnectGoResponseShouldNotBeEmpty)
	sc.Step(`^the Connect-Go request should fail$`, h.theConnectGoRequestShouldFail)
	sc.Step(`^the Connect-Go request should succeed$`, h.theConnectGoRequestShouldSucceed)
	sc.Step(`^all Connect-Go requests should succeed$`, h.allConnectGoRequestsShouldSucceed)
	sc.Step(`^all Connect-Go requests should fail$`, h.allConnectGoRequestsShouldFail)
	sc.Step(`^all Connect-Go responses should indicate serving status$`, h.allConnectGoResponsesShouldIndicateServingStatus)

	// Monitoring specific Then steps
	sc.Step(`^the JSON response should contain status "([^"]*)"$`, h.theJSONResponseShouldContainStatus)
	sc.Step(`^the response should be lightweight for monitoring$`, h.theResponseShouldBeLightweightForMonitoring)
	sc.Step(`^the content type should be "([^"]*)"$`, h.theContentTypeShouldBe)
	sc.Step(`^all requests should succeed$`, h.allRequestsShouldSucceed)
	sc.Step(`^all responses should contain status "([^"]*)"$`, h.allResponsesShouldContainStatus)
}

// Connect-Go step implementations

func (h *HealthCtx) theConnectGoClientIsConfiguredWithProtocol(protocol string) error {
	var httpClient *http.Client

	switch protocol {
	case "grpc":
		// Use HTTP/2 client for pure gRPC
		httpClient = &http.Client{
			Transport: &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					return net.Dial(network, addr)
				},
			},
		}
		h.connectClient = healthv1connect.NewHealthServiceClient(
			httpClient,
			h.baseURL,
			connect.WithGRPC(),
		)
	case "grpc-web":
		httpClient = http.DefaultClient
		h.connectClient = healthv1connect.NewHealthServiceClient(
			httpClient,
			h.baseURL,
			connect.WithGRPCWeb(),
		)
	case "http":
		httpClient = http.DefaultClient
		// No option means default HTTP/JSON protocol
		h.connectClient = healthv1connect.NewHealthServiceClient(
			httpClient,
			h.baseURL,
		)
	default:
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}

	return nil
}

func (h *HealthCtx) theConnectGoClientIsConfiguredWithProtocolAndTimeout(protocol string, timeoutMs int) error {
	err := h.theConnectGoClientIsConfiguredWithProtocol(protocol)
	if err != nil {
		return err
	}

	// Timeout is handled at request level in Connect-Go
	return nil
}

func (h *HealthCtx) iMakeAConnectGoHealthCheckRequest() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := connect.NewRequest(&healthv1.CheckRequest{})
	resp, err := h.connectClient.Check(ctx, req)

	h.connectResponse = resp
	h.connectError = err

	return nil
}

func (h *HealthCtx) iMakeConcurrentConnectGoHealthCheckRequests(numRequests int) error {
	var wg sync.WaitGroup
	results := make([]ConnectResult, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			req := connect.NewRequest(&healthv1.CheckRequest{})
			resp, err := h.connectClient.Check(ctx, req)

			results[index] = ConnectResult{
				Response: resp,
				Error:    err,
			}
		}(i)
	}

	wg.Wait()
	h.connectConcurrentResults = results

	return nil
}

func (h *HealthCtx) iMakeConcurrentGETRequestsTo(numRequests int, endpoint string) error {
	var wg sync.WaitGroup
	results := make([]ConcurrentResult, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			url := fmt.Sprintf("%s%s", h.baseURL, endpoint)
			resp, err := h.httpClient.Get(url)

			var body []byte
			if resp != nil {
				defer resp.Body.Close()
				body, _ = io.ReadAll(resp.Body)
			}

			results[index] = ConcurrentResult{
				Response: resp,
				Body:     body,
				Error:    err,
			}
		}(i)
	}

	wg.Wait()
	h.concurrentResults = results

	return nil
}

// Connect-Go Then step implementations

func (h *HealthCtx) theConnectGoResponseShouldIndicateServingStatus() error {
	if h.connectError != nil {
		return fmt.Errorf("Connect-Go request failed: %w", h.connectError)
	}

	if h.connectResponse == nil {
		return fmt.Errorf("no Connect-Go response received")
	}

	if h.connectResponse.Msg.Status != healthv1.ServingStatus_SERVING_STATUS_SERVING {
		return fmt.Errorf("expected serving status, got %v", h.connectResponse.Msg.Status)
	}

	return nil
}

func (h *HealthCtx) theConnectGoResponseShouldNotBeEmpty() error {
	if h.connectResponse == nil {
		return fmt.Errorf("Connect-Go response is empty")
	}

	if h.connectResponse.Msg == nil {
		return fmt.Errorf("Connect-Go response message is empty")
	}

	return nil
}

func (h *HealthCtx) theConnectGoRequestShouldFail() error {
	if h.connectError == nil {
		return fmt.Errorf("expected Connect-Go request to fail, but it succeeded")
	}

	return nil
}

func (h *HealthCtx) theConnectGoRequestShouldSucceed() error {
	if h.connectError != nil {
		return fmt.Errorf("expected Connect-Go request to succeed, but it failed: %w", h.connectError)
	}

	return nil
}

func (h *HealthCtx) allConnectGoRequestsShouldSucceed() error {
	for i, result := range h.connectConcurrentResults {
		if result.Error != nil {
			return fmt.Errorf("Connect-Go request %d failed: %w", i, result.Error)
		}
	}
	return nil
}

func (h *HealthCtx) allConnectGoRequestsShouldFail() error {
	for i, result := range h.connectConcurrentResults {
		if result.Error == nil {
			return fmt.Errorf("Connect-Go request %d should have failed but succeeded", i)
		}
	}
	return nil
}

func (h *HealthCtx) allConnectGoResponsesShouldIndicateServingStatus() error {
	for i, result := range h.connectConcurrentResults {
		if result.Error != nil {
			return fmt.Errorf("Connect-Go request %d failed: %w", i, result.Error)
		}
		if result.Response == nil || result.Response.Msg == nil {
			return fmt.Errorf("Connect-Go request %d has empty response", i)
		}
		if result.Response.Msg.Status != healthv1.ServingStatus_SERVING_STATUS_SERVING {
			return fmt.Errorf("Connect-Go request %d has wrong status: %v", i, result.Response.Msg.Status)
		}
	}
	return nil
}

// Monitoring step implementations

func (h *HealthCtx) theJSONResponseShouldContainStatus(expectedStatus string) error {
	// Check local response first, then shared response
	var response *http.Response
	var body []byte
	if h.lastResponse != nil {
		response = h.lastResponse
		body = h.lastBody
	} else if h.sharedResponse != nil && h.sharedResponse.Response != nil {
		response = h.sharedResponse.Response
		body = h.sharedResponse.Body
	}

	if response == nil {
		return fmt.Errorf("no HTTP response received")
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("expected status 200, got %d", response.StatusCode)
	}

	if len(body) == 0 {
		return fmt.Errorf("no response body received")
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		return fmt.Errorf("response is not valid JSON: %w", err)
	}

	status, exists := jsonData["status"]
	if !exists {
		return fmt.Errorf("response does not contain 'status' field")
	}

	statusStr, ok := status.(string)
	if !ok {
		return fmt.Errorf("status field is not a string")
	}

	if statusStr != expectedStatus {
		return fmt.Errorf("expected status '%s', got '%s'", expectedStatus, statusStr)
	}

	return nil
}

func (h *HealthCtx) theResponseShouldBeLightweightForMonitoring() error {
	// Check local response first, then shared response
	var response *http.Response
	var body []byte
	if h.lastResponse != nil {
		response = h.lastResponse
		body = h.lastBody
	} else if h.sharedResponse != nil && h.sharedResponse.Response != nil {
		response = h.sharedResponse.Response
		body = h.sharedResponse.Body
	}

	if response == nil {
		return fmt.Errorf("no HTTP response received")
	}

	if len(body) == 0 {
		return fmt.Errorf("no response body received")
	}

	// Check that response is small (lightweight)
	if len(body) > 1024 { // 1KB threshold for lightweight response
		return fmt.Errorf("response is too large for monitoring: %d bytes", len(body))
	}

	// Check that it's valid JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		return fmt.Errorf("response is not valid JSON: %w", err)
	}

	return nil
}

func (h *HealthCtx) theContentTypeShouldBe(expectedContentType string) error {
	// Check local response first, then shared response
	var response *http.Response
	if h.lastResponse != nil {
		response = h.lastResponse
	} else if h.sharedResponse != nil && h.sharedResponse.Response != nil {
		response = h.sharedResponse.Response
	}

	if response == nil {
		return fmt.Errorf("no HTTP response received")
	}

	contentType := response.Header.Get("Content-Type")
	if !strings.Contains(contentType, expectedContentType) {
		return fmt.Errorf("expected content type to contain '%s', got '%s'", expectedContentType, contentType)
	}

	return nil
}

func (h *HealthCtx) allRequestsShouldSucceed() error {
	for i, result := range h.concurrentResults {
		if result.Error != nil {
			return fmt.Errorf("request %d failed: %w", i, result.Error)
		}
		if result.Response == nil {
			return fmt.Errorf("request %d has no response", i)
		}
		if result.Response.StatusCode != 200 {
			return fmt.Errorf("request %d returned status %d", i, result.Response.StatusCode)
		}
	}
	return nil
}

func (h *HealthCtx) allResponsesShouldContainStatus(expectedStatus string) error {
	for i, result := range h.concurrentResults {
		if result.Error != nil {
			return fmt.Errorf("request %d failed: %w", i, result.Error)
		}
		if result.Response == nil {
			return fmt.Errorf("request %d has no response", i)
		}
		if result.Response.StatusCode != 200 {
			return fmt.Errorf("request %d returned status %d", i, result.Response.StatusCode)
		}

		var jsonData map[string]interface{}
		if err := json.Unmarshal(result.Body, &jsonData); err != nil {
			return fmt.Errorf("request %d response is not valid JSON: %w", i, err)
		}

		status, exists := jsonData["status"]
		if !exists {
			return fmt.Errorf("request %d response does not contain 'status' field", i)
		}

		statusStr, ok := status.(string)
		if !ok {
			return fmt.Errorf("request %d status field is not a string", i)
		}

		if statusStr != expectedStatus {
			return fmt.Errorf("request %d expected status '%s', got '%s'", i, expectedStatus, statusStr)
		}
	}
	return nil
}

// Cleanup cleans up resources
func (h *HealthCtx) Cleanup() {
	if h.grpcConn != nil {
		h.grpcConn.Close()
	}
}
