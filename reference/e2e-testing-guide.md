# E2E Testing Guide

This comprehensive guide covers End-to-End (E2E) testing for the Social Quiz Platform backend, implemented using **Godog v0.14** with **Gherkin** syntax for true BDD (Behavior Driven Development) style testing.

## Overview

The E2E testing framework uses **Godog** (Cucumber for Go) with **Gherkin** feature files to provide human-readable test scenarios. Tests verify the complete functionality of the service by making actual HTTP/gRPC requests to a running server instance.

### Migration from Custom BDD Framework

The project has migrated from a custom Go BDD framework to **Godog v0.14 + Gherkin**, providing:

- **Human-readable scenarios**: Gherkin syntax with natural language
- **Protocol separation**: Separate feature files for gRPC positive and REST negative tests
- **Scenario Outlines**: Parameterized tests using Examples tables to minimize duplication
- **Step definitions**: Reusable step functions in Go for Given/When/Then statements
- **Comprehensive coverage**: All original test scenarios preserved and enhanced

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
# Run all godog E2E tests (recommended)
make test.e2e

# Run with verbose output
VERBOSE=1 make test.e2e

# Run specific scenarios by tags
make test.e2e.only TAGS=@health

# Run specific scenarios by tags and scenario name
make test.e2e.only TAGS=@health SCENARIO="Basic health check"

# Run directly with go test
go test -v ./test/e2e

# Run with specific tags using godog
go test ./test/e2e -godog.tags="@rest"
go test ./test/e2e -godog.tags="@grpc"
go test ./test/e2e -godog.tags="@health"
go test ./test/e2e -godog.tags="@user"

# Run with protocol-specific environment variables
HEALTH_PROTOCOL=rest go test ./test/e2e
USER_PROTOCOL=grpc go test ./test/e2e
```

## Architecture

### Directory Structure

```text
test/e2e/
├── main_test.go                        # Test harness with TestMain and godog setup
├── features/                           # Gherkin feature files
│   ├── health/                         # Health-related features
│   │   ├── rest.feature               # REST health endpoints
│   │   └── grpc.feature               # gRPC health endpoints
│   └── user/                          # User-related features
│       └── grpc.feature               # gRPC user service
└── steps/                             # Step definitions
    ├── hooks.go                       # Common hooks and initialization
    ├── health_steps.go                # Health-specific step definitions
    └── user_steps.go                  # User-specific step definitions
```

### Core Components

#### 1. Feature Files (Gherkin)

Feature files use **Gherkin syntax** to define human-readable test scenarios:

- **Background**: Common setup steps shared across scenarios
- **Scenario**: Individual test cases with Given-When-Then structure
- **Scenario Outline**: Parameterized tests with Examples tables
- **Tags**: Organize tests by protocol (@rest, @grpc), domain (@health, @user), and type (@positive, @negative)

#### 2. Step Definitions (Go)

Step definitions implement the Gherkin steps in Go:

- **Health Steps** (`health_steps.go`): Health service specific operations
- **User Steps** (`user_steps.go`): User service specific operations
- **Hooks** (`hooks.go`): Common setup, teardown, and context management

#### 3. Test Harness (`main_test.go`)

- **TestMain**: Handles protocol-specific test execution
- **TestFeatures**: Main godog test runner with configurable options
- **Protocol Support**: Environment variable-based protocol selection

## Features Covered

### Health Service

#### REST Endpoints
- **`GET /api/ping`** - Simple health check
- **`GET /api/healthz`** - Comprehensive health check
- **`GET /api/health/database`** - Database connectivity check

#### gRPC Endpoints
- **`grpc.health.v1.Health/Check`** - gRPC health check (via HTTP/2 on port 8080)

### User Service

#### gRPC Endpoints
- **`user.v1.UserService/RegisterUser`** - User registration
- **`user.v1.UserService/GetUserProfile`** - User profile retrieval
- **`user.v1.UserService/UpdateUserProfile`** - User profile updates

## Writing Gherkin Feature Files

### Basic Scenario

```gherkin
@rest @health @positive
Feature: REST Health Endpoints
  As a monitoring system
  I want to check the health of the service via REST endpoints
  So that I can ensure the service is running properly

  Background:
    Given the server is running

  Scenario: Health endpoint responds correctly
    When I make a GET request to "/api/healthz"
    Then the HTTP status should be 200
    And the response should contain status "healthy"
```

### Scenario Outline with Examples

```gherkin
@rest @health @positive
Scenario Outline: Health endpoints respond correctly
  When I make a GET request to "<endpoint>"
  Then the HTTP status should be <status>
  And the response should contain status "<expected_status>"

  Examples:
    | endpoint             | status | expected_status |
    | /api/ping           | 200    | ok              |
    | /api/healthz        | 200    | healthy         |
    | /api/health/database| 200    | healthy         |
```

### gRPC Feature Example

```gherkin
@grpc @user @positive
Feature: gRPC User Service
  As a gRPC client
  I want to manage user profiles via gRPC protocol
  So that I can register users and retrieve profiles

  Background:
    Given the server is running
    And the gRPC user client is connected

  Scenario: User registration with valid data should succeed
    Given I have valid user registration data
    When I register a user with valid data
    Then the user registration should succeed
    And the response should contain a valid user ID
```

## Available Step Definitions

### Health Service Steps

#### Given Steps
- `the server is running` - Verifies server accessibility
- `the gRPC client is connected` - Establishes gRPC connection
- `the gRPC client is connected with timeout "5s"` - Establishes gRPC connection with custom timeout
- `the Connect-Go client is configured with "grpc-web" protocol` - Sets up Connect-Go client

#### When Steps
- `I call the "REST|gRPC" health endpoint` - Makes health request via specified protocol
- `I make a GET request to "/path"` - Makes HTTP GET request
- `I make a Connect-Go health request` - Makes Connect-Go style request
- `I make a gRPC health check request` - Makes gRPC health check
- `I make a gRPC health check request with metadata` - Makes gRPC request with metadata
- `I make 5 concurrent "REST|gRPC" health requests` - Makes concurrent requests

#### Then Steps
- `the status should be healthy` - Validates healthy status (both protocols)
- `the HTTP status should be 200` - Validates HTTP status code
- `the response should contain status "healthy"` - Validates JSON status field
- `the response should contain field "field" with value "value"` - Validates JSON field
- `the gRPC response should indicate serving status` - Validates gRPC SERVING status
- `the gRPC response should not be empty` - Validates gRPC response exists
- `the gRPC response should contain metadata` - Validates gRPC metadata
- `all concurrent "REST|gRPC" requests should succeed` - Validates concurrent results

### User Service Steps

#### Given Steps
- `the gRPC user client is connected` - Establishes gRPC user client connection
- `the gRPC-Web user client is connected` - Establishes gRPC-Web user client connection
- `I have valid user registration data` - Sets up valid registration data
- `I have invalid user registration data with empty Firebase UID` - Sets up invalid data
- `an existing user is registered` - Creates a test user for subsequent operations

#### When Steps
- `I register a user with valid data` - Performs user registration
- `I register a user with the same Firebase UID` - Tests duplicate registration
- `I register a user with empty Firebase UID` - Tests invalid registration
- `I get the user profile` - Retrieves user profile
- `I get the user profile with invalid ID` - Tests invalid profile retrieval
- `I get the user profile with non-existent ID` - Tests non-existent profile retrieval
- `I update the user profile with display name "..."` - Updates user profile
- `I make 5 concurrent user registrations` - Tests concurrent registrations

#### Then Steps
- `the user registration should succeed` - Validates successful registration
- `the response should contain a valid user ID` - Validates UUID format
- `the user registration should fail with AlreadyExists error` - Validates duplicate error
- `the user registration should fail with InvalidArgument error` - Validates invalid data error
- `the user profile retrieval should fail with NotFound error` - Validates not found error
- `the user profile should be retrieved` - Validates successful profile retrieval
- `the user profile should contain display name "..."` - Validates profile data
- `the user profile update should succeed` - Validates successful update
- `all concurrent user registrations should succeed` - Validates concurrent operations

## Test Organization and Tags

### Tag-Based Organization

Tests are organized using **Gherkin tags** for flexible execution:

- **Protocol Tags**: `@rest`, `@grpc`, `@connect` - Filter by communication protocol
- **Domain Tags**: `@health`, `@user` - Filter by service domain
- **Type Tags**: `@positive`, `@negative` - Filter by test type
- **Feature Tags**: `@concurrency` - Filter by specific features

### Feature File Structure

#### Health Features
- **`health/rest.feature`** - REST health endpoints with positive and negative scenarios
- **`health/grpc.feature`** - gRPC health service with timeout and metadata testing

#### User Features
- **`user/grpc.feature`** - gRPC user service with comprehensive CRUD operations

### Protocol Separation

The project follows a **dual-protocol architecture**:

- **REST**: Health checks and monitoring only (`/api/ping`, `/api/healthz`, `/api/health/database`)
- **gRPC**: All business logic (user management, future services)
- **Protocol Filter Middleware**: Blocks HTTP/JSON requests to gRPC-only endpoints

### Test Execution Patterns

- **Positive Tests**: Verify expected functionality works correctly
- **Negative Tests**: Verify proper error handling and edge cases
- **Concurrent Tests**: Validate system behavior under load
- **Protocol Tests**: Ensure protocol separation is enforced

## Test Features

### Gherkin BDD Style
Tests use **Gherkin syntax** for human-readable scenarios:

```gherkin
Feature: REST Health Endpoints
  As a monitoring system
  I want to check the health of the service
  So that I can ensure the service is running properly

  Scenario: Health endpoint responds correctly
    Given the server is running
    When I make a GET request to "/api/healthz"
    Then the HTTP status should be 200
    And the response should contain status "healthy"
```

### Scenario Outlines
**Parameterized testing** using Examples tables to minimize duplication:

```gherkin
Scenario Outline: Health endpoints respond correctly
  When I make a GET request to "<endpoint>"
  Then the HTTP status should be <status>
  And the response should contain status "<expected_status>"

  Examples:
    | endpoint             | status | expected_status |
    | /api/ping           | 200    | ok              |
    | /api/healthz        | 200    | healthy         |
    | /api/health/database| 200    | healthy         |
```

### Concurrent Testing
Built-in support for testing concurrent requests:

```gherkin
Scenario Outline: Concurrent user registrations
  When I make <num_requests> concurrent user registrations
  Then all concurrent user registrations should succeed

  Examples:
    | num_requests |
    | 3            |
    | 5            |
    | 10           |
```

### Error Handling
Comprehensive error scenario testing with natural language:

```gherkin
@negative
Scenario: User registration with duplicate Firebase UID should fail
  Given an existing user is registered
  When I register a user with the same Firebase UID
  Then the user registration should fail with AlreadyExists error
```

## Test Coverage

Unlike unit and integration tests, end-to-end tests do not generate coverage reports. This is because the tests are running against a server in a Docker container, and the coverage data cannot be collected from the running container. The e2e tests focus on verifying that the system works correctly as a whole, rather than measuring code coverage.

## Troubleshooting

### Godog Test Execution Issues

**Error**: `godog: no feature files found`

**Solution**:

1. Verify feature files exist in `/test/e2e/features/`
2. Check file extensions are `.feature`
3. Ensure working directory is correct

**Error**: `undefined step: "I make a GET request to..."`

**Solution**:

1. Check step definitions in `/test/e2e/steps/`
2. Verify step function signatures match Gherkin steps
3. Ensure step definitions are registered in `main_test.go`

### Server Connection Issues

**Error**: Server is not running on port 8080

**Solution**: Start the server using `make run` or `docker-compose up`

**Error**: Failed to connect to server at `http://localhost:8080`

**Solution**:

1. Check if the server is running
2. Verify the correct port is being used
3. Check firewall settings
4. Ensure `.env` configuration is correct

### gRPC Connection Issues

**Error**: gRPC connection failed

**Solution**:

1. Verify gRPC server is running on HTTP/2
2. Check gRPC port configuration (default: 8080)
3. Validate protobuf definitions are up-to-date
4. Ensure Connect-Go middleware is properly configured

### Test Execution Failures

**Error**: Test failed: expected status 200, got 500

**Solution**:

1. Check server logs for errors
2. Verify database connectivity
3. Check environment configuration
4. Validate test data setup in step definitions

**Error**: `panic: runtime error: invalid memory address`

**Solution**:

1. Check for nil pointer dereferences in step definitions
2. Verify proper context initialization in hooks
3. Ensure proper cleanup in scenario teardown

## Adding New Tests

### 1. Create Feature Files

Create new `.feature` files in `/test/e2e/features/<domain>/`:

```gherkin
@grpc @quiz @positive
Feature: gRPC Quiz Service
  As a quiz application
  I want to manage quizzes via gRPC protocol
  So that users can create and take quizzes

  Background:
    Given the server is running
    And the gRPC quiz client is connected

  Scenario: Create quiz with valid data should succeed
    Given I have valid quiz creation data
    When I create a quiz with valid data
    Then the quiz creation should succeed
    And the response should contain a valid quiz ID
```

### 2. Implement Step Definitions

Create step definition files in `/test/e2e/steps/<domain>_steps.go`:

```go
package steps

import (
    "context"
    "github.com/cucumber/godog"
)

type QuizContext struct {
    // Quiz-specific context fields
    quizClient QuizServiceClient
    quizData   *QuizData
    response   *QuizResponse
    err        error
}

func (qc *QuizContext) iHaveValidQuizCreationData() error {
    qc.quizData = &QuizData{
        Title:       "Sample Quiz",
        Description: "A sample quiz for testing",
    }
    return nil
}

func (qc *QuizContext) iCreateAQuizWithValidData() error {
    qc.response, qc.err = qc.quizClient.CreateQuiz(context.Background(), qc.quizData)
    return nil
}

func (qc *QuizContext) theQuizCreationShouldSucceed() error {
    if qc.err != nil {
        return fmt.Errorf("expected quiz creation to succeed, but got error: %v", qc.err)
    }
    return nil
}

func InitializeQuizScenario(ctx *godog.ScenarioContext) {
    qc := &QuizContext{}

    ctx.Given(`^I have valid quiz creation data$`, qc.iHaveValidQuizCreationData)
    ctx.When(`^I create a quiz with valid data$`, qc.iCreateAQuizWithValidData)
    ctx.Then(`^the quiz creation should succeed$`, qc.theQuizCreationShouldSucceed)
}
```

### 3. Register Step Definitions

Update `/test/e2e/main_test.go` to register new step definitions:

```go
func TestFeatures(t *testing.T) {
    suite := godog.TestSuite{
        ScenarioInitializer: func(ctx *godog.ScenarioContext) {
            // Existing step definitions
            steps.InitializeHealthScenario(ctx)
            steps.InitializeUserScenario(ctx)

            // Add new step definitions
            steps.InitializeQuizScenario(ctx)
        },
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"features"},
            TestingT: t,
        },
    }

    if suite.Run() != 0 {
        t.Fatal("non-zero status returned, failed to run feature tests")
    }
}
```

### 4. Test Execution

Run the new tests:

```bash
# Run all tests including new quiz tests
make test.e2e

# Run only quiz tests
make test.e2e.only TAGS=@quiz

# Run specific quiz scenarios
make test.e2e.only TAGS=@quiz SCENARIO="Create quiz with valid data"
```

## Best Practices

### Feature File Organization

- **One feature per service domain** (health, user, quiz, etc.)
- **Separate files by protocol** (rest.feature, grpc.feature)
- **Use descriptive scenario names** that explain the business value
- **Group related scenarios** using Background sections

### Step Definition Guidelines

- **Keep steps atomic** - each step should do one thing
- **Use context structs** for scenario state management
- **Implement proper cleanup** in hooks
- **Follow naming conventions** - use domain prefixes for context structs

### Tag Strategy

- **Protocol tags**: `@rest`, `@grpc` for protocol filtering
- **Domain tags**: `@health`, `@user`, `@quiz` for service filtering
- **Type tags**: `@positive`, `@negative` for test type filtering
- **Feature tags**: `@concurrency`, `@security` for special features

### Error Handling

- **Use descriptive error messages** in step definitions
- **Validate all assumptions** in Given steps
- **Check for nil pointers** before accessing response data
- **Implement proper timeout handling** for gRPC calls

## Contributing

When adding new E2E tests:

1. Follow the existing naming conventions and tag strategy
2. Use the Godog framework for consistency with Gherkin syntax
3. Add appropriate error handling in step definitions
4. Include both positive and negative test scenarios
5. Test concurrent scenarios where applicable
6. Organize feature files by domain and protocol
7. Update this documentation if adding new patterns or conventions

## Integration with CI/CD

E2E tests are designed to run in CI/CD pipelines:

```bash
# In CI/CD pipeline
make run &          # Start server in background
sleep 10           # Wait for server to start
make test.e2e      # Run E2E tests
```

## Summary

The E2E testing framework has successfully migrated from a custom BDD implementation to **Godog v0.14 + Gherkin**, providing:

- **Human-readable test scenarios** using natural language
- **Comprehensive protocol coverage** for both REST and gRPC
- **Flexible test execution** with tag-based filtering
- **Maintainable test organization** with domain-based feature files
- **Robust error handling** for both positive and negative scenarios
- **Concurrent testing support** for performance validation

This migration maintains all existing test coverage while providing a more standard, maintainable, and extensible testing framework for future development.

For detailed implementation examples and advanced patterns, refer to the existing test files in the `/test/e2e` directory.
