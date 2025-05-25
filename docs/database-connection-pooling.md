# PostgreSQL Connection Pooling Implementation

## Overview

This document describes the PostgreSQL connection pooling implementation for the Sudal Social Quiz Platform backend. The implementation provides configurable connection pooling, SSL/TLS support, and comprehensive connectivity verification.

## Implementation Details

### 1. Configuration System

**File**: `internal/infrastructure/config/config.go`

Enhanced the existing `DBConfig` struct with:

- **Connection Pool Parameters**:
  - `DB_MAX_OPEN_CONNS` (default: 25) - Maximum number of open connections
  - `DB_MAX_IDLE_CONNS` (default: 5) - Maximum number of idle connections
  - `DB_CONN_MAX_LIFETIME_SECONDS` (default: 3600) - Connection maximum lifetime
  - `DB_CONN_MAX_IDLE_TIME_SECONDS` (default: 300) - Connection maximum idle time
  - `DB_CONNECT_TIMEOUT_SECONDS` (default: 30) - Connection timeout

- **SSL/TLS Configuration**:
  - `DB_SSL_CERT` - Client certificate file path
  - `DB_SSL_KEY` - Client private key file path
  - `DB_SSL_ROOT_CERT` - Root certificate file path

### 2. Database Manager

**File**: `internal/infrastructure/database/postgres.go`

Implemented `PostgresManager` with:

- **Connection Pool Management**: Uses Go's standard `database/sql` package with configurable pool settings
- **SSL/TLS Support**: Configurable SSL modes and certificate-based authentication
- **Health Checking**: Comprehensive health checks with connection pool statistics
- **Structured Logging**: Integration with existing zap logger for all operations

Key methods:

- `NewPostgresManager(cfg *config.Config)` - Creates and configures connection pool
- `Ping(ctx context.Context)` - Basic connectivity check
- `HealthCheck(ctx context.Context)` - Comprehensive health check with statistics
- `GetDB()` - Returns underlying `*sql.DB` for direct access
- `Close()` - Properly closes connection pool

### 3. Utility Functions

**File**: `internal/infrastructure/database/utils.go`

Provides utility functions for:

- `VerifyDatabaseConnectivity()` - Standalone connectivity verification
- `GetConnectionPoolStats()` - Retrieve current pool statistics
- `LogConnectionPoolStats()` - Log pool statistics for monitoring

### 4. Dependency Injection

**File**: `internal/infrastructure/di/wire.go`

Added Wire providers:

- `DatabaseSet` - Wire provider set for database dependencies
- `ProvidePostgresManager()` - Provider function for PostgreSQL manager
- `InitializePostgresManager()` - Wire-generated initialization function

### 5. Environment Configuration

**File**: `.env.template`

Updated with new environment variables for:

- Connection pool configuration
- SSL/TLS certificate paths
- Connection timeout settings

## Environment Variables

### Required (Always)

- `APP_ENV` - Application environment (dev, canary, production)
- `SERVER_PORT` - Server port (default: 8080)
- `LOG_LEVEL` - Log level (debug, info, warn, error)

### Database Configuration (Required for production)

- `DB_HOST` - Database host
- `DB_USER` - Database username
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name

### Database Configuration (Optional with defaults)

- `DB_PORT=5432` - Database port
- `DB_SSLMODE=disable` - SSL mode (disable, require, verify-ca, verify-full)
- `POSTGRES_DSN` - Alternative to individual DB components

### Connection Pool Configuration (Optional with defaults)

- `DB_MAX_OPEN_CONNS=25` - Maximum open connections
- `DB_MAX_IDLE_CONNS=5` - Maximum idle connections
- `DB_CONN_MAX_LIFETIME_SECONDS=3600` - Connection lifetime (1 hour)
- `DB_CONN_MAX_IDLE_TIME_SECONDS=300` - Connection idle time (5 minutes)
- `DB_CONNECT_TIMEOUT_SECONDS=30` - Connection timeout

### SSL/TLS Configuration (Optional)

- `DB_SSL_CERT` - Client certificate file path
- `DB_SSL_KEY` - Client private key file path
- `DB_SSL_ROOT_CERT` - Root certificate file path

### Note on Other Services

Redis, Firebase, and JWT configurations are defined in the config structure but are not currently implemented or required. They will be activated in future development phases.

## Usage Examples

### Basic Initialization

```go
import (
    "context"
    "time"
    "github.com/seventeenthearth/sudal/internal/infrastructure/config"
    "github.com/seventeenthearth/sudal/internal/infrastructure/database"
)

// Load configuration
cfg, err := config.LoadConfig("")
if err != nil {
    log.Fatal("Failed to load config", zap.Error(err))
}

// Create database manager
dbManager, err := database.NewPostgresManager(cfg)
if err != nil {
    log.Fatal("Failed to create database manager", zap.Error(err))
}
defer dbManager.Close()
```

### Connectivity Verification

```go
// Verify connectivity at startup
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := database.VerifyDatabaseConnectivity(ctx, cfg); err != nil {
    log.Error("Database connectivity verification failed", zap.Error(err))
    os.Exit(1)
}
```

### Health Check

```go
// Perform health check
healthStatus, err := dbManager.HealthCheck(ctx)
if err != nil {
    log.Error("Health check failed", zap.Error(err))
    return
}

log.Info("Database health check",
    zap.String("status", healthStatus.Status),
    zap.Any("stats", healthStatus.Stats),
)
```

## Integration with Dependency Injection

```go
import "github.com/seventeenthearth/sudal/internal/infrastructure/di"

// Initialize using Wire
dbManager, err := di.InitializePostgresManager()
if err != nil {
    log.Fatal("Failed to initialize database manager", zap.Error(err))
}
defer dbManager.Close()
```

## Logging

All database operations are logged using the existing structured logging system:

- **Info Level**: Successful operations, configuration details, health check results
- **Error Level**: Connection failures, health check failures, configuration errors
- **Debug Level**: Detailed connection pool statistics, ping operations

## Connection Pool Monitoring

The implementation provides detailed connection pool statistics:

- Maximum open connections
- Current open connections
- Connections in use
- Idle connections
- Wait count and duration
- Connections closed due to max idle/lifetime limits

## Security Features

- **SSL/TLS Support**: Configurable SSL modes and certificate-based authentication
- **Connection Timeouts**: Prevents hanging connections
- **Credential Management**: Secure handling of database credentials via environment variables
- **Connection Limits**: Prevents resource exhaustion through configurable limits

## Next Steps

To complete the integration:

1. Uncomment the database initialization code in `cmd/server/main.go`
2. Add the required imports
3. Generate Wire code: `make wire-gen`
4. Set up environment variables for your database
5. Test the connectivity verification

## Manual Testing

1. Set environment variables in `.env`:

```bash
DB_HOST=localhost
DB_USER=user
DB_PASSWORD=password
DB_NAME=quizapp_db
DB_MAX_OPEN_CONNS=10
DB_CONNECT_TIMEOUT_SECONDS=5
```

2. Start the application and observe logs for:
   - Successful connection pool initialization
   - Database connectivity verification
   - Connection pool statistics

3. Test error scenarios by:
   - Setting incorrect database credentials
   - Stopping the database server
   - Setting very low timeout values
