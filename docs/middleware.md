# Middleware Documentation

This document describes the middleware components used in the Sudal application.

## Overview

Sudal uses a middleware stack to handle cross-cutting concerns such as logging, protocol filtering, and request processing. Middleware components are applied in a specific order to ensure proper request handling.

## Middleware Stack

The middleware stack is applied in the following order:

1. **Protocol Filter Middleware** - Enforces protocol restrictions
2. **Request Logger Middleware** - Logs HTTP requests and responses

## Protocol Filter Middleware

### Purpose

The Protocol Filter Middleware enforces the dual-protocol architecture by restricting certain endpoints to gRPC-only access while preserving REST access for health and monitoring endpoints.

### Location

- **File**: `internal/infrastructure/middleware/protocol_filter.go`
- **Tests**: `internal/infrastructure/middleware/protocol_filter_test.go`

### Functionality

#### gRPC-Only Enforcement

The middleware blocks HTTP/JSON requests to specified gRPC-only paths and only allows:

- **gRPC over HTTP/2** (`application/grpc` content type)
- **gRPC-Web** (`application/grpc-web` content type)
- **HTTP/2 with gRPC indicators** (TE: trailers header, gRPC user agents)

#### Protected Paths

The following paths are restricted to gRPC-only access:

- `/health.v1.HealthService/` - gRPC health service
- `/user.v1.UserService/` - All user service methods

#### Security Features

**Protocol-Level Security:**
- **Endpoint Discovery Protection**: Returns 404 Not Found for HTTP/JSON requests to hide endpoint existence
- **Binary Protocol Advantage**: gRPC's binary Protocol Buffers make traffic analysis more difficult than REST
- **Schema Obfuscation**: Attackers need .proto files to understand request/response structures
- **HTTP/2 Complexity**: More sophisticated protocol reduces casual exploration attempts

**Implementation Security:**
- **Protocol Detection**: Sophisticated detection of gRPC protocols vs HTTP/JSON
- **Path-Based Filtering**: Precise control over which endpoints are restricted
- **Detailed Logging**: Logs blocked requests and allowed gRPC requests for monitoring and security analysis

### Configuration

```go
// Get the list of gRPC-only paths
grpcOnlyPaths := middleware.GetGRPCOnlyPaths()

// Apply the middleware
protocolFilterHandler := middleware.ProtocolFilterMiddleware(grpcOnlyPaths)(mux)
```

### Usage Examples

#### Allowed Requests

```bash
# ✅ REST health endpoints
curl http://localhost:8080/api/ping
curl http://localhost:8080/api/healthz

# ✅ gRPC requests
grpcurl -plaintext -proto proto/health/v1/health.proto \
  localhost:8080 health.v1.HealthService/Check
```

#### Blocked Requests

```bash
# ❌ HTTP/JSON to gRPC-only endpoints (returns 404)
curl -X POST -H "Content-Type: application/json" \
  -d '{}' http://localhost:8080/health.v1.HealthService/Check

curl -X POST -H "Content-Type: application/json" \
  -d '{}' http://localhost:8080/user.v1.UserService/RegisterUser
```

### Implementation Details

#### Protocol Detection Logic

The middleware uses multiple indicators to detect gRPC requests:

1. **Content-Type Headers**:
   - `application/grpc*` - Standard gRPC
   - `application/grpc-web*` - gRPC-Web

2. **HTTP/2 Indicators**:
   - `TE: trailers` header (required for gRPC over HTTP/2)
   - `X-Grpc-Web` header (gRPC-Web specific)

3. **User Agent Patterns**:
   - Contains "grpc" or "connect" keywords

4. **Request Characteristics**:
   - HTTP/2 POST requests to service paths
   - Non-JSON content types on service paths

#### Path Matching

The middleware uses prefix matching to determine if a path should be restricted:

```go
func shouldRestrictToGRPC(requestPath string, grpcOnlyPaths []string) bool {
    for _, path := range grpcOnlyPaths {
        if strings.HasPrefix(requestPath, path) {
            return true
        }
    }
    return false
}
```

### Testing

The middleware includes comprehensive tests covering:

- **Non-gRPC paths**: Ensures unrestricted access to REST endpoints
- **HTTP/JSON blocking**: Verifies blocking of HTTP/JSON requests to gRPC paths
- **gRPC allowance**: Confirms proper handling of various gRPC protocols
- **Edge cases**: Handles malformed requests and partial path matches

Run tests with:

```bash
cd internal/infrastructure/middleware
ginkgo run
```

### Monitoring and Logging

The middleware provides detailed logging for monitoring and debugging:

#### Blocked Requests

```json
{
  "level": "warn",
  "message": "HTTP/JSON request blocked for gRPC-only endpoint",
  "path": "/user.v1.UserService/RegisterUser",
  "method": "POST",
  "content_type": "application/json",
  "user_agent": "curl/7.68.0"
}
```

#### Allowed Requests

```json
{
  "level": "info",
  "message": "gRPC request allowed",
  "path": "/health.v1.HealthService/Check",
  "protocol": "grpc"
}
```

## Request Logger Middleware

### Purpose

The Request Logger Middleware provides structured logging for all HTTP requests and responses.

### Location

- **File**: `internal/infrastructure/middleware/logging.go`

### Functionality

- Logs request details (method, path, user agent, etc.)
- Logs response details (status code, response time, etc.)
- Includes trace IDs for request correlation
- Uses structured JSON logging via zap

## Adding New Middleware

To add new middleware to the stack:

1. **Create the middleware function** in `internal/infrastructure/middleware/`
2. **Follow the standard middleware pattern**:
   ```go
   func MyMiddleware() func(http.Handler) http.Handler {
       return func(next http.Handler) http.Handler {
           return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
               // Pre-processing
               next.ServeHTTP(w, r)
               // Post-processing
           })
       }
   }
   ```
3. **Add tests** following the existing test patterns
4. **Update the middleware stack** in `internal/infrastructure/server/server.go`
5. **Update this documentation**

## Best Practices

- **Order matters**: Apply middleware in the correct order (security → logging → business logic)
- **Context preservation**: Pass context through the middleware chain
- **Error handling**: Handle errors gracefully and provide meaningful responses
- **Performance**: Keep middleware lightweight to avoid request latency
- **Testing**: Write comprehensive tests for all middleware components
- **Logging**: Use structured logging for observability
