package e2e

import (
	"testing"

	"github.com/seventeenthearth/sudal/test/e2e/steps"
)

const serverURL = "http://localhost:8080"

// TestConnectGoHealthService tests the Connect-Go Health Service functionality
func TestConnectGoHealthService(t *testing.T) {
	// BDD Scenarios for Connect-Go Health Service
	scenarios := []steps.BDDScenario{
		{
			Name:        "HTTP/JSON request to gRPC endpoint should be rejected",
			Description: "Should return 404 status when making HTTP/JSON request to gRPC-only endpoint",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeHealthCheckRequestUsingConnectGo(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 404)
			},
		},
		{
			Name:        "HTTP/JSON requests to gRPC endpoint should be rejected",
			Description: "Should return 404 status when making HTTP/JSON request to gRPC-only endpoint",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeHealthCheckRequestUsingHTTPJSON(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 404)
			},
		},
		{
			Name:        "Invalid content type requests to gRPC endpoint should be rejected",
			Description: "Should return 404 status when making request with invalid content type to gRPC-only endpoint",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeHealthCheckRequestWithInvalidContentType(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 404)
			},
		},
		{
			Name:        "Non-existent Connect-Go method returns 404",
			Description: "Should return 404 status when making request to non-existent endpoint",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeRequestToNonExistentEndpoint(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 404)
			},
		},
		{
			Name:        "Multiple concurrent HTTP/JSON requests to gRPC endpoint should be rejected",
			Description: "Should return 404 for all concurrent HTTP/JSON requests to gRPC-only endpoint",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeConcurrentHealthCheckRequests(ctx, 10)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenAllConcurrentRequestsShouldReturn404(ctx)
			},
		},
		{
			Name:        "HTTP/JSON request to gRPC endpoint should be rejected with proper error",
			Description: "Should return 404 with proper error message for HTTP/JSON requests to gRPC-only endpoint",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeHealthCheckRequestUsingConnectGo(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 404)
			},
		},
	}

	// Run all BDD scenarios
	steps.RunBDDScenarios(t, serverURL, scenarios)
}

// TestConnectGoHealthServiceTableDriven demonstrates table-driven BDD tests
func TestConnectGoHealthServiceTableDriven(t *testing.T) {
	// Table-driven test cases for different request scenarios
	type RequestTestCase struct {
		Name                string
		Endpoint            string
		ContentType         string
		Body                string
		ExpectedStatus      int
		ShouldContainStatus bool
		ExpectedStatusValue string
	}

	testCases := []interface{}{
		RequestTestCase{
			Name:                "HTTP/JSON request to gRPC endpoint should be rejected",
			Endpoint:            "/health.v1.HealthService/Check",
			ContentType:         "application/json",
			Body:                "{}",
			ExpectedStatus:      404,
			ShouldContainStatus: false,
			ExpectedStatusValue: "",
		},
		RequestTestCase{
			Name:                "Invalid content type to gRPC endpoint should be rejected",
			Endpoint:            "/health.v1.HealthService/Check",
			ContentType:         "text/plain",
			Body:                "{}",
			ExpectedStatus:      404,
			ShouldContainStatus: false,
		},
		RequestTestCase{
			Name:                "Non-existent method",
			Endpoint:            "/health.v1.HealthService/NonExistentMethod",
			ContentType:         "application/json",
			Body:                "{}",
			ExpectedStatus:      404,
			ShouldContainStatus: false,
		},
	}

	tableDrivenTest := steps.TableDrivenBDDTest{
		Name: "Connect-Go request scenarios",
		Given: func(ctx *steps.TestContext, testData interface{}) {
			steps.GivenServerIsRunning(ctx)
		},
		When: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(RequestTestCase)
			steps.WhenIMakePOSTRequest(ctx, testCase.Endpoint, testCase.ContentType, testCase.Body)
		},
		Then: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(RequestTestCase)
			steps.ThenHTTPStatusShouldBe(ctx, testCase.ExpectedStatus)

			if testCase.ShouldContainStatus {
				steps.ThenJSONResponseShouldContainStatus(ctx, testCase.ExpectedStatusValue)
			}
		},
	}

	steps.RunTableDrivenBDDTest(t, serverURL, tableDrivenTest, testCases)
}

// TestConnectGoHealthServiceConcurrency tests concurrent request scenarios
func TestConnectGoHealthServiceConcurrency(t *testing.T) {
	// Table-driven test for different concurrency levels
	type ConcurrencyTestCase struct {
		Name        string
		NumRequests int
		Description string
	}

	concurrencyTestCases := []interface{}{
		ConcurrencyTestCase{
			Name:        "Low concurrency",
			NumRequests: 5,
			Description: "Test with 5 concurrent requests",
		},
		ConcurrencyTestCase{
			Name:        "Medium concurrency",
			NumRequests: 10,
			Description: "Test with 10 concurrent requests",
		},
		ConcurrencyTestCase{
			Name:        "High concurrency",
			NumRequests: 20,
			Description: "Test with 20 concurrent requests",
		},
	}

	concurrencyTest := steps.TableDrivenBDDTest{
		Name: "Connect-Go concurrency scenarios",
		Given: func(ctx *steps.TestContext, testData interface{}) {
			steps.GivenServerIsRunning(ctx)
		},
		When: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(ConcurrencyTestCase)
			steps.WhenIMakeConcurrentHealthCheckRequests(ctx, testCase.NumRequests)
		},
		Then: func(ctx *steps.TestContext, testData interface{}) {
			steps.ThenAllConcurrentRequestsShouldReturn404(ctx)
		},
	}

	steps.RunTableDrivenBDDTest(t, serverURL, concurrencyTest, concurrencyTestCases)
}
