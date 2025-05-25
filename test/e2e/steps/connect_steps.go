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
		resp.Body.Close()
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
		resp.Body.Close()
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
		resp.Body.Close()
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
				resp.Body.Close()
			}

			results[index] = result
		}(i)
	}

	wg.Wait()
	ctx.ConcurrentResults = results
}

// Connect-Go specific Then Steps

// ThenResponseShouldIndicateServingStatus checks that response indicates SERVING status
func ThenResponseShouldIndicateServingStatus(ctx *TestContext) {
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

// ThenJSONResponseShouldContainServingStatusServing checks for SERVING_STATUS_SERVING in JSON
func ThenJSONResponseShouldContainServingStatusServing(ctx *TestContext) {
	ctx.AssertJSONField("status", "SERVING_STATUS_SERVING")
}

// ThenServerShouldRejectRequest checks that server rejected the request
func ThenServerShouldRejectRequest(ctx *TestContext) {
	require.NotNil(ctx.T, ctx.Response, "No response received")
	assert.Equal(ctx.T, http.StatusUnsupportedMediaType, ctx.Response.StatusCode,
		"Expected status 415 for rejected request, got %d", ctx.Response.StatusCode)
}

// ThenAllResponsesShouldIndicateServingStatus checks all concurrent responses for SERVING status
func ThenAllResponsesShouldIndicateServingStatus(ctx *TestContext) {
	require.NotEmpty(ctx.T, ctx.ConcurrentResults, "No concurrent results found")

	for i, result := range ctx.ConcurrentResults {
		assert.NoError(ctx.T, result.Error, "Request %d failed with error", i+1)
		assert.NotNil(ctx.T, result.Response, "Request %d has no response", i+1)
		assert.Equal(ctx.T, http.StatusOK, result.Response.StatusCode,
			"Request %d expected status 200, got %d", i+1, result.Response.StatusCode)

		var jsonData map[string]interface{}
		err := json.Unmarshal(result.Body, &jsonData)
		assert.NoError(ctx.T, err, "Request %d response is not valid JSON", i+1)

		status, exists := jsonData["status"]
		assert.True(ctx.T, exists, "Request %d response does not contain 'status' field", i+1)
		assert.Equal(ctx.T, "SERVING_STATUS_SERVING", status,
			"Request %d expected SERVING_STATUS_SERVING, got %v", i+1, status)
	}
}

// ThenResponseShouldContainProperConnectGoHeaders checks for proper Connect-Go headers
func ThenResponseShouldContainProperConnectGoHeaders(ctx *TestContext) {
	require.NotNil(ctx.T, ctx.Response, "No response received")

	contentType := ctx.Response.Header.Get("Content-Type")
	assert.Contains(ctx.T, contentType, "application/json",
		"Expected JSON content type for Connect-Go, got: %s", contentType)
}
