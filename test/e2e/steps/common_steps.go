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

// Common Then Steps - BDD Style

// ThenHTTPStatusShouldBe checks the HTTP status code in BDD style
func ThenHTTPStatusShouldBe(ctx *TestContext, expectedStatus int) {
	ctx.TheResponseStatusCodeShouldBe(expectedStatus)
}

// ThenJSONResponseShouldContainStatus checks for a specific status in JSON response in BDD style
func ThenJSONResponseShouldContainStatus(ctx *TestContext, expectedStatus string) {
	ctx.TheJSONResponseShouldContainField("status", expectedStatus)
}

// ThenResponseShouldNotBeEmpty checks that response is not empty in BDD style
func ThenResponseShouldNotBeEmpty(ctx *TestContext) {
	ctx.TheResponseShouldNotBeEmpty()
}

// ThenContentTypeShouldBe checks the Content-Type header in BDD style
func ThenContentTypeShouldBe(ctx *TestContext, expectedContentType string) {
	ctx.TheContentTypeShouldBe(expectedContentType)
}

// Legacy functions for backward compatibility (will be deprecated)
// ThenHTTPStatusShouldBeLegacy checks the HTTP status code using testify
func ThenHTTPStatusShouldBeLegacy(ctx *TestContext, expectedStatus int) {
	ctx.AssertStatusCode(expectedStatus)
}

// ThenAllRequestsShouldSucceed checks that all concurrent requests succeeded in BDD style
func ThenAllRequestsShouldSucceed(ctx *TestContext) {
	if len(ctx.ConcurrentResults) == 0 {
		ctx.T.Errorf("Expected concurrent results to exist, but none were found")
		return
	}

	for i, result := range ctx.ConcurrentResults {
		if result.Error != nil {
			ctx.T.Errorf("Expected request %d to succeed, but got error: %v", i+1, result.Error)
			continue
		}
		if result.Response == nil {
			ctx.T.Errorf("Expected request %d to have a response, but none was received", i+1)
			continue
		}
		if result.Response.StatusCode != http.StatusOK {
			ctx.T.Errorf("Expected request %d to have status 200, but got %d", i+1, result.Response.StatusCode)
		}
	}
}

// Legacy version using testify
func ThenAllRequestsShouldSucceedLegacy(ctx *TestContext) {
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

// ThenAllResponsesShouldContainStatus checks that all concurrent responses contain expected status in BDD style
func ThenAllResponsesShouldContainStatus(ctx *TestContext, expectedStatus string) {
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
		if len(result.Body) == 0 {
			ctx.T.Errorf("Expected concurrent request %d to have response body, but it was empty", i+1)
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
		if status != expectedStatus {
			ctx.T.Errorf("Expected concurrent request %d status to be '%s', but got '%v'", i+1, expectedStatus, status)
			failedCount++
		}
	}

	if failedCount > 0 {
		ctx.T.Errorf("Expected all %d concurrent requests to contain status '%s', but %d failed", len(ctx.ConcurrentResults), expectedStatus, failedCount)
	}
}

// Monitoring specific Then Steps - BDD Style

// ThenResponseShouldBeLightweightForMonitoring checks that response is lightweight for monitoring in BDD style
func ThenResponseShouldBeLightweightForMonitoring(ctx *TestContext) {
	if ctx.Response == nil {
		ctx.T.Errorf("Expected response to exist for monitoring check, but none was received")
		return
	}
	if len(ctx.ResponseBody) == 0 {
		ctx.T.Errorf("Expected response body to contain data for monitoring check, but it was empty")
		return
	}

	// Check response size is small (good for monitoring)
	contentLength := len(ctx.ResponseBody)
	if contentLength >= 1000 {
		ctx.T.Errorf("Expected response to be lightweight for monitoring (< 1000 bytes), but got %d bytes", contentLength)
		return
	}

	var jsonData map[string]interface{}
	err := json.Unmarshal(ctx.ResponseBody, &jsonData)
	if err != nil {
		ctx.T.Errorf("Expected response to be valid JSON for monitoring, but got error: %v", err)
		return
	}

	// Should have minimal fields for quick monitoring
	if len(jsonData) > 3 {
		ctx.T.Errorf("Expected monitoring response to have minimal fields (≤ 3), but got %d fields", len(jsonData))
	}

	// Should contain status field
	if _, exists := jsonData["status"]; !exists {
		ctx.T.Errorf("Expected monitoring response to contain 'status' field, but it was missing")
	}
}

// Legacy monitoring function for backward compatibility
func ThenResponseShouldBeLightweightForMonitoringLegacy(ctx *TestContext) {
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
