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
			Description: "Should return 200 status with healthy database information and connection statistics",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeGETRequest(ctx, "/health/database")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 200)
				steps.ThenJSONResponseShouldContainStatus(ctx, "healthy")
				steps.ThenJSONResponseShouldContainDatabaseInformation(ctx)
				steps.ThenJSONResponseShouldContainConnectionStatistics(ctx)
			},
		},
		{
			Name:        "Database health endpoint includes timestamp",
			Description: "Should return 200 status with a valid timestamp field",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeGETRequest(ctx, "/health/database")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 200)
				steps.ThenJSONResponseShouldContainTimestampField(ctx)
			},
		},
		{
			Name:        "Database connection pool status is healthy",
			Description: "Should return healthy connection pool with valid statistics",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeGETRequest(ctx, "/health/database")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 200)
				steps.ThenDatabaseConnectionPoolShouldBeHealthy(ctx)
				steps.ThenConnectionStatisticsShouldBeValid(ctx)
			},
		},
		{
			Name:        "Database health provides detailed connection metrics",
			Description: "Should return detailed connection metrics including max connections and usage",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeGETRequest(ctx, "/health/database")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 200)
				steps.ThenJSONResponseShouldContainConnectionStatistics(ctx)
				steps.ThenConnectionStatisticsShouldIncludeMaxOpenConnections(ctx)
				steps.ThenConnectionStatisticsShouldIncludeCurrentUsageMetrics(ctx)
			},
		},
		{
			Name:        "Database health endpoint performance",
			Description: "Should handle concurrent requests successfully",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				steps.WhenIMakeConcurrentRequests(ctx, 5, "/health/database")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenAllDatabaseHealthRequestsShouldSucceed(ctx)
				steps.ThenAllResponsesShouldContainValidConnectionStatistics(ctx)
			},
		},
	}

	// Run all BDD scenarios
	steps.RunBDDScenarios(t, serverURL, scenarios)
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
			Endpoint:            "/health/database",
			ExpectedStatus:      200,
			ShouldHaveDatabase:  true,
			ShouldHaveStats:     true,
			ShouldHaveTimestamp: true,
		},
		DatabaseHealthTestCase{
			Name:                "Database health with query parameters",
			Endpoint:            "/health/database?detailed=true",
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

	steps.RunTableDrivenBDDTest(t, serverURL, tableDrivenTest, testCases)
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
			Endpoint:    "/health/database",
			Description: "Test database health with 3 concurrent requests",
		},
		ConcurrencyTestCase{
			Name:        "Medium concurrency database health",
			NumRequests: 5,
			Endpoint:    "/health/database",
			Description: "Test database health with 5 concurrent requests",
		},
		ConcurrencyTestCase{
			Name:        "High concurrency database health",
			NumRequests: 10,
			Endpoint:    "/health/database",
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

	steps.RunTableDrivenBDDTest(t, serverURL, concurrencyTest, concurrencyTestCases)
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
				steps.WhenIMakeGETRequest(ctx, "/health/database")
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
				steps.WhenIMakeGETRequest(ctx, "/health/database")
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
				steps.WhenIMakeGETRequest(ctx, "/health/database")
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 200)
				steps.ThenConnectionStatisticsShouldIncludeCurrentUsageMetrics(ctx)
			},
		},
	}

	// Run validation scenarios
	steps.RunBDDScenarios(t, serverURL, validationScenarios)
}
