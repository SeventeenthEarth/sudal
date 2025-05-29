package steps

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Connect-Go specific When Steps

// WhenIMakeHealthCheckRequestUsingConnectGo makes a health check request using Connect-Go protocol
func WhenIMakeHealthCheckRequestUsingConnectGo(ctx *TestContext) {
	url := fmt.Sprintf("%s/health.v1.HealthService/Check", ctx.ServerURL)

	req, err := http.NewRequest("POST", url, strings.NewReader("{}"))
	if err != nil {
		ctx.Error = err
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ctx.HTTPClient.Do(req)
	ctx.Response = resp
	ctx.Error = err

	if resp != nil {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			ctx.Error = readErr
		} else {
			ctx.ResponseBody = body
		}
		resp.Body.Close() // nolint:errcheck
	}
}

// WhenIMakeHealthCheckRequestUsingHTTPJSON makes a health check request using HTTP/JSON
func WhenIMakeHealthCheckRequestUsingHTTPJSON(ctx *TestContext) {
	// This is essentially the same as Connect-Go for HTTP/JSON
	WhenIMakeHealthCheckRequestUsingConnectGo(ctx)
}

// WhenIMakeHealthCheckRequestWithInvalidContentType makes a request with invalid content type
func WhenIMakeHealthCheckRequestWithInvalidContentType(ctx *TestContext) {
	url := fmt.Sprintf("%s/health.v1.HealthService/Check", ctx.ServerURL)

	req, err := http.NewRequest("POST", url, strings.NewReader("{}"))
	if err != nil {
		ctx.Error = err
		return
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := ctx.HTTPClient.Do(req)
	ctx.Response = resp
	ctx.Error = err

	if resp != nil {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			ctx.Error = readErr
		} else {
			ctx.ResponseBody = body
		}
		resp.Body.Close() // nolint:errcheck
	}
}

// WhenIMakeRequestToNonExistentEndpoint makes a request to a non-existent endpoint
func WhenIMakeRequestToNonExistentEndpoint(ctx *TestContext) {
	url := fmt.Sprintf("%s/health.v1.HealthService/NonExistentMethod", ctx.ServerURL)

	req, err := http.NewRequest("POST", url, strings.NewReader("{}"))
	if err != nil {
		ctx.Error = err
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ctx.HTTPClient.Do(req)
	ctx.Response = resp
	ctx.Error = err

	if resp != nil {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			ctx.Error = readErr
		} else {
			ctx.ResponseBody = body
		}
		resp.Body.Close() // nolint:errcheck
	}
}

// WhenIMakeConcurrentHealthCheckRequests makes multiple concurrent health check requests
func WhenIMakeConcurrentHealthCheckRequests(ctx *TestContext, numRequests int) {
	url := fmt.Sprintf("%s/health.v1.HealthService/Check", ctx.ServerURL)

	var wg sync.WaitGroup
	results := make([]ConcurrentResult, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			client := &http.Client{Timeout: 5 * time.Second}

			req, err := http.NewRequest("POST", url, strings.NewReader("{}"))
			if err != nil {
				results[index] = ConcurrentResult{Error: err}
				return
			}

			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)

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
				resp.Body.Close() // nolint:errcheck
			}

			results[index] = result
		}(i)
	}

	wg.Wait()
	ctx.ConcurrentResults = results
}

// Connect-Go specific Then Steps - BDD Style

// ThenResponseShouldIndicateServingStatus checks that response indicates SERVING status in BDD style
func ThenResponseShouldIndicateServingStatus(ctx *TestContext) {
	ctx.TheResponseStatusCodeShouldBe(http.StatusOK)
	ctx.TheResponseShouldNotBeEmpty()
	ctx.TheJSONResponseShouldContainField("status", "SERVING_STATUS_SERVING")
}

// ThenJSONResponseShouldContainServingStatusServing checks for SERVING_STATUS_SERVING in JSON in BDD style
func ThenJSONResponseShouldContainServingStatusServing(ctx *TestContext) {
	ctx.TheJSONResponseShouldContainField("status", "SERVING_STATUS_SERVING")
}

// ThenServerShouldRejectRequest checks that server rejected the request in BDD style
func ThenServerShouldRejectRequest(ctx *TestContext) {
	ctx.TheResponseStatusCodeShouldBe(http.StatusUnsupportedMediaType)
}

// ThenResponseShouldContainProperConnectGoHeaders checks for proper Connect-Go headers in BDD style
func ThenResponseShouldContainProperConnectGoHeaders(ctx *TestContext) {
	ctx.TheResponseHeaderShouldContain("Content-Type", "application/json")
}

// Legacy functions for backward compatibility
func ThenResponseShouldIndicateServingStatusLegacy(ctx *TestContext) {
	ctx.AssertStatusCode(http.StatusOK)
	ctx.AssertResponseNotEmpty()

	var jsonData map[string]interface{}
	err := json.Unmarshal(ctx.ResponseBody, &jsonData)
	require.NoError(ctx.T, err, "Response is not valid JSON")

	status, exists := jsonData["status"]
	require.True(ctx.T, exists, "Response does not contain 'status' field")
	assert.Equal(ctx.T, "SERVING_STATUS_SERVING", status,
		"Expected SERVING_STATUS_SERVING, got %v", status)
}

// ThenAllResponsesShouldIndicateServingStatus checks all concurrent responses for SERVING status in BDD style
func ThenAllResponsesShouldIndicateServingStatus(ctx *TestContext) {
	if len(ctx.ConcurrentResults) == 0 {
		ctx.T.Errorf("Expected concurrent results to exist, but none were found")
		return
	}

	failedCount := 0
	for i, result := range ctx.ConcurrentResults {
		if result.Error != nil {
			ctx.T.Errorf("Expected concurrent request %d to succeed, but got error: %v", i+1, result.Error)
			failedCount++
			continue
		}
		if result.Response == nil {
			ctx.T.Errorf("Expected concurrent request %d to have a response, but none was received", i+1)
			failedCount++
			continue
		}
		if result.Response.StatusCode != http.StatusOK {
			ctx.T.Errorf("Expected concurrent request %d to have status 200, but got %d", i+1, result.Response.StatusCode)
			failedCount++
			continue
		}

		var jsonData map[string]interface{}
		err := json.Unmarshal(result.Body, &jsonData)
		if err != nil {
			ctx.T.Errorf("Expected concurrent request %d response to be valid JSON, but got error: %v", i+1, err)
			failedCount++
			continue
		}

		status, exists := jsonData["status"]
		if !exists {
			ctx.T.Errorf("Expected concurrent request %d response to contain 'status' field, but it was missing", i+1)
			failedCount++
			continue
		}
		if status != "SERVING_STATUS_SERVING" {
			ctx.T.Errorf("Expected concurrent request %d to have SERVING_STATUS_SERVING, but got %v", i+1, status)
			failedCount++
		}
	}

	if failedCount > 0 {
		ctx.T.Errorf("Expected all %d concurrent requests to indicate serving status, but %d failed", len(ctx.ConcurrentResults), failedCount)
	}
}

// ThenResponseShouldContainProperConnectGoHeadersLegacy checks for proper Connect-Go headers (legacy)
func ThenResponseShouldContainProperConnectGoHeadersLegacy(ctx *TestContext) {
	require.NotNil(ctx.T, ctx.Response, "No response received")

	contentType := ctx.Response.Header.Get("Content-Type")
	assert.Contains(ctx.T, contentType, "application/json",
		"Expected JSON content type for Connect-Go, got: %s", contentType)
}
