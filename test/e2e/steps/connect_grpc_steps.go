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

// ThenConnectGoResponseShouldIndicateServingStatus checks that Connect-Go response indicates SERVING status in BDD style
func ThenConnectGoResponseShouldIndicateServingStatus(ctx *TestContext) {
	ctx.NoErrorShouldHaveOccurred()
	ctx.TheConnectGoResponseShouldBeSuccessful()

	if ctx.ConnectGoResponse != nil && ctx.ConnectGoResponse.Msg != nil &&
		ctx.ConnectGoResponse.Msg.Status != healthv1.ServingStatus_SERVING_STATUS_SERVING {
		ctx.T.Errorf("Expected Connect-Go response status to be SERVING_STATUS_SERVING, but got %v", ctx.ConnectGoResponse.Msg.Status)
	}
}

// ThenConnectGoResponseShouldNotBeEmpty checks that Connect-Go response is not empty in BDD style
func ThenConnectGoResponseShouldNotBeEmpty(ctx *TestContext) {
	ctx.NoErrorShouldHaveOccurred()
	ctx.TheConnectGoResponseShouldBeSuccessful()

	if ctx.ConnectGoResponse != nil && ctx.ConnectGoResponse.Msg != nil &&
		ctx.ConnectGoResponse.Msg.Status == healthv1.ServingStatus_SERVING_STATUS_UNKNOWN_UNSPECIFIED {
		ctx.T.Errorf("Expected Connect-Go response status to not be unknown/unspecified, but got %v", ctx.ConnectGoResponse.Msg.Status)
	}
}

// ThenConnectGoRequestShouldFail checks that Connect-Go request failed as expected in BDD style
func ThenConnectGoRequestShouldFail(ctx *TestContext) {
	if ctx.Error == nil {
		ctx.T.Errorf("Expected Connect-Go request to fail, but it succeeded")
		return
	}
	// Log the expected failure
	ctx.T.Logf("Connect-Go request failed as expected: %v", ctx.Error)
}

// ThenGRPCWebShouldSucceedAndHTTPShouldFail checks that gRPC-Web succeeds while HTTP/JSON fails
func ThenGRPCWebShouldSucceedAndHTTPShouldFail(ctx *TestContext) {
	if len(ctx.ConnectGoConcurrentResults) == 0 {
		ctx.T.Errorf("Expected protocol comparison results to exist, but none were found")
		return
	}

	grpcWebFound := false
	httpFound := false

	for _, result := range ctx.ConnectGoConcurrentResults {
		switch result.Protocol {
		case "grpc-web":
			grpcWebFound = true
			if result.Error != nil {
				ctx.T.Errorf("Expected gRPC-Web request to succeed, but got error: %v", result.Error)
			} else if result.Response == nil {
				ctx.T.Errorf("Expected gRPC-Web request to have a response, but none was received")
			} else {
				ctx.T.Logf("gRPC-Web request succeeded as expected")
			}
		case "http":
			httpFound = true
			if result.Error == nil {
				ctx.T.Errorf("Expected HTTP/JSON request to fail for gRPC-only endpoint, but it succeeded")
			} else {
				ctx.T.Logf("HTTP/JSON request failed as expected: %v", result.Error)
			}
		}
	}

	if !grpcWebFound {
		ctx.T.Errorf("Expected to find gRPC-Web protocol test result, but none was found")
	}
	if !httpFound {
		ctx.T.Errorf("Expected to find HTTP/JSON protocol test result, but none was found")
	}
}

// ThenAllConnectGoRequestsShouldSucceed checks that all concurrent Connect-Go requests succeeded in BDD style
func ThenAllConnectGoRequestsShouldSucceed(ctx *TestContext) {
	if len(ctx.ConnectGoConcurrentResults) == 0 {
		ctx.T.Errorf("Expected concurrent Connect-Go results to exist, but none were found")
		return
	}

	failedCount := 0
	for i, result := range ctx.ConnectGoConcurrentResults {
		if result.Error != nil {
			ctx.T.Errorf("Expected concurrent Connect-Go request %d to succeed, but got error: %v", i+1, result.Error)
			failedCount++
			continue
		}
		if result.Response == nil {
			ctx.T.Errorf("Expected concurrent Connect-Go request %d to have a response, but none was received", i+1)
			failedCount++
			continue
		}
		if result.Response.Msg == nil {
			ctx.T.Errorf("Expected concurrent Connect-Go request %d response message to exist, but it was nil", i+1)
			failedCount++
		}
	}

	if failedCount > 0 {
		ctx.T.Errorf("Expected all %d concurrent Connect-Go requests to succeed, but %d failed", len(ctx.ConnectGoConcurrentResults), failedCount)
	}
}

// ThenAllConnectGoRequestsShouldFail checks that all concurrent Connect-Go requests failed as expected in BDD style
func ThenAllConnectGoRequestsShouldFail(ctx *TestContext) {
	if len(ctx.ConnectGoConcurrentResults) == 0 {
		ctx.T.Errorf("Expected concurrent Connect-Go results to exist, but none were found")
		return
	}

	successCount := 0
	for i, result := range ctx.ConnectGoConcurrentResults {
		if result.Error == nil {
			ctx.T.Errorf("Expected concurrent Connect-Go request %d to fail, but it succeeded", i+1)
			successCount++
		} else {
			ctx.T.Logf("Concurrent Connect-Go request %d failed as expected: %v", i+1, result.Error)
		}
	}

	if successCount > 0 {
		ctx.T.Errorf("Expected all %d concurrent Connect-Go requests to fail, but %d succeeded", len(ctx.ConnectGoConcurrentResults), successCount)
	}
}

// ThenAllConnectGoResponsesShouldIndicateServingStatus checks all concurrent Connect-Go responses for SERVING status
func ThenAllConnectGoResponsesShouldIndicateServingStatus(ctx *TestContext) {
	if len(ctx.ConnectGoConcurrentResults) == 0 {
		ctx.T.Errorf("Expected concurrent Connect-Go results to exist, but none were found")
		return
	}

	failedCount := 0
	for i, result := range ctx.ConnectGoConcurrentResults {
		if result.Error != nil {
			ctx.T.Errorf("Expected concurrent Connect-Go request %d to succeed, but got error: %v", i+1, result.Error)
			failedCount++
			continue
		}
		if result.Response == nil {
			ctx.T.Errorf("Expected concurrent Connect-Go request %d to have a response, but none was received", i+1)
			failedCount++
			continue
		}
		if result.Response.Msg == nil {
			ctx.T.Errorf("Expected concurrent Connect-Go request %d response message to exist, but it was nil", i+1)
			failedCount++
			continue
		}
		if result.Response.Msg.Status != healthv1.ServingStatus_SERVING_STATUS_SERVING {
			ctx.T.Errorf("Expected concurrent Connect-Go request %d status to be SERVING_STATUS_SERVING, but got %v", i+1, result.Response.Msg.Status)
			failedCount++
		}
	}

	if failedCount > 0 {
		ctx.T.Errorf("Expected all %d concurrent Connect-Go responses to indicate serving status, but %d failed", len(ctx.ConnectGoConcurrentResults), failedCount)
	}
}

// ThenAllProtocolsShouldReturnConsistentResults checks that all protocols return consistent results
func ThenAllProtocolsShouldReturnConsistentResults(ctx *TestContext) {
	if len(ctx.ConnectGoConcurrentResults) == 0 {
		ctx.T.Errorf("Expected protocol comparison results to exist, but none were found")
		return
	}
	if len(ctx.ConnectGoConcurrentResults) < 2 {
		ctx.T.Errorf("Expected at least 2 protocols to compare, but got %d", len(ctx.ConnectGoConcurrentResults))
		return
	}

	// Check that all requests succeeded
	failedCount := 0
	for _, result := range ctx.ConnectGoConcurrentResults {
		if result.Error != nil {
			ctx.T.Errorf("Expected protocol %s request to succeed, but got error: %v", result.Protocol, result.Error)
			failedCount++
			continue
		}
		if result.Response == nil {
			ctx.T.Errorf("Expected protocol %s to have a response, but none was received", result.Protocol)
			failedCount++
			continue
		}
		if result.Response.Msg == nil {
			ctx.T.Errorf("Expected protocol %s response message to exist, but it was nil", result.Protocol)
			failedCount++
		}
	}

	if failedCount > 0 {
		ctx.T.Errorf("Expected all protocols to succeed, but %d failed", failedCount)
		return
	}

	// Check that all protocols return the same status
	expectedStatus := ctx.ConnectGoConcurrentResults[0].Response.Msg.Status
	for _, result := range ctx.ConnectGoConcurrentResults {
		if result.Response.Msg.Status != expectedStatus {
			ctx.T.Errorf("Expected protocol %s to return status %v, but got %v", result.Protocol, expectedStatus, result.Response.Msg.Status)
		} else {
			// Log successful protocol test
			ctx.T.Logf("Protocol %s: Successfully returned status %v", result.Protocol, result.Response.Msg.Status)
		}
	}
}
