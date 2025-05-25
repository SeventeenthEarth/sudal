# E2E Tests

This directory contains End-to-End (E2E) tests for the Social Quiz Platform backend, implemented using Go with testify in a BDD style.

## Overview

The E2E tests verify the complete functionality of the service by making actual HTTP requests to a running server instance. These tests use Behavior Driven Development (BDD) approach with Given-When-Then structure.

## Test Structure

```
test/e2e/
├── steps/                              # Reusable step functions
│   ├── bdd_helpers.go                  # BDD framework and test context
│   ├── common_steps.go                 # Common Given/When/Then steps
│   ├── connect_steps.go                # Connect-Go specific steps
│   └── database_steps.go               # Database health specific steps
├── connect_health_service_test.go      # Connect-Go health service tests
├── rest_database_health_test.go        # REST database health tests
├── rest_monitoring_test.go             # REST monitoring endpoint tests
└── run_go_tests.sh                     # Test execution script
```

## Prerequisites

1. **Running Server**: The server must be running on the specified port (default: 8080)
2. **Go Dependencies**: Ensure testify is installed (`go mod tidy`)

## Setup

1. **Start the Server**:
   ```bash
   # Option 1: Using Make
   make run

   # Option 2: Using Docker Compose directly
   docker-compose up --build
   ```

## Running Tests

### Using Make (Recommended)

```bash
# Run all E2E tests
make test.e2e

# Run only E2E tests (without preparation steps)
make test.e2e.only
```

### Using go test directly

```bash
# Run all E2E tests
go test -v ./test/e2e

# Run specific test
go test -v ./test/e2e -run TestConnectGoHealthService

# Run with race detection
go test -v -race ./test/e2e

# Run with coverage
go test -v -coverprofile=coverage.out ./test/e2e
```

### Using the test runner script

```bash
# Make the script executable (first time only)
chmod +x test/e2e/run_go_tests.sh

# Run all tests
./test/e2e/run_go_tests.sh

# Run specific test
./test/e2e/run_go_tests.sh TestConnectGoHealthService

# Skip server check
./test/e2e/run_go_tests.sh -s
```

## Test Features

### BDD Style Testing

Tests are organized using Given-When-Then structure for clear scenario definition.

### Table-Driven Tests

Support for parameterized test scenarios with multiple test cases.

### Concurrent Testing

Built-in support for testing concurrent requests and performance validation.

## Test Categories

### Connect-Go Health Service Tests
- Health check using Connect-Go client
- HTTP/JSON over Connect-Go
- Invalid content type handling
- Non-existent endpoint handling
- Concurrent request testing

### REST Database Health Tests
- Database health endpoint validation
- Connection statistics verification
- Timestamp field validation
- Connection pool health checks
- Concurrent database health requests

### REST Monitoring Tests
- Server ping endpoint
- Basic health endpoint
- Lightweight response validation
- Multiple endpoint accessibility
- Performance characteristics

## Available Step Functions

### Common Steps
- `GivenServerIsRunning(ctx)`: Server availability check
- `WhenIMakeGETRequest(ctx, endpoint)`: HTTP GET requests
- `WhenIMakePOSTRequest(ctx, endpoint, contentType, body)`: HTTP POST requests
- `ThenHTTPStatusShouldBe(ctx, status)`: Status code validation
- `ThenJSONResponseShouldContainStatus(ctx, status)`: JSON status validation

### Connect-Go Steps
- `WhenIMakeHealthCheckRequestUsingConnectGo(ctx)`: Connect-Go health check
- `ThenResponseShouldIndicateServingStatus(ctx)`: SERVING status validation
- `ThenResponseShouldContainProperConnectGoHeaders(ctx)`: Header validation

### Database Steps
- `ThenJSONResponseShouldContainDatabaseInformation(ctx)`: Database info validation
- `ThenDatabaseConnectionPoolShouldBeHealthy(ctx)`: Pool health validation
- `ThenConnectionStatisticsShouldBeValid(ctx)`: Statistics validation

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

## Adding New Tests

1. **Create test functions** in appropriate `*_test.go` files
2. **Use existing step functions** from the `steps/` package
3. **Follow BDD structure** with Given-When-Then organization
4. **Add new step functions** to appropriate `steps/*.go` files if needed
5. **Use table-driven tests** for parameterized scenarios

## Best Practices

1. **Reuse step functions** to maintain consistency
2. **Use descriptive test names** that explain the scenario
3. **Keep tests independent** - each test should be able to run in isolation
4. **Use concurrent testing** for performance validation
5. **Add proper error handling** in custom step functions
6. **Follow Go testing conventions** for test organization

For detailed information about the BDD framework and advanced usage, see [Go E2E Testing Documentation](../../docs/e2e-testing-go.md).