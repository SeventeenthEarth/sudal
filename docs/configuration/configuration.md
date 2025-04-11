# Configuration Management

Sudal uses a robust configuration management system that supports different environments (development, staging, production) and multiple configuration sources.

## Recommended Configuration Methods by Environment

| Environment | Primary Method | Secondary Method | Notes |
|-------------|---------------|------------------|-------|
| Development | `.env` file with Docker Compose | `config.yaml` | Simplifies local development |
| Staging | Cloud Run environment variables | - | Easy deployment and testing |
| Production | Cloud Run environment variables + Secret Manager | - | Enhanced security for sensitive data |

## Configuration Sources

The application loads configuration from the following sources, in order of precedence (highest to lowest):

1. **Environment Variables**: Highest priority, overrides other sources
2. **Configuration Files**: YAML, JSON, or TOML files (optional)
3. **Default Values**: Fallback values defined in the code

```mermaid
flowchart TD
    A[Start] --> B{Environment Variables?}
    B -->|Yes| C[Use Environment Variables]
    B -->|No| D{Config File?}
    D -->|Yes| E[Use Config File]
    D -->|No| F[Use Default Values]
    C --> G[Final Configuration]
    E --> G
    F --> G
```

## Environment-Specific Configuration

### Local Development

For local development, we recommend using a `.env` file with Docker Compose. However, you have several options:

1. **Using .env file**:
   - Create a `.env` file in the project root (based on `.env.example`)
   - The application will automatically load variables from this file
   - Required variables are clearly marked in `.env.example` with `[REQUIRED]` tags
   - Optional variables are marked with `[OPTIONAL]` tags
   - Example:
     ```
     SERVER_PORT=8080
     LOG_LEVEL=debug
     ENVIRONMENT=development
     ```

2. **Using config.yaml**:
   - Create a `configs/config.yaml` file (or copy from the example)
   - Run the application with `--config` flag: `./bin/server --config=./configs/config.yaml`
   - Example:
     ```yaml
     server_port: 8080
     log_level: debug
     environment: development
     ```

3. **Using Docker Compose**:
   - Create a `.env` file in the project root
   - Docker Compose will automatically load variables from this file
   - Container-specific settings are overridden in `docker-compose.yml`
   - Run with: `docker-compose up`

### Staging and Production Environments

For staging and production environments on Cloud Run, follow these guidelines:

## Cloud Run Deployment

1. **Basic Deployment with Environment Variables**:

   ```bash
   # For staging
   gcloud run deploy sudal-server-staging \
     --image gcr.io/your-project/sudal:staging \
     --set-env-vars="ENVIRONMENT=staging,LOG_LEVEL=info" \
     --region us-central1
   
   # For production
   gcloud run deploy sudal-server \
     --image gcr.io/your-project/sudal:production \
     --set-env-vars="ENVIRONMENT=production,LOG_LEVEL=info" \
     --region us-central1
   ```

   Note: Cloud Run automatically sets the `PORT` environment variable, so you don't need to specify `SERVER_PORT`.

2. **Using Secret Manager for Sensitive Information**:

   First, create secrets in Secret Manager:

   ```bash
   # Create secrets
   echo -n "postgres://user:password@host:5432/dbname" | \
     gcloud secrets create postgres-dsn --data-file=-
   
   echo -n "redis-host:6379" | \
     gcloud secrets create redis-addr --data-file=-
   
   echo -n "your-super-secret-key" | \
     gcloud secrets create jwt-secret --data-file=-
   ```

   Then, reference them in your Cloud Run deployment:

   ```bash
   gcloud run deploy sudal-server \
     --image gcr.io/your-project/sudal:production \
     --set-env-vars="ENVIRONMENT=production,LOG_LEVEL=info" \
     --update-secrets="POSTGRES_DSN=postgres-dsn:latest,REDIS_ADDR=redis-addr:latest,JWT_SECRET_KEY=jwt-secret:latest" \
     --region us-central1
   ```

3. **Mounting Firebase Credentials**:

   Store your Firebase credentials in Secret Manager:

   ```bash
   # Upload Firebase credentials to Secret Manager
   gcloud secrets create firebase-credentials \
     --data-file=./path/to/firebase-credentials.json
   ```

   Mount them as a volume in Cloud Run:

   ```bash
   gcloud run deploy sudal-server \
     --image gcr.io/your-project/sudal:production \
     --set-env-vars="ENVIRONMENT=production,FIREBASE_PROJECT_ID=your-project-id,FIREBASE_CREDENTIALS_JSON=/secrets/firebase-credentials.json" \
     --update-secrets="/secrets/firebase-credentials.json=firebase-credentials:latest" \
     --region us-central1
   ```

## Complete Deployment Example

Here's a complete example for production deployment:

```bash
gcloud run deploy sudal-server \
  --image gcr.io/your-project/sudal:production \
  --set-env-vars="ENVIRONMENT=production,LOG_LEVEL=info,FIREBASE_PROJECT_ID=your-project-id,FIREBASE_CREDENTIALS_JSON=/secrets/firebase-credentials.json" \
  --update-secrets="POSTGRES_DSN=postgres-dsn:latest,REDIS_ADDR=redis-addr:latest,REDIS_PASSWORD=redis-password:latest,JWT_SECRET_KEY=jwt-secret:latest,/secrets/firebase-credentials.json=firebase-credentials:latest" \
  --region us-central1 \
  --allow-unauthenticated
```

## Required Configuration Parameters

| Parameter | Description | Default | Required In | Notes |
|-----------|-------------|---------|-------------|-------|
| `SERVER_PORT` | HTTP server port | 8080 | All environments | Cloud Run provides `PORT` automatically |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | info | All environments | Use `info` or `error` in production |
| `ENVIRONMENT` | Deployment environment | development | All environments | Set to `production` in production |
| `POSTGRES_DSN` | PostgreSQL connection string | - | All environments | Use Secret Manager in production |
| `REDIS_ADDR` | Redis server address | - | All environments | Use Secret Manager in production |
| `REDIS_PASSWORD` | Redis password | - | If Redis requires auth | Use Secret Manager in production |
| `FIREBASE_PROJECT_ID` | Firebase project ID | - | All environments | - |
| `FIREBASE_CREDENTIALS_JSON` | Path to Firebase credentials JSON | - | All environments | Mount as volume in Cloud Run |
| `JWT_SECRET_KEY` | Secret key for JWT signing | - | All environments | Use Secret Manager in production |

Note: For database and Redis configuration, you can either set the DSN/ADDR directly or use the individual component variables (e.g., `DB_HOST`, `DB_PORT`, etc.).

## Configuration Flow

```mermaid
sequenceDiagram
    participant App as Application
    participant Env as Environment Variables
    participant File as Config File
    participant Defaults as Default Values
    
    App->>Env: Load environment variables
    App->>File: Load config file (if specified)
    App->>Defaults: Apply default values
    
    Note over App,Defaults: Environment variables override config file values
    Note over App,Defaults: Config file values override defaults
    
    App->>App: Validate configuration
    App->>App: Start application with final config
```
