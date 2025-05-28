package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

// Helper methods for common assertions

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
