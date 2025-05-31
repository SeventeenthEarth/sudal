package e2e

import (
	"testing"

	"github.com/seventeenthearth/sudal/test/e2e/steps"
)

// TestBDDStyleExample demonstrates pure BDD style testing without testify assertions
func TestBDDStyleExample(t *testing.T) {
	// BDD Scenarios using pure BDD style assertions
	scenarios := []steps.BDDScenario{
		{
			Name:        "gRPC-only endpoint should reject HTTP/JSON requests",
			Description: "When I make an HTTP/JSON request to gRPC-only endpoint, then it should return 404",
			Given: func(ctx *steps.TestContext) {
				// Given the server is running
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				// When I make an HTTP/JSON request to gRPC-only endpoint
				steps.WhenIMakeHealthCheckRequestUsingConnectGo(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				// Then the response status code should be 404 (gRPC-only endpoint)
				ctx.TheResponseStatusCodeShouldBe(404)
			},
		},
		{
			Name:        "Database health endpoint should provide connection information",
			Description: "When I request database health, then I should receive detailed connection statistics",
			Given: func(ctx *steps.TestContext) {
				// Given the server is running
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				// When I make a GET request to the database health endpoint
				steps.WhenIMakeGETRequest(ctx, "/api/health/database")
			},
			Then: func(ctx *steps.TestContext) {
				// Then the response status code should be 200
				ctx.TheResponseStatusCodeShouldBe(200)
				// And the JSON response should contain status field
				ctx.TheJSONResponseShouldContain("status")
				// And the JSON response should contain database information
				ctx.TheJSONResponseShouldContain("database")
				// And the content type should be JSON
				ctx.TheContentTypeShouldBe("application/json")
			},
		},
		{
			Name:        "gRPC-only endpoint should reject invalid content type",
			Description: "When I send a request with invalid content type to gRPC endpoint, then it should return 404",
			Given: func(ctx *steps.TestContext) {
				// Given the server is running
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				// When I make a health check request with invalid content type
				steps.WhenIMakeHealthCheckRequestWithInvalidContentType(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				// Then the response status code should be 404 (gRPC-only endpoint blocks HTTP requests)
				ctx.TheResponseStatusCodeShouldBe(404)
			},
		},
		{
			Name:        "Concurrent HTTP requests to gRPC endpoint should be rejected",
			Description: "When I make multiple concurrent HTTP requests to gRPC endpoint, then all should return 404",
			Given: func(ctx *steps.TestContext) {
				// Given the server is running
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				// When I make 5 concurrent health check requests (HTTP/JSON to gRPC endpoint)
				steps.WhenIMakeConcurrentHealthCheckRequests(ctx, 5)
			},
			Then: func(ctx *steps.TestContext) {
				// Then all requests should return 404 (gRPC-only endpoint)
				steps.ThenAllConcurrentRequestsShouldReturn404(ctx)
			},
		},
	}

	// Run all BDD scenarios
	steps.RunBDDScenarios(t, ServerURL, scenarios)
}

// TestBDDStyleTableDriven demonstrates table-driven BDD tests with pure BDD style
func TestBDDStyleTableDriven(t *testing.T) {
	// Table-driven test cases for different endpoint scenarios
	type EndpointTestCase struct {
		Name           string
		Endpoint       string
		ExpectedStatus int
		ShouldHaveJSON bool
	}

	testCases := []interface{}{
		EndpointTestCase{
			Name:           "Ping endpoint should return OK",
			Endpoint:       "/api/ping",
			ExpectedStatus: 200,
			ShouldHaveJSON: false,
		},
		EndpointTestCase{
			Name:           "Database health should return detailed info",
			Endpoint:       "/api/health/database",
			ExpectedStatus: 200,
			ShouldHaveJSON: true,
		},
		EndpointTestCase{
			Name:           "Non-existent endpoint should return 404",
			Endpoint:       "/non-existent",
			ExpectedStatus: 404,
			ShouldHaveJSON: false,
		},
	}

	tableDrivenTest := steps.TableDrivenBDDTest{
		Name: "Endpoint response scenarios",
		Given: func(ctx *steps.TestContext, testData interface{}) {
			// Given the server is running
			steps.GivenServerIsRunning(ctx)
		},
		When: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(EndpointTestCase)
			// When I make a GET request to the endpoint
			steps.WhenIMakeGETRequest(ctx, testCase.Endpoint)
		},
		Then: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(EndpointTestCase)
			// Then the response status code should match expected
			ctx.TheResponseStatusCodeShouldBe(testCase.ExpectedStatus)

			if testCase.ShouldHaveJSON {
				// And the content type should be JSON
				ctx.TheContentTypeShouldBe("application/json")
				// And the response should not be empty
				ctx.TheResponseShouldNotBeEmpty()
			}
		},
	}

	steps.RunTableDrivenBDDTest(t, ServerURL, tableDrivenTest, testCases)
}
