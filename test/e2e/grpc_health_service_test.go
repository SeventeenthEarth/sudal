package e2e

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/seventeenthearth/sudal/test/e2e/steps"
	"golang.org/x/net/http2"

	healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
	"github.com/seventeenthearth/sudal/gen/go/health/v1/healthv1connect"
)

// TestGRPCHealthService tests the gRPC Health Service functionality using native gRPC client
func TestGRPCHealthService(t *testing.T) {
	// BDD Scenarios for native gRPC Health Service
	scenarios := []steps.BDDScenario{
		{
			Name:        "Health check using native gRPC client",
			Description: "Should return SERVING status when making a health check request using native gRPC client",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
				steps.GivenGRPCClientIsConnected(ctx, GRPCServerURL)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeGRPCHealthCheckRequest(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenGRPCResponseShouldIndicateServingStatus(ctx)
				steps.ThenGRPCResponseShouldNotBeEmpty(ctx)
			},
		},
		{
			Name:        "Multiple concurrent gRPC health requests",
			Description: "Should handle multiple concurrent gRPC requests successfully",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
				steps.GivenGRPCClientIsConnected(ctx, GRPCServerURL)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeConcurrentGRPCHealthCheckRequests(ctx, 10)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenAllGRPCRequestsShouldSucceed(ctx)
				steps.ThenAllGRPCResponsesShouldIndicateServingStatus(ctx)
			},
		},
		{
			Name:        "gRPC connection timeout handling",
			Description: "Should handle connection timeout gracefully",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
				steps.GivenGRPCClientWithTimeout(ctx, GRPCServerURL, 1*time.Second)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeGRPCHealthCheckRequest(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenGRPCResponseShouldIndicateServingStatus(ctx)
			},
		},
		{
			Name:        "gRPC metadata handling",
			Description: "Should handle gRPC metadata correctly",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
				steps.GivenGRPCClientIsConnected(ctx, GRPCServerURL)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeGRPCHealthCheckRequestWithMetadata(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenGRPCResponseShouldIndicateServingStatus(ctx)
				steps.ThenGRPCResponseShouldContainMetadata(ctx)
			},
		},
	}

	// Run all BDD scenarios
	steps.RunBDDScenarios(t, ServerURL, scenarios)
}

// TestGRPCHealthServiceTableDriven demonstrates table-driven BDD tests for gRPC
func TestGRPCHealthServiceTableDriven(t *testing.T) {
	// Table-driven test cases for different gRPC scenarios
	type GRPCTestCase struct {
		Name           string
		Timeout        time.Duration
		WithMetadata   bool
		ExpectedStatus healthv1.ServingStatus
		ShouldSucceed  bool
		Description    string
	}

	testCases := []interface{}{
		GRPCTestCase{
			Name:           "Valid gRPC health request",
			Timeout:        5 * time.Second,
			WithMetadata:   false,
			ExpectedStatus: healthv1.ServingStatus_SERVING_STATUS_SERVING,
			ShouldSucceed:  true,
			Description:    "Standard gRPC health check should succeed",
		},
		GRPCTestCase{
			Name:           "gRPC health request with metadata",
			Timeout:        5 * time.Second,
			WithMetadata:   true,
			ExpectedStatus: healthv1.ServingStatus_SERVING_STATUS_SERVING,
			ShouldSucceed:  true,
			Description:    "gRPC health check with metadata should succeed",
		},
		GRPCTestCase{
			Name:           "gRPC health request with short timeout",
			Timeout:        100 * time.Millisecond,
			WithMetadata:   false,
			ExpectedStatus: healthv1.ServingStatus_SERVING_STATUS_SERVING,
			ShouldSucceed:  true,
			Description:    "gRPC health check with short timeout should still succeed",
		},
	}

	tableDrivenTest := steps.TableDrivenBDDTest{
		Name: "gRPC request scenarios",
		Given: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(GRPCTestCase)
			steps.GivenServerIsRunning(ctx)
			steps.GivenGRPCClientWithTimeout(ctx, GRPCServerURL, testCase.Timeout)
		},
		When: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(GRPCTestCase)
			if testCase.WithMetadata {
				steps.WhenIMakeGRPCHealthCheckRequestWithMetadata(ctx)
			} else {
				steps.WhenIMakeGRPCHealthCheckRequest(ctx)
			}
		},
		Then: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(GRPCTestCase)
			if testCase.ShouldSucceed {
				steps.ThenGRPCResponseShouldIndicateServingStatus(ctx)
				steps.ThenGRPCResponseShouldNotBeEmpty(ctx)
			}
		},
	}

	steps.RunTableDrivenBDDTest(t, ServerURL, tableDrivenTest, testCases)
}

// TestGRPCHealthServiceConcurrency tests concurrent gRPC request scenarios
func TestGRPCHealthServiceConcurrency(t *testing.T) {
	// Table-driven test for different concurrency levels
	type ConcurrencyTestCase struct {
		Name        string
		NumRequests int
		Description string
	}

	concurrencyTestCases := []interface{}{
		ConcurrencyTestCase{
			Name:        "Low gRPC concurrency",
			NumRequests: 5,
			Description: "Test with 5 concurrent gRPC requests",
		},
		ConcurrencyTestCase{
			Name:        "Medium gRPC concurrency",
			NumRequests: 15,
			Description: "Test with 15 concurrent gRPC requests",
		},
		ConcurrencyTestCase{
			Name:        "High gRPC concurrency",
			NumRequests: 30,
			Description: "Test with 30 concurrent gRPC requests",
		},
	}

	concurrencyTest := steps.TableDrivenBDDTest{
		Name: "gRPC concurrency scenarios",
		Given: func(ctx *steps.TestContext, testData interface{}) {
			steps.GivenServerIsRunning(ctx)
			steps.GivenGRPCClientIsConnected(ctx, GRPCServerURL)
		},
		When: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(ConcurrencyTestCase)
			steps.WhenIMakeConcurrentGRPCHealthCheckRequests(ctx, testCase.NumRequests)
		},
		Then: func(ctx *steps.TestContext, testData interface{}) {
			steps.ThenAllGRPCRequestsShouldSucceed(ctx)
			steps.ThenAllGRPCResponsesShouldIndicateServingStatus(ctx)
		},
	}

	steps.RunTableDrivenBDDTest(t, ServerURL, concurrencyTest, concurrencyTestCases)
}

// TestGRPCHealthServiceDirectClient demonstrates Connect-Go client usage with different protocols
func TestGRPCHealthServiceDirectClient(t *testing.T) {
	// Test Connect-Go client with gRPC protocol
	t.Run("Connect-Go gRPC protocol health check", func(t *testing.T) {
		// Given: HTTP/2 client for gRPC protocol
		h2Client := &http.Client{
			Transport: &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					return net.Dial(network, addr)
				},
			},
		}

		client := healthv1connect.NewHealthServiceClient(
			h2Client,
			ServerURL,
			connect.WithGRPC(), // Use gRPC protocol
		)

		// When: Making health check request
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req := connect.NewRequest(&healthv1.CheckRequest{})
		resp, err := client.Check(ctx, req)

		// Then: Response should indicate serving status (BDD style)
		if err != nil {
			t.Errorf("Expected Connect-Go gRPC request to succeed, but got error: %v", err)
			return
		}
		if resp == nil {
			t.Errorf("Expected Connect-Go gRPC response to exist, but it was nil")
			return
		}
		if resp.Msg == nil {
			t.Errorf("Expected Connect-Go gRPC response message to exist, but it was nil")
			return
		}
		if resp.Msg.Status != healthv1.ServingStatus_SERVING_STATUS_SERVING {
			t.Errorf("Expected Connect-Go gRPC response status to be SERVING_STATUS_SERVING, but got %v", resp.Msg.Status)
		}
	})

	// Test Connect-Go client with gRPC-Web protocol
	t.Run("Connect-Go gRPC-Web protocol health check", func(t *testing.T) {
		// Given: Connect-Go client with gRPC-Web protocol
		client := healthv1connect.NewHealthServiceClient(
			http.DefaultClient,
			ServerURL,
			connect.WithGRPCWeb(), // Use gRPC-Web protocol
		)

		// When: Making health check request
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req := connect.NewRequest(&healthv1.CheckRequest{})
		resp, err := client.Check(ctx, req)

		// Then: Response should indicate serving status (BDD style)
		if err != nil {
			t.Errorf("Expected Connect-Go gRPC-Web request to succeed, but got error: %v", err)
			return
		}
		if resp == nil {
			t.Errorf("Expected Connect-Go gRPC-Web response to exist, but it was nil")
			return
		}
		if resp.Msg == nil {
			t.Errorf("Expected Connect-Go gRPC-Web response message to exist, but it was nil")
			return
		}
		if resp.Msg.Status != healthv1.ServingStatus_SERVING_STATUS_SERVING {
			t.Errorf("Expected Connect-Go gRPC-Web response status to be SERVING_STATUS_SERVING, but got %v", resp.Msg.Status)
		}
	})

	// Test Connect-Go client with default protocol (HTTP/JSON)
	// This should fail because gRPC-only endpoints reject HTTP/JSON protocol
	t.Run("Connect-Go default protocol health check", func(t *testing.T) {
		// Given: Connect-Go client with default protocol
		client := healthv1connect.NewHealthServiceClient(
			http.DefaultClient,
			ServerURL,
			// No protocol specified - uses default HTTP/JSON
		)

		// When: Making health check request
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req := connect.NewRequest(&healthv1.CheckRequest{})
		resp, err := client.Check(ctx, req)

		// Then: Request should fail because gRPC-only endpoints reject HTTP/JSON (BDD style)
		if err == nil {
			t.Errorf("Expected Connect-Go default protocol request to fail on gRPC-only endpoint, but it succeeded")
			return
		}

		// Verify that the error indicates the expected rejection (404 Not Found or unimplemented)
		errorStr := err.Error()
		if !strings.Contains(errorStr, "404") && !strings.Contains(errorStr, "unimplemented") {
			t.Errorf("Expected Connect-Go default protocol error to indicate 404 or unimplemented, but got: %v", err)
		}

		// Response should be nil when request fails
		if resp != nil {
			t.Errorf("Expected Connect-Go default protocol response to be nil when request fails, but got: %v", resp)
		}
	})
}
