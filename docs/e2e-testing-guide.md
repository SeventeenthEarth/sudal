# E2E Testing Guide

This comprehensive guide covers End-to-End (E2E) testing for the Social Quiz Platform backend, implemented using Go with a custom BDD framework optimized for gRPC testing.

## Overview

The E2E testing framework provides a pure BDD (Behavior Driven Development) style testing approach that maintains the Given-When-Then structure while leveraging Go's type safety and performance benefits. Tests verify the complete functionality of the service by making actual HTTP/gRPC requests to a running server instance.

## Quick Start

### Prerequisites

1. **Running Server**: The server must be running on the specified port (default: 8080)
   ```bash
   make run
   ```

2. **Go Dependencies**: Ensure all dependencies are installed
   ```bash
   go mod tidy
   ```

### Running Tests

```bash
# Run all E2E tests (recommended)
make test.e2e

# Run E2E tests without server check
make test.e2e.only

# Run specific E2E test
make test.e2e.run TEST=TestGRPCHealthService

# Run directly with go test
go test -v ./test/e2e

# Run specific test with go test
go test -v ./test/e2e -run TestConnectGoHealthService

# Run with race detection
go test -v -race ./test/e2e
```

## Architecture

### Directory Structure

```
test/e2e/
├── steps/                              # Reusable step functions
│   ├── bdd_helpers.go                  # BDD framework and test context
│   ├── common_steps.go                 # Common Given/When/Then steps
│   ├── connect_steps.go                # Connect-Go specific steps
│   ├── connect_grpc_steps.go           # gRPC protocol steps
│   ├── grpc_steps.go                   # gRPC service steps
│   ├── user_steps.go                   # User service steps
│   ├── cache_steps.go                  # Cache utility steps
│   └── database_steps.go               # Database health specific steps
├── connect_protocols_test.go           # Tests using both gRPC and REST protocols
├── grpc_health_service_test.go         # Tests using only gRPC protocol
├── grpc_user_service_test.go           # Tests using only gRPC protocol
├── rest_database_health_test.go        # Tests using only REST protocol
├── rest_health_service_test.go         # Tests using only REST protocol
├── rest_monitoring_test.go             # Tests using only REST protocol
├── rest_user_service_test.go           # Tests using only REST protocol
├── cache_utility_test.go               # Cache utility tests
└── const.go                            # Test constants
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
- **gRPC Steps** (`grpc_steps.go`): gRPC service operations
- **User Steps** (`user_steps.go`): User service specific operations
- **Cache Steps** (`cache_steps.go`): Cache utility operations
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
            ctx.TheResponseStatusCodeShouldBe(200)
            ctx.TheJSONResponseShouldContainField("status", "healthy")
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
            ctx.TheResponseStatusCodeShouldBe(testCase.ExpectedStatus)
            ctx.TheJSONResponseShouldContainField("status", testCase.ExpectedValue)
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

#### Then Steps (BDD Style)
- `ctx.TheResponseStatusCodeShouldBe(expectedStatus)`: Checks HTTP status in natural language
- `ctx.TheJSONResponseShouldContainField("status", expectedValue)`: Checks JSON field value
- `ctx.TheResponseShouldNotBeEmpty()`: Validates response is not empty
- `ctx.TheContentTypeShouldBe(expectedContentType)`: Checks Content-Type header
- `ctx.TheJSONResponseShouldHaveStructure([]string{"field1", "field2"})`: Validates JSON structure
- `ctx.AllConcurrentRequestsShouldSucceed()`: Validates all concurrent requests succeeded

### Connect-Go Steps

#### When Steps
- `WhenIMakeHealthCheckRequestUsingConnectGo(ctx)`: Connect-Go health check
- `WhenIMakeHealthCheckRequestWithInvalidContentType(ctx)`: Invalid content type test
- `WhenIMakeConcurrentHealthCheckRequests(ctx, numRequests)`: Concurrent Connect-Go requests

#### Then Steps
- `ThenResponseShouldIndicateServingStatus(ctx)`: Checks SERVING status
- `ThenResponseShouldContainProperConnectGoHeaders(ctx)`: Validates headers

### gRPC Steps

#### When Steps
- `WhenIMakeGRPCHealthCheckRequest(ctx)`: gRPC health check
- `WhenIMakeGRPCUserServiceRequest(ctx, request)`: gRPC user service requests

#### Then Steps
- `ctx.TheGRPCResponseShouldBeSuccessful()`: Validates gRPC response
- `ctx.TheConnectGoResponseShouldBeSuccessful()`: Validates Connect-Go response

### Database Steps

#### Then Steps
- `ThenJSONResponseShouldContainDatabaseInformation(ctx)`: Database info validation
- `ThenJSONResponseShouldContainConnectionStatistics(ctx)`: Connection stats validation
- `ThenDatabaseConnectionPoolShouldBeHealthy(ctx)`: Pool health validation
- `ThenConnectionStatisticsShouldBeValid(ctx)`: Statistics consistency validation

### User Steps

#### When Steps
- `WhenIMakeUserRegistrationRequest(ctx, userData)`: User registration
- `WhenIMakeUserProfileRequest(ctx, userID)`: User profile retrieval

#### Then Steps
- `ThenUserShouldBeCreatedSuccessfully(ctx)`: User creation validation
- `ThenUserProfileShouldBeReturned(ctx)`: Profile data validation

### Cache Steps

#### When Steps
- `WhenICacheData(ctx, key, value)`: Cache data operations
- `WhenIRetrieveCachedData(ctx, key)`: Cache retrieval operations

#### Then Steps
- `ThenCachedDataShouldBeRetrieved(ctx)`: Cache hit validation
- `ThenCacheShouldBeEmpty(ctx)`: Cache miss validation

## Test Categories and Naming Conventions

E2E tests follow a specific naming convention based on the protocol they use:

- **grpc_**: Tests that use only the gRPC protocol
- **rest_**: Tests that use only the REST protocol
- **connect_**: Tests that use both gRPC and REST protocols

This naming convention helps to clearly identify the protocol being tested and ensures consistency across the test suite.

### Connect-Go Protocol Tests
- HTTP/JSON over Connect-Go
- gRPC protocol validation
- Invalid content type handling
- Protocol switching tests
- Concurrent request testing

### gRPC Service Tests
- Health service gRPC tests
- User service gRPC tests
- Service method validation
- Error handling scenarios

### REST API Tests
- Database health endpoint validation
- Monitoring endpoint tests
- Server ping endpoint
- Connection statistics verification

### Cache Utility Tests
- Redis cache operations
- Cache hit/miss scenarios
- Cache invalidation tests
- Concurrent cache access

## Test Features

### BDD Style Testing
Tests are organized using Given-When-Then structure for clear scenario definition with natural language assertions that enhance readability and maintainability.

### Table-Driven Tests
Support for parameterized test scenarios with multiple test cases for comprehensive coverage.

### Concurrent Testing
Built-in support for testing concurrent requests and performance validation:

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

### BDD Style Assertions
Natural language assertions for enhanced readability:

```go
// BDD Style - Natural Language
ctx.TheResponseStatusCodeShouldBe(200)
ctx.TheJSONResponseShouldContainField("status", "healthy")
ctx.TheContentTypeShouldBe("application/json")
ctx.TheResponseShouldNotBeEmpty()
ctx.AllConcurrentRequestsShouldSucceed()

// gRPC and Connect-Go specific
ctx.TheGRPCResponseShouldBeSuccessful()
ctx.TheConnectGoResponseShouldBeSuccessful()
```

## Test Coverage

Unlike unit and integration tests, end-to-end tests do not generate coverage reports. This is because the tests are running against a server in a Docker container, and the coverage data cannot be collected from the running container. The e2e tests focus on verifying that the system works correctly as a whole, rather than measuring code coverage.

## Troubleshooting

### Server Not Running
**Error**: Server is not running on port 8080
**Solution**: Start the server using `make run` or `docker-compose up`

### Connection Refused
**Error**: Failed to connect to server at http://localhost:8080
**Solution**:
1. Check if the server is running
2. Verify the correct port is being used
3. Check firewall settings

### Test Failures
**Error**: Test failed: expected status 200, got 500
**Solution**:
1. Check server logs for errors
2. Verify database connectivity
3. Check environment configuration

### gRPC Connection Issues
**Error**: gRPC connection failed
**Solution**:
1. Verify gRPC server is running
2. Check gRPC port configuration
3. Validate protobuf definitions

## Adding New Tests

1. **Create test functions** in appropriate `*_test.go` files
2. **Use existing step functions** from the `steps/` package
3. **Follow BDD structure** with Given-When-Then organization
4. **Add new step functions** to appropriate `steps/*.go` files if needed
5. **Use table-driven tests** for parameterized scenarios

### Adding New Step Functions

1. Create step functions in appropriate `steps/*.go` files
2. Follow the naming convention: `Given*/When*/Then*`
3. Use the TestContext for state management
4. Add proper error handling and assertions

### Adding New Test Files

1. Create test files in `test/e2e/` directory
2. Import the steps package
3. Use the BDD framework helpers
4. Follow the established patterns for consistency

## Best Practices

1. **Organize Steps**: Group related steps in appropriate files
2. **Reuse Steps**: Use existing step functions when possible
3. **Clear Naming**: Use descriptive scenario names
4. **Error Handling**: Always check for errors in step functions
5. **Concurrent Safety**: Ensure step functions are thread-safe for concurrent tests
6. **Test Isolation**: Each test should be independent and not rely on other tests
7. **Reuse step functions** to maintain consistency
8. **Use descriptive test names** that explain the scenario
9. **Keep tests independent** - each test should be able to run in isolation
10. **Use concurrent testing** for performance validation
11. **Add proper error handling** in custom step functions
12. **Follow Go testing conventions** for test organization

## Migration from Python

The Go BDD framework maintains the same test scenarios as the previous Python implementation:

- **Feature Parity**: All Python test scenarios have been converted
- **BDD Structure**: Maintains Given-When-Then organization
- **Test Coverage**: Same test coverage with improved performance
- **Type Safety**: Compile-time validation of test code

## Advanced Usage

### Custom Test Context

You can extend the TestContext for specific test needs:

```go
type CustomTestContext struct {
    *steps.TestContext
    CustomData map[string]interface{}
}
```

### Performance Testing

Use concurrent testing for performance validation:

```go
steps.WhenIMakeConcurrentRequests(ctx, 100, "/api/endpoint")
ctx.AllConcurrentRequestsShouldCompleteWithin(time.Second * 5)
```

### Integration with CI/CD

E2E tests are designed to run in CI/CD pipelines:

```bash
# In CI/CD pipeline
make run &          # Start server in background
sleep 10           # Wait for server to start
make test.e2e      # Run E2E tests
```

For detailed implementation examples and advanced patterns, refer to the existing test files in the `/test/e2e` directory.
