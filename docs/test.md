# Testing

## Running Tests

The project has three types of tests:

1. **Unit Tests**: Test individual components in isolation
2. **Integration Tests**: Test interactions between components
3. **End-to-End Tests**: Test the entire system with a running server

### Running All Tests

To run both unit and integration tests (with preparation steps run only once):

```bash
make test
```

### Running Specific Test Types

To run only unit tests (with preparation steps):

```bash
make test.unit
```

To run only unit tests (without preparation steps):

```bash
make test.unit.only
```

To run only integration tests (with preparation steps):

```bash
make test.int
```

To run only integration tests (without preparation steps):

```bash
make test.int.only
```

To run end-to-end tests (with preparation steps):

```bash
make test.e2e
```

To run end-to-end tests (without preparation steps):

```bash
make test.e2e.only
```

### What Each Test Command Does

Each test command first runs `make test.prepare`, which:
1. Formats the code with `go fmt`
2. Runs static analysis with `go vet`
3. Runs linter checks with `golangci-lint`
4. Runs all code generation tasks via `make generate`

Then, the test command:
1. Runs the specified tests with Ginkgo
2. Generates a coverage report (both console summary and HTML report)

After running tests, you can view the detailed coverage reports:
- Unit tests: `coverage.unit.html`
- Integration tests: `coverage.int.html`
- End-to-end tests: `coverage.e2e.html`

Note: Integration and E2E tests measure coverage of the internal packages using the `-coverpkg` flag.

To run specific tests manually:

```bash
go test ./path/to/package -v
```

## Testing Strategy

The project follows a Behavior-Driven Development (BDD) approach to testing using the following tools:

- **Ginkgo**: A BDD-style testing framework for Go that provides a more expressive and readable syntax for writing unit and integration tests.
- **Gomega**: An assertion library that pairs with Ginkgo to provide a rich set of matchers for making assertions in tests.
- **Godog**: A Cucumber-style BDD framework for Go used specifically for end-to-end tests with Gherkin syntax.
- **mockgen**: Used to generate mock implementations of interfaces for testing.
- **httptest**: Standard library package for testing HTTP handlers and servers.
- **go-sqlmock**: Used for mocking database interactions in tests.

## Test Structure

- **Unit Tests**: Located alongside the code they test in the `/internal` directory
- **Integration Tests**: Located in the `/test/integration` directory
- **End-to-End Tests**: Located in the `/test/e2e` directory

## End-to-End Tests

End-to-End tests use **Godog v0.14** with **Gherkin** syntax for human-readable BDD scenarios. These tests verify the complete functionality of the service by making actual HTTP/gRPC requests to a running server instance.

### Quick Start

```bash
# Run all E2E tests
make test.e2e

# Run specific tests by tags
make test.e2e.only TAGS=@health
make test.e2e.only TAGS=@user
make test.e2e.only TAGS=@rest
make test.e2e.only TAGS=@grpc
```

### Features Covered

- **Health Service**: REST endpoints (`/api/ping`, `/api/healthz`, `/api/health/database`) and gRPC health checks
- **User Service**: gRPC user management (registration, profile retrieval, updates)

### Test Organization

- **Feature Files**: Located in `/test/e2e/features/` using Gherkin syntax
- **Step Definitions**: Located in `/test/e2e/steps/` with Go implementations
- **Tag-Based Execution**: Use `@rest`, `@grpc`, `@health`, `@user`, `@positive`, `@negative` tags

For comprehensive information about E2E testing, including Gherkin syntax, step definitions, troubleshooting, and adding new tests, please refer to the [E2E Testing Guide](e2e-testing-guide.md).

## Writing BDD Tests

Tests follow the BDD style using Ginkgo's `Describe`, `Context`, and `It` blocks to organize test cases:

```go
var _ = Describe("Handler", func() {
    Context("when handling a valid request", func() {
        It("should return a successful response", func() {
            // Test code here
            Expect(response.StatusCode).To(Equal(http.StatusOK))
        })
    })
})
```

## Mocking Dependencies

### Mock Generation with mockgen

The project uses `mockgen` to automatically generate mock implementations of interfaces for testing. All mock files are stored in the `/internal/mocks` directory to maintain consistency and organization.

#### Mock File Location

- **Directory**: All generated mocks are stored in `/internal/mocks`
- **Naming Convention**: Mock files follow the pattern `mock_<feature_name>.go`
- **Package**: All mocks use the `mocks` package

#### Generating Mocks

Interfaces are mocked using `mockgen` with the following pattern:

```go
//go:generate mockgen -destination=../../internal/mocks/mock_service.go -package=mocks github.com/seventeenthearth/sudal/internal/feature/example Service
```

#### Example Mock Generation

For a repository interface in the health feature:

```go
//go:generate mockgen -destination=../../internal/mocks/mock_health_repository.go -package=mocks github.com/seventeenthearth/sudal/internal/feature/health/domain/repo HealthRepository
```

#### Using Generated Mocks

Generated mocks can be used in tests as follows:

```go
import (
    "github.com/seventeenthearth/sudal/internal/mocks"
    "go.uber.org/mock/gomock"
)

func TestExample(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockHealthRepository(ctrl)
    mockRepo.EXPECT().SomeMethod().Return(expectedResult, nil)

    // Use the mock in your test
}
```

#### Regenerating All Mocks

To regenerate all mocks in the project:

```bash
make generate
```

This command will run all `//go:generate` directives and update the mock files in `/internal/mocks`.

## Testing Protocol Buffer APIs

### JSON Serialization in Tests

When testing APIs that use Protocol Buffers with Connect-go, be aware of how protobuf types are serialized to JSON:

1. **Enum Values**: Protobuf enum values are serialized to JSON using their full enum name. For example:

```go
// In your .proto file
enum ServingStatus {
    SERVING_STATUS_UNKNOWN_UNSPECIFIED = 0;
    SERVING_STATUS_SERVING = 1;
    SERVING_STATUS_NOT_SERVING = 2;
}

// In your test, expect the full enum name in JSON responses
Expect(response.Status).To(Equal("SERVING_STATUS_SERVING"))
```

2. **Field Names**: Protobuf field names use camelCase in JSON (e.g., `user_id` becomes `userId`).

3. **Default Values**: Fields with default values are typically omitted from JSON output.

These behaviors are important to consider when writing tests that verify API responses, especially in end-to-end tests that make HTTP/JSON requests to the server.
