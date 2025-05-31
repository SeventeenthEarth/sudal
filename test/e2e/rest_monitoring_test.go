package e2e

import (
	"testing"

	"github.com/seventeenthearth/sudal/test/e2e/steps"
)

// TestRESTMonitoring tests the REST Monitoring functionality
func TestRESTMonitoring(t *testing.T) {
	// BDD Scenarios for REST Monitoring
	scenarios := []steps.BDDScenario{
		{
			Name:        "Server ping endpoint responds correctly",
			Description: "Should return 200 status with 'ok' status when pinging server",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeGETRequest(ctx, "/api/ping")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 200)
				steps.ThenJSONResponseShouldContainStatus(ctx, "ok")
			},
		},
		{
			Name:        "Basic health endpoint responds correctly",
			Description: "Should return 200 status with 'healthy' status",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeGETRequest(ctx, "/api/healthz")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 200)
				steps.ThenJSONResponseShouldContainStatus(ctx, "healthy")
			},
		},
		{
			Name:        "Health endpoint provides simple status",
			Description: "Should return lightweight response suitable for monitoring",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeGETRequest(ctx, "/api/healthz")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 200)
				steps.ThenJSONResponseShouldContainStatus(ctx, "healthy")
				steps.ThenResponseShouldBeLightweightForMonitoring(ctx)
			},
		},
	}

	// Run all BDD scenarios
	steps.RunBDDScenarios(t, ServerURL, scenarios)
}

// TestRESTMonitoringMultipleEndpoints tests multiple monitoring endpoints accessibility
func TestRESTMonitoringMultipleEndpoints(t *testing.T) {
	t.Run("Multiple monitoring endpoints are accessible", func(t *testing.T) {
		ctx := steps.NewTestContext(t, ServerURL)

		// Given
		steps.GivenServerIsRunning(ctx)

		// When - Test ping endpoint
		steps.WhenIMakeGETRequest(ctx, "/api/ping")

		// Then
		steps.ThenHTTPStatusShouldBe(ctx, 200)

		// When - Test health endpoint
		steps.WhenIMakeGETRequest(ctx, "/api/healthz")

		// Then
		steps.ThenHTTPStatusShouldBe(ctx, 200)
	})
}

// TestRESTMonitoringTableDriven demonstrates table-driven BDD tests for monitoring endpoints
func TestRESTMonitoringTableDriven(t *testing.T) {
	// Table-driven test cases for different monitoring endpoints
	type MonitoringTestCase struct {
		Name                string
		Endpoint            string
		ExpectedStatus      int
		ExpectedValue       string
		ShouldBeLightweight bool
	}

	testCases := []interface{}{
		MonitoringTestCase{
			Name:                "Ping endpoint",
			Endpoint:            "/api/ping",
			ExpectedStatus:      200,
			ExpectedValue:       "ok",
			ShouldBeLightweight: true,
		},
		MonitoringTestCase{
			Name:                "Health endpoint",
			Endpoint:            "/api/healthz",
			ExpectedStatus:      200,
			ExpectedValue:       "healthy",
			ShouldBeLightweight: true,
		},
	}

	tableDrivenTest := steps.TableDrivenBDDTest{
		Name: "Monitoring endpoint scenarios",
		Given: func(ctx *steps.TestContext, testData interface{}) {
			steps.GivenServerIsRunning(ctx)
		},
		When: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(MonitoringTestCase)
			steps.WhenIMakeGETRequest(ctx, testCase.Endpoint)
		},
		Then: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(MonitoringTestCase)
			steps.ThenHTTPStatusShouldBe(ctx, testCase.ExpectedStatus)
			steps.ThenJSONResponseShouldContainStatus(ctx, testCase.ExpectedValue)

			if testCase.ShouldBeLightweight {
				steps.ThenResponseShouldBeLightweightForMonitoring(ctx)
			}
		},
	}

	steps.RunTableDrivenBDDTest(t, ServerURL, tableDrivenTest, testCases)
}

// TestRESTMonitoringConcurrency tests concurrent monitoring requests
func TestRESTMonitoringConcurrency(t *testing.T) {
	// Table-driven test for different concurrency scenarios
	type ConcurrencyTestCase struct {
		Name           string
		NumRequests    int
		Endpoint       string
		ExpectedStatus string
		Description    string
	}

	concurrencyTestCases := []interface{}{
		ConcurrencyTestCase{
			Name:           "Concurrent ping requests",
			NumRequests:    5,
			Endpoint:       "/api/ping",
			ExpectedStatus: "ok",
			Description:    "Test ping endpoint with 5 concurrent requests",
		},
		ConcurrencyTestCase{
			Name:           "Concurrent health requests",
			NumRequests:    5,
			Endpoint:       "/api/healthz",
			ExpectedStatus: "healthy",
			Description:    "Test health endpoint with 5 concurrent requests",
		},
		ConcurrencyTestCase{
			Name:           "High concurrency ping requests",
			NumRequests:    10,
			Endpoint:       "/api/ping",
			ExpectedStatus: "ok",
			Description:    "Test ping endpoint with 10 concurrent requests",
		},
		ConcurrencyTestCase{
			Name:           "High concurrency health requests",
			NumRequests:    10,
			Endpoint:       "/api/healthz",
			ExpectedStatus: "healthy",
			Description:    "Test health endpoint with 10 concurrent requests",
		},
	}

	concurrencyTest := steps.TableDrivenBDDTest{
		Name: "Monitoring concurrency scenarios",
		Given: func(ctx *steps.TestContext, testData interface{}) {
			steps.GivenServerIsRunning(ctx)
		},
		When: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(ConcurrencyTestCase)
			steps.WhenIMakeConcurrentRequests(ctx, testCase.NumRequests, testCase.Endpoint)
		},
		Then: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(ConcurrencyTestCase)
			steps.ThenAllRequestsShouldSucceed(ctx)
			steps.ThenAllResponsesShouldContainStatus(ctx, testCase.ExpectedStatus)
		},
	}

	steps.RunTableDrivenBDDTest(t, ServerURL, concurrencyTest, concurrencyTestCases)
}

// TestRESTMonitoringPerformance tests monitoring endpoint performance characteristics
func TestRESTMonitoringPerformance(t *testing.T) {
	// BDD Scenarios for monitoring performance
	performanceScenarios := []steps.BDDScenario{
		{
			Name:        "Ping endpoint performance",
			Description: "Should respond quickly and be lightweight",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeGETRequest(ctx, "/api/ping")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 200)
				steps.ThenResponseShouldBeLightweightForMonitoring(ctx)
				steps.ThenContentTypeShouldBe(ctx, "application/json")
			},
		},
		{
			Name:        "Health endpoint performance",
			Description: "Should respond quickly and be lightweight",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeGETRequest(ctx, "/api/healthz")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 200)
				steps.ThenResponseShouldBeLightweightForMonitoring(ctx)
				steps.ThenContentTypeShouldBe(ctx, "application/json")
			},
		},
	}

	// Run performance scenarios
	steps.RunBDDScenarios(t, ServerURL, performanceScenarios)
}
