package e2e

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/seventeenthearth/sudal/gen/go/health/v1/healthv1connect"
	"github.com/seventeenthearth/sudal/test/e2e/steps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"

	healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
)

// TestConnectGoProtocols tests Connect-Go with different protocols
func TestConnectGoProtocols(t *testing.T) {
	// BDD Scenarios for different Connect-Go protocols
	scenarios := []steps.BDDScenario{
		{
			Name:        "Health check using Connect-Go gRPC-Web protocol",
			Description: "Should return SERVING status when making a health check request using gRPC-Web protocol",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
				steps.GivenConnectGoClientWithProtocol(ctx, ServerURL, "grpc-web")
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeConnectGoHealthCheckRequest(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenConnectGoResponseShouldIndicateServingStatus(ctx)
				steps.ThenConnectGoResponseShouldNotBeEmpty(ctx)
			},
		},
		{
			Name:        "Connect-Go HTTP/JSON protocol should be rejected for gRPC-only endpoints",
			Description: "Should return error when making HTTP/JSON request to gRPC-only endpoint",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
				steps.GivenConnectGoClientWithProtocol(ctx, ServerURL, "http")
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeConnectGoHealthCheckRequest(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenConnectGoRequestShouldFail(ctx)
			},
		},
		{
			Name:        "Multiple concurrent Connect-Go gRPC-Web requests",
			Description: "Should handle multiple concurrent gRPC-Web requests successfully",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
				steps.GivenConnectGoClientWithProtocol(ctx, ServerURL, "grpc-web")
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeConcurrentConnectGoHealthCheckRequests(ctx, 10)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenAllConnectGoRequestsShouldSucceed(ctx)
				steps.ThenAllConnectGoResponsesShouldIndicateServingStatus(ctx)
			},
		},
		{
			Name:        "Connect-Go protocol comparison with gRPC-only restriction",
			Description: "Should show that only gRPC-Web succeeds while HTTP/JSON fails for gRPC-only endpoints",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeHealthCheckRequestsWithDifferentProtocols(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenGRPCWebShouldSucceedAndHTTPShouldFail(ctx)
			},
		},
	}

	// Run all BDD scenarios
	steps.RunBDDScenarios(t, ServerURL, scenarios)
}

// TestConnectGoProtocolsTableDriven demonstrates table-driven BDD tests for different protocols
func TestConnectGoProtocolsTableDriven(t *testing.T) {
	// Table-driven test cases for different protocol scenarios
	type ProtocolTestCase struct {
		Name           string
		Protocol       string
		Timeout        time.Duration
		ExpectedStatus healthv1.ServingStatus
		ShouldSucceed  bool
		Description    string
	}

	testCases := []interface{}{
		ProtocolTestCase{
			Name:           "gRPC-Web protocol health request",
			Protocol:       "grpc-web",
			Timeout:        5 * time.Second,
			ExpectedStatus: healthv1.ServingStatus_SERVING_STATUS_SERVING,
			ShouldSucceed:  true,
			Description:    "gRPC-Web protocol health check should succeed",
		},
		ProtocolTestCase{
			Name:           "HTTP/JSON protocol health request should fail",
			Protocol:       "http",
			Timeout:        5 * time.Second,
			ExpectedStatus: healthv1.ServingStatus_SERVING_STATUS_SERVING,
			ShouldSucceed:  false,
			Description:    "HTTP/JSON protocol health check should fail for gRPC-only endpoint",
		},
		ProtocolTestCase{
			Name:           "gRPC-Web with short timeout",
			Protocol:       "grpc-web",
			Timeout:        100 * time.Millisecond,
			ExpectedStatus: healthv1.ServingStatus_SERVING_STATUS_SERVING,
			ShouldSucceed:  true,
			Description:    "gRPC-Web with short timeout should still succeed",
		},
	}

	tableDrivenTest := steps.TableDrivenBDDTest{
		Name: "Connect-Go protocol scenarios",
		Given: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(ProtocolTestCase)
			steps.GivenServerIsRunning(ctx)
			steps.GivenConnectGoClientWithProtocolAndTimeout(ctx, ServerURL, testCase.Protocol, testCase.Timeout)
		},
		When: func(ctx *steps.TestContext, testData interface{}) {
			steps.WhenIMakeConnectGoHealthCheckRequest(ctx)
		},
		Then: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(ProtocolTestCase)
			if testCase.ShouldSucceed {
				steps.ThenConnectGoResponseShouldIndicateServingStatus(ctx)
				steps.ThenConnectGoResponseShouldNotBeEmpty(ctx)
			} else {
				steps.ThenConnectGoRequestShouldFail(ctx)
			}
		},
	}

	steps.RunTableDrivenBDDTest(t, ServerURL, tableDrivenTest, testCases)
}

// TestConnectGoProtocolsConcurrency tests concurrent request scenarios with different protocols
func TestConnectGoProtocolsConcurrency(t *testing.T) {
	// Table-driven test for different concurrency levels and protocols
	type ConcurrencyTestCase struct {
		Name          string
		Protocol      string
		NumRequests   int
		ShouldSucceed bool
		Description   string
	}

	concurrencyTestCases := []interface{}{
		ConcurrencyTestCase{
			Name:          "Low gRPC-Web concurrency",
			Protocol:      "grpc-web",
			NumRequests:   5,
			ShouldSucceed: true,
			Description:   "Test with 5 concurrent gRPC-Web requests",
		},
		ConcurrencyTestCase{
			Name:          "Medium HTTP/JSON concurrency should fail",
			Protocol:      "http",
			NumRequests:   15,
			ShouldSucceed: false,
			Description:   "Test with 15 concurrent HTTP/JSON requests that should fail for gRPC-only endpoint",
		},
		ConcurrencyTestCase{
			Name:          "High gRPC-Web concurrency",
			Protocol:      "grpc-web",
			NumRequests:   30,
			ShouldSucceed: true,
			Description:   "Test with 30 concurrent gRPC-Web requests",
		},
	}

	concurrencyTest := steps.TableDrivenBDDTest{
		Name: "Connect-Go concurrency scenarios",
		Given: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(ConcurrencyTestCase)
			steps.GivenServerIsRunning(ctx)
			steps.GivenConnectGoClientWithProtocol(ctx, ServerURL, testCase.Protocol)
		},
		When: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(ConcurrencyTestCase)
			steps.WhenIMakeConcurrentConnectGoHealthCheckRequests(ctx, testCase.NumRequests)
		},
		Then: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(ConcurrencyTestCase)
			if testCase.ShouldSucceed {
				steps.ThenAllConnectGoRequestsShouldSucceed(ctx)
				steps.ThenAllConnectGoResponsesShouldIndicateServingStatus(ctx)
			} else {
				steps.ThenAllConnectGoRequestsShouldFail(ctx)
			}
		},
	}

	steps.RunTableDrivenBDDTest(t, ServerURL, concurrencyTest, concurrencyTestCases)
}

// TestConnectGoDirectProtocolComparison demonstrates direct protocol comparison
func TestConnectGoDirectProtocolComparison(t *testing.T) {
	protocols := []struct {
		name          string
		option        connect.ClientOption
		useH2         bool
		shouldSucceed bool
	}{
		{"gRPC", connect.WithGRPC(), true, true},         // Pure gRPC with HTTP/2
		{"gRPC-Web", connect.WithGRPCWeb(), false, true}, // gRPC-Web with HTTP/1.1
		{"HTTP/JSON", nil, false, false},                 // Default protocol with HTTP/1.1 - should fail for gRPC-only endpoint
	}

	for _, protocol := range protocols {
		t.Run("Direct "+protocol.name+" health check", func(t *testing.T) {
			// Given: Connect-Go client with specific protocol
			var client healthv1connect.HealthServiceClient
			var httpClient *http.Client

			if protocol.useH2 {
				// Use HTTP/2 client for pure gRPC
				httpClient = &http.Client{
					Transport: &http2.Transport{
						AllowHTTP: true,
						DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
							return net.Dial(network, addr)
						},
					},
				}
			} else {
				httpClient = http.DefaultClient
			}

			if protocol.option != nil {
				client = healthv1connect.NewHealthServiceClient(
					httpClient,
					ServerURL,
					protocol.option,
				)
			} else {
				client = healthv1connect.NewHealthServiceClient(
					httpClient,
					ServerURL,
				)
			}

			// When: Making health check request
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			req := connect.NewRequest(&healthv1.CheckRequest{})
			resp, err := client.Check(ctx, req)

			// Then: Check based on expected behavior
			if protocol.shouldSucceed {
				// Should succeed for gRPC and gRPC-Web
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)
				assert.Equal(t, healthv1.ServingStatus_SERVING_STATUS_SERVING, resp.Msg.Status)
				t.Logf("Successfully tested %s protocol", protocol.name)
			} else {
				// Should fail for HTTP/JSON on gRPC-only endpoint
				require.Error(t, err)
				t.Logf("Expected failure for %s protocol: %v", protocol.name, err)
			}
		})
	}
}
