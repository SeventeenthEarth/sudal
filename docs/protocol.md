# API Protocol Documentation

This document provides a comprehensive overview of the API protocols available in the Sudal application. The application supports both REST API and gRPC interfaces.

## Protocol Layer Structure

The protocol layer in the clean architecture is responsible for handling external communication. Each feature's protocol layer typically contains:

- **grpc_manager.go**: Handles gRPC protocol communication
- **rest_handler.go**: Handles REST API endpoints (may not exist for features that don't use REST)

## REST API Endpoints

### Health Check Endpoints

#### Ping Endpoint

- **Type**: REST
- **Endpoint**: `GET /ping`
- **Role**: Simple health check to verify the service is alive and responding
- **Input**: None
- **Output**: JSON response with status field
  ```json
  {
    "status": "ok"
  }
  ```
- **Usage**:
  ```bash
  curl -X GET http://localhost:8080/ping
  ```
- **Notes**: This endpoint is useful for simple liveness probes in container orchestration systems like Kubernetes.

#### Health Check Endpoint

- **Type**: REST
- **Endpoint**: `GET /healthz`
- **Role**: Comprehensive health check that verifies the service and its dependencies are functioning correctly
- **Input**: None
- **Output**: JSON response with status field
  ```json
  {
    "status": "healthy"
  }
  ```
- **Usage**:
  ```bash
  curl -X GET http://localhost:8080/healthz
  ```
- **Notes**: This endpoint is suitable for readiness probes in Kubernetes, as it checks the health of dependencies like databases.

## gRPC Services

### Health Service

- **Type**: gRPC (with Connect-go)
- **Service**: `health.v1.HealthService`
- **Endpoint**: `/health.v1.HealthService/Check`
- **Role**: Provides health status information about the service
- **Methods**:

#### Check

- **Role**: Returns the current health status of the service
- **Input**: Empty `CheckRequest` message
  ```protobuf
  message CheckRequest {
    // Empty for now, can be extended in the future
  }
  ```
- **Output**: `CheckResponse` message with a status enum
  ```protobuf
  message CheckResponse {
    ServingStatus status = 1;
  }
  
  enum ServingStatus {
    SERVING_STATUS_UNKNOWN_UNSPECIFIED = 0;
    SERVING_STATUS_SERVING = 1;
    SERVING_STATUS_NOT_SERVING = 2;
  }
  ```
- **Usage**:
  - **gRPC**: Use a gRPC client to call the service
  - **HTTP/JSON**: Send a POST request with an empty JSON body
    ```bash
    curl -X POST \
      -H "Content-Type: application/json" \
      -d '{}' \
      http://localhost:8080/health.v1.HealthService/Check
    ```
- **Notes**: 
  - The service maps internal status values to the protobuf enum:
    - "healthy" → `SERVING_STATUS_SERVING`
    - "unhealthy" → `SERVING_STATUS_NOT_SERVING`
    - other values → `SERVING_STATUS_UNKNOWN_UNSPECIFIED`
  - This endpoint supports both gRPC and HTTP/JSON protocols with the same implementation.

## Using Connect-go for API Access

The application uses Connect-go, which provides a unified way to access APIs via both gRPC and HTTP/JSON:

1. **gRPC Protocol**: Efficient binary protocol ideal for service-to-service communication
2. **HTTP/JSON Protocol**: REST-like interface for browser clients or debugging

### JSON Serialization Notes

When using the HTTP/JSON protocol with Connect-go, be aware of the following serialization behaviors:

- **Enum Values**: Serialized using their string representation (e.g., `"SERVING_STATUS_SERVING"`)
- **Empty Messages**: Represented as empty JSON objects (`{}`)
- **Field Names**: Uses JSON camelCase naming convention (e.g., `servingStatus` instead of `serving_status`)

## API Development

New APIs should be defined using Protocol Buffers (`.proto` files) in the `proto` directory. The build system will automatically generate:

1. Go code for the messages and services
2. Connect-go handlers and clients
3. OpenAPI specifications

To generate code from proto definitions:

```bash
make buf-generate
```