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

- **Ginkgo**: A BDD-style testing framework for Go that provides a more expressive and readable syntax for writing tests.
- **Gomega**: An assertion library that pairs with Ginkgo to provide a rich set of matchers for making assertions in tests.
- **mockgen**: Used to generate mock implementations of interfaces for testing.
- **httptest**: Standard library package for testing HTTP handlers and servers.
- **go-sqlmock**: Used for mocking database interactions in tests.

## Test Structure

- **Unit Tests**: Located alongside the code they test in the `/internal` directory
- **Integration Tests**: Located in the `/test/integration` directory
- **End-to-End Tests**: Located in the `/test/e2e` directory

## End-to-End Tests

End-to-end tests verify that the entire system works correctly by testing against a running server. These tests:

1. Require the server to be running in Docker (using `make run` in a separate terminal)
2. Connect to the server and verify that it responds correctly
3. Test actual functionality by making requests to the server
4. Fail if the server is not accessible or not functioning correctly

To run end-to-end tests:

```bash
# First, start the server in a separate terminal
make run

# Then, in another terminal, run the e2e tests
make test.e2e
```

### Coverage for End-to-End Tests

Unlike unit and integration tests, end-to-end tests do not generate coverage reports. This is because the tests are running against a server in a Docker container, and the coverage data cannot be collected from the running container. The e2e tests focus on verifying that the system works correctly as a whole, rather than measuring code coverage.

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

Interfaces are mocked using `mockgen` to isolate the component being tested:

```go
//go:generate mockgen -destination=../mocks/mock_service.go -package=mocks github.com/seventeenthearth/sudal/internal/feature/example Service
```

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
