# E2E Authentication Tests with Godog

## Overview

This document describes the end-to-end (E2E) authentication tests implemented using Godog for the Sudal Social Quiz Platform. These tests validate the complete user authentication and management lifecycle using real Firebase authentication.

## Architecture

### Components

1. **Firebase REST API Client** (`/test/e2e/helpers/firebase_auth_client.go`)
   - Interacts with Firebase Auth REST API
   - Creates and deletes test users
   - Obtains real Firebase ID tokens

2. **User Authentication Feature** (`/test/e2e/features/user/user_auth.feature`)
   - Gherkin scenarios for authentication testing
   - Covers positive and negative test cases
   - Tests real-world user flows

3. **User Authentication Steps** (`/test/e2e/steps/user_auth_steps.go`)
   - Step definitions for the feature file
   - Implements the actual test logic
   - Manages test state and cleanup

## Test Scenarios

### 1. Successful Registration and Profile Retrieval
- Creates a new Firebase user with email/password
- Registers the user with our gRPC service using the Firebase ID token
- Retrieves the user profile using authentication
- Validates that all data matches

### 2. Duplicate Registration Handling
- Creates and registers a user
- Attempts to register again with the same Firebase credentials
- Validates that no duplicate user is created
- Ensures the same user ID is returned

### 3. Invalid Token Authentication
- Attempts to access protected resources with an invalid token
- Validates that the request fails with unauthenticated error

### 4. Missing Token Authentication
- Attempts to access protected resources without any token
- Validates that the request fails with unauthenticated error

## Configuration

### Environment Variables

```bash
# Firebase Web API Key (required for E2E tests)
FIREBASE_WEB_API_KEY=your-firebase-web-api-key

# Firebase Admin SDK credentials (already configured)
GOOGLE_APPLICATION_CREDENTIALS=secrets/firebase_admin_key.json

# Test server endpoints
BASE_URL=http://localhost:8080
GRPC_ADDR=localhost:8080
```

### Firebase Web API Key

The `FIREBASE_WEB_API_KEY` is different from the Firebase Admin SDK credentials:

1. **Admin SDK Key**: Used by the server for token verification (server-side)
2. **Web API Key**: Used by clients for authentication (client-side)

To find your Firebase Web API Key:
1. Go to Firebase Console → Project Settings
2. Under "General" tab, find "Web API Key"
3. Copy the key and set it in your `.env` file

## Test Execution

### Running Authentication Tests

```bash
# Run all E2E tests (includes authentication tests)
make test.e2e

# Run only user authentication tests
go test ./test/e2e/... -godog.tags="@user_auth"

# Run with specific format
go test ./test/e2e/... -godog.format=pretty
```

### Test Environment Setup

1. **Start the server:**
   ```bash
   APP_ENV=dev GOOGLE_APPLICATION_CREDENTIALS=./secrets/firebase_admin_key.json go run cmd/server/main.go
   ```

2. **Set environment variables:**
   ```bash
   export FIREBASE_WEB_API_KEY=your-firebase-web-api-key
   export BASE_URL=http://localhost:8080
   export GRPC_ADDR=localhost:8080
   ```

3. **Run the tests:**
   ```bash
   make test.e2e
   ```

## Test Flow

### Authentication Test Flow

```
1. Generate random email/password
2. Sign up with Firebase Auth REST API
3. Obtain Firebase ID token
4. Register user with our gRPC service (using ID token)
5. Test protected operations (GetUserProfile, UpdateUserProfile)
6. Cleanup: Delete Firebase user
```

### Error Testing Flow

```
1. Create user and obtain valid token
2. Test with invalid token → Expect unauthenticated error
3. Test without token → Expect unauthenticated error
4. Cleanup: Delete Firebase user
```

## Test Isolation and Cleanup

### User Isolation
- Each test scenario uses a unique email address
- Email format: `test-user-{timestamp}@example.com`
- Prevents conflicts between test runs

### Cleanup Strategy
- Firebase users are deleted after each scenario
- Uses `godog.ScenarioContext.After` hook for cleanup
- Ensures tests are idempotent

### Error Handling
- Cleanup failures are logged but don't fail tests
- Graceful degradation for network issues
- Timeout protection for all operations

## Security Considerations

### Test Data
- Uses disposable email addresses
- Generates secure random passwords
- No real user data in tests

### Token Management
- ID tokens are short-lived (1 hour)
- Tokens are not logged or persisted
- Proper cleanup prevents token leakage

### Firebase Project
- Use a dedicated Firebase project for testing
- Separate from production environment
- Configure appropriate security rules

## Troubleshooting

### Common Issues

1. **Missing FIREBASE_WEB_API_KEY**
   ```
   Error: FIREBASE_WEB_API_KEY environment variable is required
   Solution: Set the Firebase Web API key in .env file
   ```

2. **Firebase Authentication Disabled**
   ```
   Error: Firebase Auth error: ADMIN_ONLY_OPERATION
   Solution: Enable Email/Password authentication in Firebase Console
   ```

3. **Server Not Running**
   ```
   Error: failed to make request: connection refused
   Solution: Start the server before running tests
   ```

4. **Invalid Credentials**
   ```
   Error: failed to initialize Firebase app
   Solution: Check GOOGLE_APPLICATION_CREDENTIALS path and file
   ```

### Debug Mode

Enable debug logging for detailed test output:

```bash
LOG_LEVEL=debug go test ./test/e2e/... -godog.format=pretty
```

## Integration with CI/CD

### GitHub Actions Example

```yaml
- name: Set up Firebase credentials
  env:
    FIREBASE_WEB_API_KEY: ${{ secrets.FIREBASE_WEB_API_KEY }}
    GOOGLE_APPLICATION_CREDENTIALS: ./secrets/firebase_admin_key.json
  run: |
    echo "${{ secrets.FIREBASE_ADMIN_KEY }}" > ./secrets/firebase_admin_key.json

- name: Run E2E Authentication Tests
  run: make test.e2e
```

### Required Secrets
- `FIREBASE_WEB_API_KEY`: Firebase Web API key
- `FIREBASE_ADMIN_KEY`: Firebase Admin SDK JSON key (base64 encoded)

## Future Enhancements

### Additional Test Scenarios
- OAuth provider authentication (Google, Apple)
- Token refresh testing
- Rate limiting validation
- Concurrent authentication testing

### Test Improvements
- Parallel test execution
- Test data factories
- Custom assertions
- Performance benchmarks

### Monitoring
- Test execution metrics
- Firebase quota monitoring
- Error rate tracking
- Test duration analysis
