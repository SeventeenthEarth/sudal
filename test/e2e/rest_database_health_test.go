package e2e

import (
	"testing"

	"github.com/seventeenthearth/sudal/test/e2e/steps"
)

// TestRESTDatabaseHealth tests the REST Database Health functionality
func TestRESTDatabaseHealth(t *testing.T) {
	// BDD Scenarios for REST Database Health
	scenarios := []steps.BDDScenario{
		{
			Name:        "Database health endpoint responds correctly",
			Description: "Should return 200 status with healthy database information",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeGETRequest(ctx, "/api/health/database")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 200)
				steps.ThenJSONResponseShouldContainStatus(ctx, "healthy")
				steps.ThenJSONResponseShouldContainField(ctx, "database", "connected")
			},
		},
		{
			Name:        "Database health endpoint basic response",
			Description: "Should return 200 status with basic database health information",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeGETRequest(ctx, "/api/health/database")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 200)
				steps.ThenJSONResponseShouldContainStatus(ctx, "healthy")
			},
		},
		{
			Name:        "Database connection status is healthy",
			Description: "Should return healthy database connection status",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeGETRequest(ctx, "/api/health/database")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 200)
				steps.ThenJSONResponseShouldContainStatus(ctx, "healthy")
			},
		},
		{
			Name:        "Database health provides basic response",
			Description: "Should return basic database health response",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeGETRequest(ctx, "/api/health/database")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 200)
				steps.ThenJSONResponseShouldContainStatus(ctx, "healthy")
			},
		},
		{
			Name:        "Database health endpoint performance",
			Description: "Should handle concurrent requests successfully",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeConcurrentRequests(ctx, 5, "/api/health/database")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenAllDatabaseHealthRequestsShouldSucceed(ctx)
				steps.ThenAllResponsesShouldContainValidConnectionStatistics(ctx)
			},
		},
	}

	// Run all BDD scenarios
	steps.RunBDDScenarios(t, ServerURL, scenarios)
}

// TestRESTDatabaseHealthTableDriven demonstrates table-driven BDD tests for database health
func TestRESTDatabaseHealthTableDriven(t *testing.T) {
	// Table-driven test cases for different database health scenarios
	type DatabaseHealthTestCase struct {
		Name                string
		Endpoint            string
		ExpectedStatus      int
		ShouldHaveDatabase  bool
		ShouldHaveStats     bool
		ShouldHaveTimestamp bool
	}

	testCases := []interface{}{
		DatabaseHealthTestCase{
			Name:                "Standard database health check",
			Endpoint:            "/api/health/database",
			ExpectedStatus:      200,
			ShouldHaveDatabase:  true,
			ShouldHaveStats:     true,
			ShouldHaveTimestamp: true,
		},
		DatabaseHealthTestCase{
			Name:                "Database health with query parameters",
			Endpoint:            "/api/health/database?detailed=true",
			ExpectedStatus:      200,
			ShouldHaveDatabase:  true,
			ShouldHaveStats:     true,
			ShouldHaveTimestamp: true,
		},
	}

	tableDrivenTest := steps.TableDrivenBDDTest{
		Name: "Database health scenarios",
		Given: func(ctx *steps.TestContext, testData interface{}) {
			steps.GivenServerIsRunning(ctx)
		},
		When: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(DatabaseHealthTestCase)
			steps.WhenIMakeGETRequest(ctx, testCase.Endpoint)
		},
		Then: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(DatabaseHealthTestCase)
			steps.ThenHTTPStatusShouldBe(ctx, testCase.ExpectedStatus)

			if testCase.ShouldHaveDatabase {
				steps.ThenJSONResponseShouldContainDatabaseInformation(ctx)
			}

			if testCase.ShouldHaveStats {
				steps.ThenJSONResponseShouldContainConnectionStatistics(ctx)
			}

			if testCase.ShouldHaveTimestamp {
				steps.ThenJSONResponseShouldContainTimestampField(ctx)
			}
		},
	}

	steps.RunTableDrivenBDDTest(t, ServerURL, tableDrivenTest, testCases)
}

// TestRESTDatabaseHealthConcurrency tests concurrent database health requests
func TestRESTDatabaseHealthConcurrency(t *testing.T) {
	// Table-driven test for different concurrency levels
	type ConcurrencyTestCase struct {
		Name        string
		NumRequests int
		Endpoint    string
		Description string
	}

	concurrencyTestCases := []interface{}{
		ConcurrencyTestCase{
			Name:        "Low concurrency database health",
			NumRequests: 3,
			Endpoint:    "/api/health/database",
			Description: "Test database health with 3 concurrent requests",
		},
		ConcurrencyTestCase{
			Name:        "Medium concurrency database health",
			NumRequests: 5,
			Endpoint:    "/api/health/database",
			Description: "Test database health with 5 concurrent requests",
		},
		ConcurrencyTestCase{
			Name:        "High concurrency database health",
			NumRequests: 10,
			Endpoint:    "/api/health/database",
			Description: "Test database health with 10 concurrent requests",
		},
	}

	concurrencyTest := steps.TableDrivenBDDTest{
		Name: "Database health concurrency scenarios",
		Given: func(ctx *steps.TestContext, testData interface{}) {
			steps.GivenServerIsRunning(ctx)
		},
		When: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(ConcurrencyTestCase)
			steps.WhenIMakeConcurrentRequests(ctx, testCase.NumRequests, testCase.Endpoint)
		},
		Then: func(ctx *steps.TestContext, testData interface{}) {
			steps.ThenAllDatabaseHealthRequestsShouldSucceed(ctx)
			steps.ThenAllResponsesShouldContainValidConnectionStatistics(ctx)
		},
	}

	steps.RunTableDrivenBDDTest(t, ServerURL, concurrencyTest, concurrencyTestCases)
}

// TestRESTDatabaseHealthValidation tests database health response validation
func TestRESTDatabaseHealthValidation(t *testing.T) {
	// BDD Scenarios for database health validation
	validationScenarios := []steps.BDDScenario{
		{
			Name:        "Connection statistics consistency validation",
			Description: "Should validate that connection statistics are mathematically consistent",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeGETRequest(ctx, "/api/health/database")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 200)
				steps.ThenConnectionStatisticsShouldBeValid(ctx)
			},
		},
		{
			Name:        "Max open connections validation",
			Description: "Should validate that max_open_connections is properly configured",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeGETRequest(ctx, "/api/health/database")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 200)
				steps.ThenConnectionStatisticsShouldIncludeMaxOpenConnections(ctx)
			},
		},
		{
			Name:        "Current usage metrics validation",
			Description: "Should validate that current usage metrics are present and valid",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeGETRequest(ctx, "/api/health/database")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 200)
				steps.ThenConnectionStatisticsShouldIncludeCurrentUsageMetrics(ctx)
			},
		},
	}

	// Run validation scenarios
	steps.RunBDDScenarios(t, ServerURL, validationScenarios)
}
