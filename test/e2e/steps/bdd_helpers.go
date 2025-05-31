package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
	"github.com/seventeenthearth/sudal/gen/go/health/v1/healthv1connect"
)

// CacheTestContext holds cache-specific test context
type CacheTestContext struct {
	CacheUtil     interface{} // Will be *cacheutil.CacheUtil but avoiding import cycle
	LastError     error
	LastValue     string
	TestKeyPrefix string
	CreatedKeys   []string
	mu            interface{} // Will be sync.Mutex but avoiding import
}

// HTTPTestContext holds HTTP-specific test context
type HTTPTestContext struct {
	RequestEndpoint    string
	RequestContentType string
	RequestBody        string
	LastResponse       *http.Response
	LastResponseBody   []byte
	LastError          error
}

// UserTestContext holds user-specific test context
type UserTestContext struct {
	UserClient                interface{} // Will be userv1connect.UserServiceClient but avoiding import cycle
	RegisterUserRequest       interface{} // Will be *userv1.RegisterUserRequest
	RegisterUserResponse      interface{} // Will be *connect.Response[userv1.RegisterUserResponse]
	GetUserProfileRequest     interface{} // Will be *userv1.GetUserProfileRequest
	GetUserProfileResponse    interface{} // Will be *connect.Response[userv1.UserProfile]
	UpdateUserProfileRequest  interface{} // Will be *userv1.UpdateUserProfileRequest
	UpdateUserProfileResponse interface{} // Will be *connect.Response[userv1.UpdateUserProfileResponse]
	LastError                 error
	CreatedUserID             string
	TestFirebaseUID           string
	TestDisplayName           string
	TestAvatarURL             string
	Protocol                  string
	Timeout                   time.Duration
	ConcurrentResults         []interface{} // Will be []UserResult but avoiding import cycle
}

// TestContext holds the context for BDD test scenarios
type TestContext struct {
	T                 *testing.T
	ServerURL         string
	HTTPClient        *http.Client
	Response          *http.Response
	ResponseBody      []byte
	Error             error
	ConcurrentResults []ConcurrentResult
	Context           context.Context
	// gRPC related fields
	GRPCConn              *grpc.ClientConn
	GRPCClient            healthv1.HealthServiceClient
	GRPCResponse          *healthv1.CheckResponse
	GRPCMetadata          metadata.MD
	GRPCTimeout           time.Duration
	GRPCConcurrentResults []GRPCResult
	// Connect-Go related fields
	ConnectGoClient            healthv1connect.HealthServiceClient
	ConnectGoResponse          *connect.Response[healthv1.CheckResponse]
	ConnectGoProtocol          string
	ConnectGoTimeout           time.Duration
	ConnectGoConcurrentResults []ConnectGoResult
	// Cache related fields
	CacheTestContext *CacheTestContext
	// User related fields
	UserTestContext *UserTestContext
	// HTTP related fields
	HTTPTestContext *HTTPTestContext
}

// ConcurrentResult holds the result of a concurrent request
type ConcurrentResult struct {
	Response *http.Response
	Body     []byte
	Error    error
}

// GRPCResult represents the result of a gRPC call
type GRPCResult struct {
	Response *healthv1.CheckResponse
	Error    error
	Metadata metadata.MD
}

// ConnectGoResult represents the result of a Connect-Go call
type ConnectGoResult struct {
	Response *connect.Response[healthv1.CheckResponse]
	Error    error
	Protocol string
}

// NewTestContext creates a new test context for BDD scenarios
func NewTestContext(t *testing.T, serverURL string) *TestContext {
	return &TestContext{
		T:         t,
		ServerURL: serverURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		Context: context.Background(),
	}
}

// BDDScenario represents a BDD scenario with Given-When-Then structure
type BDDScenario struct {
	Name        string
	Description string
	Given       func(*TestContext)
	When        func(*TestContext)
	Then        func(*TestContext)
}

// RunBDDScenario executes a BDD scenario with proper structure
func RunBDDScenario(t *testing.T, serverURL string, scenario BDDScenario) {
	t.Run(scenario.Name, func(t *testing.T) {
		ctx := NewTestContext(t, serverURL)

		// Given
		if scenario.Given != nil {
			scenario.Given(ctx)
		}

		// When
		if scenario.When != nil {
			scenario.When(ctx)
		}

		// Then
		if scenario.Then != nil {
			scenario.Then(ctx)
		}
	})
}

// RunBDDScenarios executes multiple BDD scenarios
func RunBDDScenarios(t *testing.T, serverURL string, scenarios []BDDScenario) {
	for _, scenario := range scenarios {
		RunBDDScenario(t, serverURL, scenario)
	}
}

// TableDrivenBDDTest represents a table-driven BDD test
type TableDrivenBDDTest struct {
	Name     string
	TestData interface{}
	Given    func(*TestContext, interface{})
	When     func(*TestContext, interface{})
	Then     func(*TestContext, interface{})
}

// RunTableDrivenBDDTest executes a table-driven BDD test
func RunTableDrivenBDDTest(t *testing.T, serverURL string, test TableDrivenBDDTest, testCases []interface{}) {
	t.Run(test.Name, func(t *testing.T) {
		for i, testCase := range testCases {
			t.Run(fmt.Sprintf("case_%d", i+1), func(t *testing.T) {
				ctx := NewTestContext(t, serverURL)

				// Given
				if test.Given != nil {
					test.Given(ctx, testCase)
				}

				// When
				if test.When != nil {
					test.When(ctx, testCase)
				}

				// Then
				if test.Then != nil {
					test.Then(ctx, testCase)
				}
			})
		}
	})
}

// BDD-style assertion methods

// TheResponseStatusCodeShouldBe checks the HTTP status code in BDD style
func (ctx *TestContext) TheResponseStatusCodeShouldBe(expectedCode int) {
	if ctx.Response == nil {
		ctx.T.Errorf("Expected response to exist, but no response was received")
		return
	}
	if ctx.Response.StatusCode != expectedCode {
		ctx.T.Errorf("Expected status code to be %d, but got %d", expectedCode, ctx.Response.StatusCode)
	}
}

// TheJSONResponseShouldContainField checks a specific field in JSON response in BDD style
func (ctx *TestContext) TheJSONResponseShouldContainField(fieldPath string, expectedValue interface{}) {
	if ctx.Response == nil {
		ctx.T.Errorf("Expected response to exist, but no response was received")
		return
	}
	if len(ctx.ResponseBody) == 0 {
		ctx.T.Errorf("Expected response body to contain data, but it was empty")
		return
	}

	var jsonData map[string]interface{}
	err := json.Unmarshal(ctx.ResponseBody, &jsonData)
	if err != nil {
		ctx.T.Errorf("Expected response to be valid JSON, but got error: %v", err)
		return
	}

	value, exists := jsonData[fieldPath]
	if !exists {
		ctx.T.Errorf("Expected JSON response to contain field '%s', but it was not found", fieldPath)
		return
	}
	if value != expectedValue {
		ctx.T.Errorf("Expected field '%s' to be %v, but got %v", fieldPath, expectedValue, value)
	}
}

// TheJSONResponseShouldContain checks if JSON response contains a specific field in BDD style
func (ctx *TestContext) TheJSONResponseShouldContain(fieldPath string) {
	if ctx.Response == nil {
		ctx.T.Errorf("Expected response to exist, but no response was received")
		return
	}
	if len(ctx.ResponseBody) == 0 {
		ctx.T.Errorf("Expected response body to contain data, but it was empty")
		return
	}

	var jsonData map[string]interface{}
	err := json.Unmarshal(ctx.ResponseBody, &jsonData)
	if err != nil {
		ctx.T.Errorf("Expected response to be valid JSON, but got error: %v", err)
		return
	}

	_, exists := jsonData[fieldPath]
	if !exists {
		ctx.T.Errorf("Expected JSON response to contain field '%s', but it was not found", fieldPath)
	}
}

// TheResponseShouldNotBeEmpty checks that response body is not empty in BDD style
func (ctx *TestContext) TheResponseShouldNotBeEmpty() {
	if ctx.Response == nil {
		ctx.T.Errorf("Expected response to exist, but no response was received")
		return
	}
	if len(ctx.ResponseBody) == 0 {
		ctx.T.Errorf("Expected response body to contain data, but it was empty")
	}
}

// TheContentTypeShouldBe checks the Content-Type header in BDD style
func (ctx *TestContext) TheContentTypeShouldBe(expectedContentType string) {
	if ctx.Response == nil {
		ctx.T.Errorf("Expected response to exist, but no response was received")
		return
	}
	contentType := ctx.Response.Header.Get("Content-Type")
	if !strings.Contains(contentType, expectedContentType) {
		ctx.T.Errorf("Expected Content-Type to contain '%s', but got '%s'", expectedContentType, contentType)
	}
}

// NoErrorShouldHaveOccurred checks that no error occurred in BDD style
func (ctx *TestContext) NoErrorShouldHaveOccurred() {
	if ctx.Error != nil {
		ctx.T.Errorf("Expected no error to occur, but got: %v", ctx.Error)
	}
}

// AnErrorShouldHaveOccurred checks that an error occurred in BDD style
func (ctx *TestContext) AnErrorShouldHaveOccurred() {
	if ctx.Error == nil {
		ctx.T.Errorf("Expected an error to occur, but none was received")
	}
}

// Advanced BDD-style assertions for gRPC and Connect-Go testing

// TheGRPCResponseShouldBeSuccessful checks gRPC response success in BDD style
func (ctx *TestContext) TheGRPCResponseShouldBeSuccessful() {
	if ctx.GRPCResponse == nil {
		ctx.T.Errorf("Expected gRPC response to exist, but none was received")
		return
	}
	// gRPC specific validations can be added here
}

// TheConnectGoResponseShouldBeSuccessful checks Connect-Go response success in BDD style
func (ctx *TestContext) TheConnectGoResponseShouldBeSuccessful() {
	if ctx.ConnectGoResponse == nil {
		ctx.T.Errorf("Expected Connect-Go response to exist, but none was received")
		return
	}
	if ctx.ConnectGoResponse.Msg == nil {
		ctx.T.Errorf("Expected Connect-Go response message to exist, but it was nil")
		return
	}
}

// TheResponseHeaderShouldContain checks if response contains specific header in BDD style
func (ctx *TestContext) TheResponseHeaderShouldContain(headerName, expectedValue string) {
	if ctx.Response == nil {
		ctx.T.Errorf("Expected response to exist, but none was received")
		return
	}
	actualValue := ctx.Response.Header.Get(headerName)
	if actualValue == "" {
		ctx.T.Errorf("Expected response to contain header '%s', but it was not found", headerName)
		return
	}
	if !strings.Contains(actualValue, expectedValue) {
		ctx.T.Errorf("Expected header '%s' to contain '%s', but got '%s'", headerName, expectedValue, actualValue)
	}
}

// TheJSONResponseShouldHaveStructure checks JSON structure in BDD style
func (ctx *TestContext) TheJSONResponseShouldHaveStructure(requiredFields []string) {
	if ctx.Response == nil {
		ctx.T.Errorf("Expected response to exist, but none was received")
		return
	}
	if len(ctx.ResponseBody) == 0 {
		ctx.T.Errorf("Expected response body to contain data, but it was empty")
		return
	}

	var jsonData map[string]interface{}
	err := json.Unmarshal(ctx.ResponseBody, &jsonData)
	if err != nil {
		ctx.T.Errorf("Expected response to be valid JSON, but got error: %v", err)
		return
	}

	for _, field := range requiredFields {
		if _, exists := jsonData[field]; !exists {
			ctx.T.Errorf("Expected JSON response to contain required field '%s', but it was missing", field)
		}
	}
}

// AllConcurrentRequestsShouldSucceed checks all concurrent requests in BDD style
func (ctx *TestContext) AllConcurrentRequestsShouldSucceed() {
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
		}
	}

	if failedCount > 0 {
		ctx.T.Errorf("Expected all %d concurrent requests to succeed, but %d failed", len(ctx.ConcurrentResults), failedCount)
	}
}

// Legacy methods for backward compatibility (will be deprecated)
// AssertStatusCode checks the HTTP status code
func (ctx *TestContext) AssertStatusCode(expectedCode int) {
	require.NotNil(ctx.T, ctx.Response, "No response received")
	assert.Equal(ctx.T, expectedCode, ctx.Response.StatusCode,
		"Expected status code %d, got %d", expectedCode, ctx.Response.StatusCode)
}

// AssertJSONField checks a specific field in JSON response
func (ctx *TestContext) AssertJSONField(fieldPath string, expectedValue interface{}) {
	require.NotNil(ctx.T, ctx.Response, "No response received")
	require.NotEmpty(ctx.T, ctx.ResponseBody, "Response body is empty")

	var jsonData map[string]interface{}
	err := json.Unmarshal(ctx.ResponseBody, &jsonData)
	require.NoError(ctx.T, err, "Failed to parse JSON response")

	// Simple field path parsing (can be extended for nested fields)
	value, exists := jsonData[fieldPath]
	require.True(ctx.T, exists, "Field '%s' not found in JSON response", fieldPath)
	assert.Equal(ctx.T, expectedValue, value, "Expected field '%s' to be %v, got %v", fieldPath, expectedValue, value)
}

// AssertJSONContains checks if JSON response contains a specific field
func (ctx *TestContext) AssertJSONContains(fieldPath string) {
	require.NotNil(ctx.T, ctx.Response, "No response received")
	require.NotEmpty(ctx.T, ctx.ResponseBody, "Response body is empty")

	var jsonData map[string]interface{}
	err := json.Unmarshal(ctx.ResponseBody, &jsonData)
	require.NoError(ctx.T, err, "Failed to parse JSON response")

	_, exists := jsonData[fieldPath]
	assert.True(ctx.T, exists, "Field '%s' not found in JSON response", fieldPath)
}

// AssertNoError checks that no error occurred
func (ctx *TestContext) AssertNoError() {
	assert.NoError(ctx.T, ctx.Error, "Unexpected error occurred")
}

// AssertError checks that an error occurred
func (ctx *TestContext) AssertError() {
	assert.Error(ctx.T, ctx.Error, "Expected an error but none occurred")
}

// AssertResponseNotEmpty checks that response body is not empty
func (ctx *TestContext) AssertResponseNotEmpty() {
	require.NotNil(ctx.T, ctx.Response, "No response received")
	assert.NotEmpty(ctx.T, ctx.ResponseBody, "Response body should not be empty")
}

// AssertContentType checks the Content-Type header
func (ctx *TestContext) AssertContentType(expectedContentType string) {
	require.NotNil(ctx.T, ctx.Response, "No response received")
	contentType := ctx.Response.Header.Get("Content-Type")
	assert.Contains(ctx.T, contentType, expectedContentType,
		"Expected Content-Type to contain '%s', got '%s'", expectedContentType, contentType)
}
