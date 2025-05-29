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
			Name:        "Health check should return serving status",
			Description: "When I make a health check request, then the response should indicate serving status",
			Given: func(ctx *steps.TestContext) {
				// Given the server is running
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				// When I make a health check request using Connect-Go
				steps.WhenIMakeHealthCheckRequestUsingConnectGo(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				// Then the response status code should be 200
				ctx.TheResponseStatusCodeShouldBe(200)
				// And the response should not be empty
				ctx.TheResponseShouldNotBeEmpty()
				// And the JSON response should contain the serving status
				ctx.TheJSONResponseShouldContainField("status", "SERVING_STATUS_SERVING")
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
				steps.WhenIMakeGETRequest(ctx, "/health/database")
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
			Name:        "Invalid content type should be rejected",
			Description: "When I send a request with invalid content type, then the server should reject it",
			Given: func(ctx *steps.TestContext) {
				// Given the server is running
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				// When I make a health check request with invalid content type
				steps.WhenIMakeHealthCheckRequestWithInvalidContentType(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				// Then the response status code should be 415 (Unsupported Media Type)
				ctx.TheResponseStatusCodeShouldBe(415)
			},
		},
		{
			Name:        "Concurrent requests should all succeed",
			Description: "When I make multiple concurrent requests, then all should succeed",
			Given: func(ctx *steps.TestContext) {
				// Given the server is running
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				// When I make 5 concurrent health check requests
				steps.WhenIMakeConcurrentHealthCheckRequests(ctx, 5)
			},
			Then: func(ctx *steps.TestContext) {
				// Then all requests should succeed
				ctx.AllConcurrentRequestsShouldSucceed()
				// And all responses should indicate serving status
				steps.ThenAllResponsesShouldIndicateServingStatus(ctx)
			},
		},
	}

	// Run all BDD scenarios
	steps.RunBDDScenarios(t, serverURL, scenarios)
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
			Endpoint:       "/ping",
			ExpectedStatus: 200,
			ShouldHaveJSON: false,
		},
		EndpointTestCase{
			Name:           "Database health should return detailed info",
			Endpoint:       "/health/database",
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

	steps.RunTableDrivenBDDTest(t, serverURL, tableDrivenTest, testCases)
}
