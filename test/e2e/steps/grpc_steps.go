package steps

import (
	"context"
	"sync"
	"time"

	"github.com/stretchr/testify/assert"
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

// ThenGRPCResponseShouldIndicateServingStatus checks that gRPC response indicates SERVING status in BDD style
func ThenGRPCResponseShouldIndicateServingStatus(ctx *TestContext) {
	ctx.NoErrorShouldHaveOccurred()
	ctx.TheGRPCResponseShouldBeSuccessful()

	if ctx.GRPCResponse != nil && ctx.GRPCResponse.Status != healthv1.ServingStatus_SERVING_STATUS_SERVING {
		ctx.T.Errorf("Expected gRPC response status to be SERVING_STATUS_SERVING, but got %v", ctx.GRPCResponse.Status)
	}
}

// ThenGRPCResponseShouldNotBeEmpty checks that gRPC response is not empty in BDD style
func ThenGRPCResponseShouldNotBeEmpty(ctx *TestContext) {
	ctx.NoErrorShouldHaveOccurred()
	ctx.TheGRPCResponseShouldBeSuccessful()

	if ctx.GRPCResponse != nil && ctx.GRPCResponse.Status == healthv1.ServingStatus_SERVING_STATUS_UNKNOWN_UNSPECIFIED {
		ctx.T.Errorf("Expected gRPC response status to not be unknown/unspecified, but got %v", ctx.GRPCResponse.Status)
	}
}

// ThenAllGRPCRequestsShouldSucceed checks that all concurrent gRPC requests succeeded in BDD style
func ThenAllGRPCRequestsShouldSucceed(ctx *TestContext) {
	if len(ctx.GRPCConcurrentResults) == 0 {
		ctx.T.Errorf("Expected concurrent gRPC results to exist, but none were found")
		return
	}

	failedCount := 0
	for i, result := range ctx.GRPCConcurrentResults {
		if result.Error != nil {
			ctx.T.Errorf("Expected concurrent gRPC request %d to succeed, but got error: %v", i+1, result.Error)
			failedCount++
			continue
		}
		if result.Response == nil {
			ctx.T.Errorf("Expected concurrent gRPC request %d to have a response, but none was received", i+1)
			failedCount++
		}
	}

	if failedCount > 0 {
		ctx.T.Errorf("Expected all %d concurrent gRPC requests to succeed, but %d failed", len(ctx.GRPCConcurrentResults), failedCount)
	}
}

// ThenAllGRPCResponsesShouldIndicateServingStatus checks all concurrent gRPC responses for SERVING status in BDD style
func ThenAllGRPCResponsesShouldIndicateServingStatus(ctx *TestContext) {
	if len(ctx.GRPCConcurrentResults) == 0 {
		ctx.T.Errorf("Expected concurrent gRPC results to exist for serving status check, but none were found")
		return
	}

	failedCount := 0
	for i, result := range ctx.GRPCConcurrentResults {
		if result.Error != nil {
			ctx.T.Errorf("Expected concurrent gRPC request %d to succeed for serving status check, but got error: %v", i+1, result.Error)
			failedCount++
			continue
		}
		if result.Response == nil {
			ctx.T.Errorf("Expected concurrent gRPC request %d to have a response for serving status check, but none was received", i+1)
			failedCount++
			continue
		}
		if result.Response.Status != healthv1.ServingStatus_SERVING_STATUS_SERVING {
			ctx.T.Errorf("Expected concurrent gRPC request %d to have SERVING_STATUS_SERVING, but got %v", i+1, result.Response.Status)
			failedCount++
		}
	}

	if failedCount > 0 {
		ctx.T.Errorf("Expected all %d concurrent gRPC responses to indicate serving status, but %d failed", len(ctx.GRPCConcurrentResults), failedCount)
	}
}

// ThenGRPCResponseShouldContainMetadata checks that gRPC response contains metadata in BDD style
func ThenGRPCResponseShouldContainMetadata(ctx *TestContext) {
	ctx.NoErrorShouldHaveOccurred()
	ctx.TheGRPCResponseShouldBeSuccessful()

	if ctx.GRPCMetadata == nil {
		ctx.T.Errorf("Expected gRPC metadata to exist, but it was nil")
	}
	// Note: We don't assert specific metadata content as the server might not return
	// specific metadata, but we verify that the metadata mechanism is working
}

// ThenGRPCConnectionShouldBeEstablished checks that gRPC connection is established in BDD style
func ThenGRPCConnectionShouldBeEstablished(ctx *TestContext) {
	if ctx.GRPCConn == nil {
		ctx.T.Errorf("Expected gRPC connection to be established, but it was nil")
		return
	}
	if ctx.GRPCClient == nil {
		ctx.T.Errorf("Expected gRPC client to be created, but it was nil")
		return
	}

	// Test connection by making a simple call
	grpcCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := ctx.GRPCClient.Check(grpcCtx, &healthv1.CheckRequest{})
	if err != nil {
		ctx.T.Errorf("Expected gRPC connection to be working, but got error: %v", err)
	}
}

// ThenGRPCConnectionShouldBeClosed checks that gRPC connection is properly closed in BDD style
func ThenGRPCConnectionShouldBeClosed(ctx *TestContext) {
	if ctx.GRPCConn != nil {
		err := ctx.GRPCConn.Close()
		if err != nil {
			ctx.T.Errorf("Expected gRPC connection to close without error, but got: %v", err)
		}
	}
}
