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

// Common Given Steps

// GivenServerIsRunning ensures the server is running and accessible
func GivenServerIsRunning(ctx *TestContext) {
	// This step is typically handled by test setup
	// We can add a health check here if needed
	resp, err := ctx.HTTPClient.Get(fmt.Sprintf("%s/ping", ctx.ServerURL))
	if err == nil && resp != nil {
		resp.Body.Close()
	}
	// Note: We don't fail here as the server might not be ready yet
	// The actual test will fail if the server is not accessible
}

// Common When Steps

// WhenIMakeGETRequest makes a GET request to the specified endpoint
func WhenIMakeGETRequest(ctx *TestContext, endpoint string) {
	url := fmt.Sprintf("%s%s", ctx.ServerURL, endpoint)
	resp, err := ctx.HTTPClient.Get(url)

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

// WhenIMakePOSTRequest makes a POST request to the specified endpoint
func WhenIMakePOSTRequest(ctx *TestContext, endpoint string, contentType string, body string) {
	url := fmt.Sprintf("%s%s", ctx.ServerURL, endpoint)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		ctx.Error = err
		return
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	if body != "" {
		req.Body = io.NopCloser(strings.NewReader(body))
		req.ContentLength = int64(len(body))
	}

	resp, err := ctx.HTTPClient.Do(req)

	ctx.Response = resp
	ctx.Error = err

	if resp != nil {
		responseBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			ctx.Error = readErr
		} else {
			ctx.ResponseBody = responseBody
		}
		resp.Body.Close()
	}
}

// WhenIMakeConcurrentRequests makes multiple concurrent requests to an endpoint
func WhenIMakeConcurrentRequests(ctx *TestContext, numRequests int, endpoint string) {
	url := fmt.Sprintf("%s%s", ctx.ServerURL, endpoint)

	var wg sync.WaitGroup
	results := make([]ConcurrentResult, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get(url)

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

// WhenIMakeConcurrentPOSTRequests makes multiple concurrent POST requests
func WhenIMakeConcurrentPOSTRequests(ctx *TestContext, numRequests int, endpoint string, contentType string, body string) {
	url := fmt.Sprintf("%s%s", ctx.ServerURL, endpoint)

	var wg sync.WaitGroup
	results := make([]ConcurrentResult, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			client := &http.Client{Timeout: 5 * time.Second}

			req, err := http.NewRequest("POST", url, nil)
			if err != nil {
				results[index] = ConcurrentResult{Error: err}
				return
			}

			if contentType != "" {
				req.Header.Set("Content-Type", contentType)
			}

			if body != "" {
				req.Body = io.NopCloser(strings.NewReader(body))
				req.ContentLength = int64(len(body))
			}

			resp, err := client.Do(req)

			result := ConcurrentResult{
				Response: resp,
				Error:    err,
			}

			if resp != nil {
				responseBody, readErr := io.ReadAll(resp.Body)
				if readErr != nil {
					result.Error = readErr
				} else {
					result.Body = responseBody
				}
				resp.Body.Close()
			}

			results[index] = result
		}(i)
	}

	wg.Wait()
	ctx.ConcurrentResults = results
}

// Common Then Steps

// ThenHTTPStatusShouldBe checks the HTTP status code
func ThenHTTPStatusShouldBe(ctx *TestContext, expectedStatus int) {
	ctx.AssertStatusCode(expectedStatus)
}

// ThenJSONResponseShouldContainStatus checks for a specific status in JSON response
func ThenJSONResponseShouldContainStatus(ctx *TestContext, expectedStatus string) {
	ctx.AssertJSONField("status", expectedStatus)
}

// ThenResponseShouldNotBeEmpty checks that response is not empty
func ThenResponseShouldNotBeEmpty(ctx *TestContext) {
	ctx.AssertResponseNotEmpty()
}

// ThenContentTypeShouldBe checks the Content-Type header
func ThenContentTypeShouldBe(ctx *TestContext, expectedContentType string) {
	ctx.AssertContentType(expectedContentType)
}

// ThenAllRequestsShouldSucceed checks that all concurrent requests succeeded
func ThenAllRequestsShouldSucceed(ctx *TestContext) {
	require.NotEmpty(ctx.T, ctx.ConcurrentResults, "No concurrent results found")

	for i, result := range ctx.ConcurrentResults {
		assert.NoError(ctx.T, result.Error, "Request %d failed with error", i+1)
		assert.NotNil(ctx.T, result.Response, "Request %d has no response", i+1)
		if result.Response != nil {
			assert.Equal(ctx.T, http.StatusOK, result.Response.StatusCode,
				"Request %d expected status 200, got %d", i+1, result.Response.StatusCode)
		}
	}
}

// ThenAllResponsesShouldContainStatus checks that all concurrent responses contain expected status
func ThenAllResponsesShouldContainStatus(ctx *TestContext, expectedStatus string) {
	require.NotEmpty(ctx.T, ctx.ConcurrentResults, "No concurrent results found")

	for i, result := range ctx.ConcurrentResults {
		assert.NoError(ctx.T, result.Error, "Request %d failed with error", i+1)
		assert.NotNil(ctx.T, result.Response, "Request %d has no response", i+1)
		assert.NotEmpty(ctx.T, result.Body, "Request %d has empty response body", i+1)

		var jsonData map[string]interface{}
		err := json.Unmarshal(result.Body, &jsonData)
		assert.NoError(ctx.T, err, "Request %d response is not valid JSON", i+1)

		status, exists := jsonData["status"]
		assert.True(ctx.T, exists, "Request %d response does not contain 'status' field", i+1)
		assert.Equal(ctx.T, expectedStatus, status,
			"Request %d expected status '%s', got '%v'", i+1, expectedStatus, status)
	}
}

// Monitoring specific Then Steps

// ThenResponseShouldBeLightweightForMonitoring checks that response is lightweight for monitoring
func ThenResponseShouldBeLightweightForMonitoring(ctx *TestContext) {
	require.NotNil(ctx.T, ctx.Response, "No response received")
	require.NotEmpty(ctx.T, ctx.ResponseBody, "Response body is empty")

	// Check response size is small (good for monitoring)
	contentLength := len(ctx.ResponseBody)
	assert.Less(ctx.T, contentLength, 1000,
		"Response should be lightweight for monitoring, got %d bytes", contentLength)

	var jsonData map[string]interface{}
	err := json.Unmarshal(ctx.ResponseBody, &jsonData)
	require.NoError(ctx.T, err, "Response is not valid JSON")

	// Should have minimal fields for quick monitoring
	assert.LessOrEqual(ctx.T, len(jsonData), 3,
		"Monitoring response should have minimal fields, got %d fields", len(jsonData))

	// Should contain status field
	_, exists := jsonData["status"]
	assert.True(ctx.T, exists, "Monitoring response should contain 'status' field")
}
