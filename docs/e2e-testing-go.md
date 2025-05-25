# Go BDD Style E2E Testing

This document describes the Go-based End-to-End (E2E) testing framework that replaces the previous Python pytest-bdd implementation.

## Overview

The Go E2E testing framework provides a BDD (Behavior Driven Development) style testing approach using Go's standard testing package and testify assertions. It maintains the Given-When-Then structure while leveraging Go's type safety and performance benefits.

## Architecture

### Directory Structure

```
test/e2e/
├── steps/                          # Reusable step functions
│   ├── bdd_helpers.go              # BDD framework and test context
│   ├── common_steps.go             # Common Given/When/Then steps
│   ├── connect_steps.go            # Connect-Go specific steps
│   └── database_steps.go           # Database health specific steps
├── connect_health_service_test.go  # Connect-Go health service tests
├── rest_database_health_test.go    # REST database health tests
└── rest_monitoring_test.go         # REST monitoring tests
```

### Core Components

#### 1. BDD Helpers (`steps/bdd_helpers.go`)

- **TestContext**: Holds test state including HTTP client, responses, and errors
- **BDDScenario**: Represents a Given-When-Then scenario
- **TableDrivenBDDTest**: Support for parameterized test scenarios
- **Helper Methods**: Common assertions and utilities

#### 2. Step Functions

Step functions are organized by feature area:

- **Common Steps** (`common_steps.go`): HTTP requests, status checks, JSON validation
- **Connect Steps** (`connect_steps.go`): Connect-Go protocol specific operations
- **Database Steps** (`database_steps.go`): Database health and connection statistics

## Writing BDD Tests

### Basic BDD Scenario

```go
func TestExample(t *testing.T) {
    scenario := steps.BDDScenario{
        Name: "Health check responds correctly",
        Given: func(ctx *steps.TestContext) {
            steps.GivenServerIsRunning(ctx)
        },
        When: func(ctx *steps.TestContext) {
            steps.WhenIMakeGETRequest(ctx, "/healthz")
        },
        Then: func(ctx *steps.TestContext) {
            steps.ThenHTTPStatusShouldBe(ctx, 200)
            steps.ThenJSONResponseShouldContainStatus(ctx, "healthy")
        },
    }
    
    steps.RunBDDScenario(t, serverURL, scenario)
}
```

### Table-Driven BDD Tests

```go
func TestTableDriven(t *testing.T) {
    type TestCase struct {
        Endpoint       string
        ExpectedStatus int
        ExpectedValue  string
    }

    testCases := []interface{}{
        TestCase{"/ping", 200, "ok"},
        TestCase{"/healthz", 200, "healthy"},
    }

    test := steps.TableDrivenBDDTest{
        Name: "Monitoring endpoints",
        Given: func(ctx *steps.TestContext, testData interface{}) {
            steps.GivenServerIsRunning(ctx)
        },
        When: func(ctx *steps.TestContext, testData interface{}) {
            testCase := testData.(TestCase)
            steps.WhenIMakeGETRequest(ctx, testCase.Endpoint)
        },
        Then: func(ctx *steps.TestContext, testData interface{}) {
            testCase := testData.(TestCase)
            steps.ThenHTTPStatusShouldBe(ctx, testCase.ExpectedStatus)
            steps.ThenJSONResponseShouldContainStatus(ctx, testCase.ExpectedValue)
        },
    }

    steps.RunTableDrivenBDDTest(t, serverURL, test, testCases)
}
```

### Multiple Scenarios

```go
func TestMultipleScenarios(t *testing.T) {
    scenarios := []steps.BDDScenario{
        {
            Name: "Scenario 1",
            Given: func(ctx *steps.TestContext) { /* ... */ },
            When:  func(ctx *steps.TestContext) { /* ... */ },
            Then:  func(ctx *steps.TestContext) { /* ... */ },
        },
        {
            Name: "Scenario 2",
            Given: func(ctx *steps.TestContext) { /* ... */ },
            When:  func(ctx *steps.TestContext) { /* ... */ },
            Then:  func(ctx *steps.TestContext) { /* ... */ },
        },
    }

    steps.RunBDDScenarios(t, serverURL, scenarios)
}
```

## Available Step Functions

### Common Steps

#### Given Steps
- `GivenServerIsRunning(ctx)`: Ensures server is accessible

#### When Steps
- `WhenIMakeGETRequest(ctx, endpoint)`: Makes GET request
- `WhenIMakePOSTRequest(ctx, endpoint, contentType, body)`: Makes POST request
- `WhenIMakeConcurrentRequests(ctx, numRequests, endpoint)`: Makes concurrent requests

#### Then Steps
- `ThenHTTPStatusShouldBe(ctx, expectedStatus)`: Checks HTTP status
- `ThenJSONResponseShouldContainStatus(ctx, expectedStatus)`: Checks JSON status field
- `ThenResponseShouldNotBeEmpty(ctx)`: Validates response is not empty
- `ThenContentTypeShouldBe(ctx, expectedContentType)`: Checks Content-Type header

### Connect-Go Steps

#### When Steps
- `WhenIMakeHealthCheckRequestUsingConnectGo(ctx)`: Connect-Go health check
- `WhenIMakeHealthCheckRequestWithInvalidContentType(ctx)`: Invalid content type test
- `WhenIMakeConcurrentHealthCheckRequests(ctx, numRequests)`: Concurrent Connect-Go requests

#### Then Steps
- `ThenResponseShouldIndicateServingStatus(ctx)`: Checks SERVING status
- `ThenResponseShouldContainProperConnectGoHeaders(ctx)`: Validates headers

### Database Steps

#### Then Steps
- `ThenJSONResponseShouldContainDatabaseInformation(ctx)`: Database info validation
- `ThenJSONResponseShouldContainConnectionStatistics(ctx)`: Connection stats validation
- `ThenDatabaseConnectionPoolShouldBeHealthy(ctx)`: Pool health validation
- `ThenConnectionStatisticsShouldBeValid(ctx)`: Statistics consistency validation

## Running Tests

### Command Line

```bash
# Run all E2E tests
make test.e2e

# Run only E2E tests (no preparation)
make test.e2e.go.only

# Run directly with go test
go test -v ./test/e2e

# Run specific test
go test -v ./test/e2e -run TestConnectGoHealthService
```

### Prerequisites

1. **Server Running**: The server must be running on `localhost:8080`
   ```bash
   make run
   ```

2. **Dependencies**: Ensure testify is installed
   ```bash
   go mod tidy
   ```

## Test Features

### Concurrent Testing
Built-in support for testing concurrent requests:

```go
steps.WhenIMakeConcurrentRequests(ctx, 10, "/healthz")
steps.ThenAllRequestsShouldSucceed(ctx)
```

### Error Handling
Comprehensive error scenario testing:

```go
steps.WhenIMakeHealthCheckRequestWithInvalidContentType(ctx)
steps.ThenHTTPStatusShouldBe(ctx, 415)
```

### Assertions
Rich assertion library through testify:

```go
ctx.AssertStatusCode(200)
ctx.AssertJSONField("status", "healthy")
ctx.AssertContentType("application/json")
```

## Migration from Python

The Go BDD framework maintains the same test scenarios as the previous Python implementation:

- **Feature Parity**: All Python test scenarios have been converted
- **BDD Structure**: Maintains Given-When-Then organization
- **Test Coverage**: Same test coverage with improved performance
- **Type Safety**: Compile-time validation of test code

## Best Practices

1. **Organize Steps**: Group related steps in appropriate files
2. **Reuse Steps**: Use existing step functions when possible
3. **Clear Naming**: Use descriptive scenario names
4. **Error Handling**: Always check for errors in step functions
5. **Concurrent Safety**: Ensure step functions are thread-safe for concurrent tests
6. **Test Isolation**: Each test should be independent and not rely on other tests

## Extending the Framework

### Adding New Steps

1. Create step functions in appropriate `steps/*.go` files
2. Follow the naming convention: `Given*/When*/Then*`
3. Use the TestContext for state management
4. Add proper error handling and assertions

### Adding New Test Files

1. Create test files in `test/e2e/` directory
2. Import the steps package
3. Use the BDD framework helpers
4. Follow the established patterns for consistency
