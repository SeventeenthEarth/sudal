package steps

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"

	healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
	"github.com/seventeenthearth/sudal/gen/go/health/v1/healthv1connect"
)

// Connect-Go specific Given Steps

// GivenConnectGoClientWithProtocol establishes a Connect-Go client with specific protocol
func GivenConnectGoClientWithProtocol(ctx *TestContext, serverURL, protocol string) {
	var client healthv1connect.HealthServiceClient

	switch protocol {
	case "grpc":
		// Use HTTP/2 client for pure gRPC protocol
		h2Client := &http.Client{
			Transport: &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					return net.Dial(network, addr)
				},
			},
		}
		client = healthv1connect.NewHealthServiceClient(
			h2Client,
			serverURL,
			connect.WithGRPC(),
		)
	case "grpc-web":
		client = healthv1connect.NewHealthServiceClient(
			http.DefaultClient,
			serverURL,
			connect.WithGRPCWeb(),
		)
	case "http":
		fallthrough
	default:
		client = healthv1connect.NewHealthServiceClient(
			http.DefaultClient,
			serverURL,
		)
	}

	ctx.ConnectGoClient = client
	ctx.ConnectGoProtocol = protocol
}

// GivenConnectGoClientWithProtocolAndTimeout establishes a Connect-Go client with protocol and timeout
func GivenConnectGoClientWithProtocolAndTimeout(ctx *TestContext, serverURL, protocol string, timeout time.Duration) {
	GivenConnectGoClientWithProtocol(ctx, serverURL, protocol)
	ctx.ConnectGoTimeout = timeout
}

// Connect-Go specific When Steps

// WhenIMakeConnectGoHealthCheckRequest makes a Connect-Go health check request
func WhenIMakeConnectGoHealthCheckRequest(ctx *TestContext) {
	if ctx.ConnectGoClient == nil {
		ctx.Error = assert.AnError
		return
	}

	timeout := ctx.ConnectGoTimeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	connectCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req := connect.NewRequest(&healthv1.CheckRequest{})
	resp, err := ctx.ConnectGoClient.Check(connectCtx, req)
	ctx.ConnectGoResponse = resp
	ctx.Error = err
}

// WhenIMakeConcurrentConnectGoHealthCheckRequests makes multiple concurrent Connect-Go health check requests
func WhenIMakeConcurrentConnectGoHealthCheckRequests(ctx *TestContext, numRequests int) {
	if ctx.ConnectGoClient == nil {
		ctx.Error = assert.AnError
		return
	}

	var wg sync.WaitGroup
	results := make([]ConnectGoResult, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			timeout := ctx.ConnectGoTimeout
			if timeout == 0 {
				timeout = 5 * time.Second
			}

			connectCtx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			req := connect.NewRequest(&healthv1.CheckRequest{})
			resp, err := ctx.ConnectGoClient.Check(connectCtx, req)

			results[index] = ConnectGoResult{
				Response: resp,
				Error:    err,
				Protocol: ctx.ConnectGoProtocol,
			}
		}(i)
	}

	wg.Wait()
	ctx.ConnectGoConcurrentResults = results
}

// WhenIMakeHealthCheckRequestsWithDifferentProtocols makes requests with different protocols
func WhenIMakeHealthCheckRequestsWithDifferentProtocols(ctx *TestContext) {
	protocols := []string{"grpc-web", "http"}
	results := make([]ConnectGoResult, len(protocols))

	for i, protocol := range protocols {
		// Create client for this protocol
		GivenConnectGoClientWithProtocol(ctx, ctx.ServerURL, protocol)

		// Make request
		connectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		req := connect.NewRequest(&healthv1.CheckRequest{})
		resp, err := ctx.ConnectGoClient.Check(connectCtx, req)
		cancel()

		results[i] = ConnectGoResult{
			Response: resp,
			Error:    err,
			Protocol: protocol,
		}
	}

	ctx.ConnectGoConcurrentResults = results
}

// Connect-Go specific Then Steps

// ThenConnectGoResponseShouldIndicateServingStatus checks that Connect-Go response indicates SERVING status
func ThenConnectGoResponseShouldIndicateServingStatus(ctx *TestContext) {
	require.NoError(ctx.T, ctx.Error, "Connect-Go request should not fail")
	require.NotNil(ctx.T, ctx.ConnectGoResponse, "Connect-Go response should not be nil")
	require.NotNil(ctx.T, ctx.ConnectGoResponse.Msg, "Connect-Go response message should not be nil")
	assert.Equal(ctx.T, healthv1.ServingStatus_SERVING_STATUS_SERVING, ctx.ConnectGoResponse.Msg.Status,
		"Expected SERVING_STATUS_SERVING, got %v", ctx.ConnectGoResponse.Msg.Status)
}

// ThenConnectGoResponseShouldNotBeEmpty checks that Connect-Go response is not empty
func ThenConnectGoResponseShouldNotBeEmpty(ctx *TestContext) {
	require.NoError(ctx.T, ctx.Error, "Connect-Go request should not fail")
	require.NotNil(ctx.T, ctx.ConnectGoResponse, "Connect-Go response should not be nil")
	require.NotNil(ctx.T, ctx.ConnectGoResponse.Msg, "Connect-Go response message should not be nil")
	assert.NotEqual(ctx.T, healthv1.ServingStatus_SERVING_STATUS_UNKNOWN_UNSPECIFIED, ctx.ConnectGoResponse.Msg.Status,
		"Response status should not be unknown/unspecified")
}

// ThenAllConnectGoRequestsShouldSucceed checks that all concurrent Connect-Go requests succeeded
func ThenAllConnectGoRequestsShouldSucceed(ctx *TestContext) {
	require.NotEmpty(ctx.T, ctx.ConnectGoConcurrentResults, "No concurrent Connect-Go results found")

	for i, result := range ctx.ConnectGoConcurrentResults {
		assert.NoError(ctx.T, result.Error, "Connect-Go request %d failed with error", i+1)
		assert.NotNil(ctx.T, result.Response, "Connect-Go request %d has no response", i+1)
		assert.NotNil(ctx.T, result.Response.Msg, "Connect-Go request %d response message is nil", i+1)
	}
}

// ThenAllConnectGoResponsesShouldIndicateServingStatus checks all concurrent Connect-Go responses for SERVING status
func ThenAllConnectGoResponsesShouldIndicateServingStatus(ctx *TestContext) {
	require.NotEmpty(ctx.T, ctx.ConnectGoConcurrentResults, "No concurrent Connect-Go results found")

	for i, result := range ctx.ConnectGoConcurrentResults {
		assert.NoError(ctx.T, result.Error, "Connect-Go request %d failed with error", i+1)
		assert.NotNil(ctx.T, result.Response, "Connect-Go request %d has no response", i+1)
		assert.NotNil(ctx.T, result.Response.Msg, "Connect-Go request %d response message is nil", i+1)
		assert.Equal(ctx.T, healthv1.ServingStatus_SERVING_STATUS_SERVING, result.Response.Msg.Status,
			"Connect-Go request %d expected SERVING_STATUS_SERVING, got %v", i+1, result.Response.Msg.Status)
	}
}

// ThenAllProtocolsShouldReturnConsistentResults checks that all protocols return consistent results
func ThenAllProtocolsShouldReturnConsistentResults(ctx *TestContext) {
	require.NotEmpty(ctx.T, ctx.ConnectGoConcurrentResults, "No protocol comparison results found")
	require.GreaterOrEqual(ctx.T, len(ctx.ConnectGoConcurrentResults), 2, "Need at least 2 protocols to compare")

	// Check that all requests succeeded
	for _, result := range ctx.ConnectGoConcurrentResults {
		assert.NoError(ctx.T, result.Error, "Protocol %s request failed with error", result.Protocol)
		assert.NotNil(ctx.T, result.Response, "Protocol %s has no response", result.Protocol)
		assert.NotNil(ctx.T, result.Response.Msg, "Protocol %s response message is nil", result.Protocol)
	}

	// Check that all protocols return the same status
	expectedStatus := ctx.ConnectGoConcurrentResults[0].Response.Msg.Status
	for _, result := range ctx.ConnectGoConcurrentResults {
		assert.Equal(ctx.T, expectedStatus, result.Response.Msg.Status,
			"Protocol %s returned different status than expected. Expected: %v, Got: %v",
			result.Protocol, expectedStatus, result.Response.Msg.Status)

		// Log successful protocol test
		ctx.T.Logf("Protocol %s: Successfully returned status %v", result.Protocol, result.Response.Msg.Status)
	}
}
