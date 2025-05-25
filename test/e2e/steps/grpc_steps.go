package steps

import (
	"context"
	"sync"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
)

// gRPC specific Given Steps

// GivenGRPCClientIsConnected establishes a gRPC client connection
func GivenGRPCClientIsConnected(ctx *TestContext, serverAddr string) {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		ctx.Error = err
		return
	}

	ctx.GRPCConn = conn
	ctx.GRPCClient = healthv1.NewHealthServiceClient(conn)
}

// GivenGRPCClientWithTimeout establishes a gRPC client connection with timeout
func GivenGRPCClientWithTimeout(ctx *TestContext, serverAddr string, timeout time.Duration) {
	conn, err := grpc.NewClient(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		ctx.Error = err
		return
	}

	ctx.GRPCConn = conn
	ctx.GRPCClient = healthv1.NewHealthServiceClient(conn)
	ctx.GRPCTimeout = timeout
}

// gRPC specific When Steps

// WhenIMakeGRPCHealthCheckRequest makes a gRPC health check request
func WhenIMakeGRPCHealthCheckRequest(ctx *TestContext) {
	if ctx.GRPCClient == nil {
		ctx.Error = assert.AnError
		return
	}

	timeout := ctx.GRPCTimeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	grpcCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := ctx.GRPCClient.Check(grpcCtx, &healthv1.CheckRequest{})
	ctx.GRPCResponse = resp
	ctx.Error = err
}

// WhenIMakeGRPCHealthCheckRequestWithMetadata makes a gRPC health check request with metadata
func WhenIMakeGRPCHealthCheckRequestWithMetadata(ctx *TestContext) {
	if ctx.GRPCClient == nil {
		ctx.Error = assert.AnError
		return
	}

	timeout := ctx.GRPCTimeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	// Add metadata to the request
	md := metadata.New(map[string]string{
		"user-agent":     "grpc-e2e-test",
		"request-id":     "test-request-123",
		"client-version": "1.0.0",
	})

	grpcCtx := metadata.NewOutgoingContext(context.Background(), md)
	grpcCtx, cancel := context.WithTimeout(grpcCtx, timeout)
	defer cancel()

	var header metadata.MD
	resp, err := ctx.GRPCClient.Check(grpcCtx, &healthv1.CheckRequest{}, grpc.Header(&header))
	ctx.GRPCResponse = resp
	ctx.GRPCMetadata = header
	ctx.Error = err
}

// WhenIMakeConcurrentGRPCHealthCheckRequests makes multiple concurrent gRPC health check requests
func WhenIMakeConcurrentGRPCHealthCheckRequests(ctx *TestContext, numRequests int) {
	if ctx.GRPCClient == nil {
		ctx.Error = assert.AnError
		return
	}

	var wg sync.WaitGroup
	results := make([]GRPCResult, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			timeout := ctx.GRPCTimeout
			if timeout == 0 {
				timeout = 5 * time.Second
			}

			grpcCtx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			var header metadata.MD
			resp, err := ctx.GRPCClient.Check(grpcCtx, &healthv1.CheckRequest{}, grpc.Header(&header))

			results[index] = GRPCResult{
				Response: resp,
				Error:    err,
				Metadata: header,
			}
		}(i)
	}

	wg.Wait()
	ctx.GRPCConcurrentResults = results
}

// gRPC specific Then Steps

// ThenGRPCResponseShouldIndicateServingStatus checks that gRPC response indicates SERVING status
func ThenGRPCResponseShouldIndicateServingStatus(ctx *TestContext) {
	require.NoError(ctx.T, ctx.Error, "gRPC request should not fail")
	require.NotNil(ctx.T, ctx.GRPCResponse, "gRPC response should not be nil")
	assert.Equal(ctx.T, healthv1.ServingStatus_SERVING_STATUS_SERVING, ctx.GRPCResponse.Status,
		"Expected SERVING_STATUS_SERVING, got %v", ctx.GRPCResponse.Status)
}

// ThenGRPCResponseShouldNotBeEmpty checks that gRPC response is not empty
func ThenGRPCResponseShouldNotBeEmpty(ctx *TestContext) {
	require.NoError(ctx.T, ctx.Error, "gRPC request should not fail")
	require.NotNil(ctx.T, ctx.GRPCResponse, "gRPC response should not be nil")
	assert.NotEqual(ctx.T, healthv1.ServingStatus_SERVING_STATUS_UNKNOWN_UNSPECIFIED, ctx.GRPCResponse.Status,
		"Response status should not be unknown/unspecified")
}

// ThenAllGRPCRequestsShouldSucceed checks that all concurrent gRPC requests succeeded
func ThenAllGRPCRequestsShouldSucceed(ctx *TestContext) {
	require.NotEmpty(ctx.T, ctx.GRPCConcurrentResults, "No concurrent gRPC results found")

	for i, result := range ctx.GRPCConcurrentResults {
		assert.NoError(ctx.T, result.Error, "gRPC request %d failed with error", i+1)
		assert.NotNil(ctx.T, result.Response, "gRPC request %d has no response", i+1)
	}
}

// ThenAllGRPCResponsesShouldIndicateServingStatus checks all concurrent gRPC responses for SERVING status
func ThenAllGRPCResponsesShouldIndicateServingStatus(ctx *TestContext) {
	require.NotEmpty(ctx.T, ctx.GRPCConcurrentResults, "No concurrent gRPC results found")

	for i, result := range ctx.GRPCConcurrentResults {
		assert.NoError(ctx.T, result.Error, "gRPC request %d failed with error", i+1)
		assert.NotNil(ctx.T, result.Response, "gRPC request %d has no response", i+1)
		assert.Equal(ctx.T, healthv1.ServingStatus_SERVING_STATUS_SERVING, result.Response.Status,
			"gRPC request %d expected SERVING_STATUS_SERVING, got %v", i+1, result.Response.Status)
	}
}

// ThenGRPCResponseShouldContainMetadata checks that gRPC response contains metadata
func ThenGRPCResponseShouldContainMetadata(ctx *TestContext) {
	require.NoError(ctx.T, ctx.Error, "gRPC request should not fail")
	require.NotNil(ctx.T, ctx.GRPCResponse, "gRPC response should not be nil")

	// Check that we received some metadata (even if empty, the map should exist)
	assert.NotNil(ctx.T, ctx.GRPCMetadata, "gRPC metadata should not be nil")

	// Note: We don't assert specific metadata content as the server might not return
	// specific metadata, but we verify that the metadata mechanism is working
}

// ThenGRPCConnectionShouldBeEstablished checks that gRPC connection is established
func ThenGRPCConnectionShouldBeEstablished(ctx *TestContext) {
	require.NotNil(ctx.T, ctx.GRPCConn, "gRPC connection should be established")
	require.NotNil(ctx.T, ctx.GRPCClient, "gRPC client should be created")

	// Test connection by making a simple call
	grpcCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := ctx.GRPCClient.Check(grpcCtx, &healthv1.CheckRequest{})
	assert.NoError(ctx.T, err, "gRPC connection should be working")
}

// ThenGRPCConnectionShouldBeClosed checks that gRPC connection is properly closed
func ThenGRPCConnectionShouldBeClosed(ctx *TestContext) {
	if ctx.GRPCConn != nil {
		err := ctx.GRPCConn.Close()
		assert.NoError(ctx.T, err, "gRPC connection should close without error")
	}
}
