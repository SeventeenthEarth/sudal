# Firebase Authentication Implementation

This document describes the Firebase Admin SDK integration and token verification middleware implemented for the Sudal Social Quiz Platform.

## Overview

The Firebase authentication system provides secure token-based authentication for gRPC services using Firebase ID tokens. The implementation follows clean architecture principles and integrates seamlessly with the existing Connect-go server infrastructure.

## Architecture

### Components

1. **Firebase Handler** (`/internal/infrastructure/firebase/handler.go`)
   - Initializes Firebase Admin SDK
   - Verifies Firebase ID tokens
   - Manages user creation and retrieval
   - Handles authentication provider detection

2. **Authentication Middleware** (`/internal/infrastructure/middleware/auth.go`)
   - Connect-go interceptor for gRPC services
   - HTTP middleware for REST endpoints
   - Token extraction and validation
   - User context injection

3. **Dependency Injection** (`/internal/infrastructure/di/wire.go`)
   - Wire provider sets for Firebase components
   - Configuration management
   - Service initialization

## Configuration

### Environment Variables

```bash
# Firebase Admin SDK credentials
GOOGLE_APPLICATION_CREDENTIALS=./secrets/firebase_admin_key.json

# Application environment (required)
APP_ENV=dev
```

### Configuration File (`configs/config.yaml`)

```yaml
# Firebase Configuration
firebase_project_id: sudal-14497
firebase_credentials_json: ./secrets/firebase_admin_key.json
```

## Authentication Flow

### Optimized Authentication Architecture

The implementation uses **selective authentication** to eliminate duplicate logic and improve efficiency:

#### **Before (Problematic)**
```
RegisterUser: Auth Middleware → Create User → Handler → Create User Again ❌ (Duplicate)
GetUserProfile: Auth Middleware → Retrieve User → Handler → Query DB Again ❌ (Inefficient)
UpdateUserProfile: Auth Middleware → Retrieve User → Handler → Query DB Again ❌ (Inefficient)
```

#### **After (Optimized)**
```
RegisterUser: Direct Firebase Verification → Create User ✅ (No middleware)
GetUserProfile: Auth Middleware → Retrieve User → Handler uses Context ✅ (Efficient)
UpdateUserProfile: Auth Middleware → Retrieve User → Handler uses Context ✅ (Efficient)
```

### Key Improvements

1. **Selective Authentication**: Only `GetUserProfile` and `UpdateUserProfile` use authentication middleware
2. **Context-Based User Access**: Authenticated endpoints get user from context (no DB re-query)
3. **Direct Firebase Verification**: `RegisterUser` handles Firebase token verification directly
4. **Permission Validation**: Users can only access/modify their own profiles

### 1. Token Verification Process

```
Client Request → Extract Bearer Token → Verify with Firebase → Query/Create User → Inject User Context → Continue Request
```

### 2. User Management

- **Existing Users**: Retrieved from database using Firebase UID
- **New Users**: Automatically created with Firebase UID and auth provider
- **Auth Providers**: Supports Google and Email/Password authentication

### 3. Error Handling

- **Missing Token**: Returns `CodeUnauthenticated` with "missing authorization header"
- **Invalid Token**: Returns `CodeUnauthenticated` with "invalid or expired ID token"
- **Database Errors**: Returns `CodeUnauthenticated` with "failed to query user"

## Usage

### Protected gRPC Services

The UserService is automatically protected with authentication middleware:

```go
// Server automatically applies authentication to UserService
userPath, userHTTPHandler := userv1connect.NewUserServiceHandler(
    userConnectHandler,
    connect.WithInterceptors(authInterceptor),
)
```

### Accessing Authenticated User

In gRPC handlers, retrieve the authenticated user from context:

```go
import "github.com/seventeenthearth/sudal/internal/infrastructure/middleware"

func (h *UserHandler) GetUserProfile(ctx context.Context, req *connect.Request[userv1.GetUserProfileRequest]) (*connect.Response[userv1.UserProfile], error) {
    // Get authenticated user from context
    user, err := middleware.GetAuthenticatedUser(ctx)
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    
    // Use user.ID, user.FirebaseUID, etc.
    // ...
}
```

## Security Features

### Token Validation

- Verifies token signature using Firebase Admin SDK
- Checks token expiration and validity
- Validates Firebase project ID and issuer

### User Context

- Complete user entity injected into request context
- Includes user ID, Firebase UID, display name, and other profile data
- Consistent across all authenticated requests

### Error Responses

For HTTP/JSON clients, authentication errors return standardized format:

```json
{
  "code": "unauthenticated",
  "message": "ID token has expired"
}
```

## Supported Authentication Providers

### Google OAuth

- Provider: `"google"`
- Detected from Firebase token claims: `identities.google.com`
- Sign-in provider: `"google.com"`

### Email/Password

- Provider: `"email"`
- Detected from Firebase token claims: `identities.email`
- Sign-in provider: `"password"`

## Testing

### Manual Testing

1. **Missing Authorization Header**
   ```bash
   grpcurl -plaintext localhost:8080 user.v1.UserService/GetUserProfile
   # Expected: Code = Unauthenticated
   ```

2. **Invalid Token**
   ```bash
   grpcurl -plaintext -H "Authorization: Bearer invalid_token" localhost:8080 user.v1.UserService/GetUserProfile
   # Expected: Code = Unauthenticated
   ```

3. **Valid Token (New User)**
   ```bash
   grpcurl -plaintext -H "Authorization: Bearer <valid_firebase_token>" localhost:8080 user.v1.UserService/GetUserProfile
   # Expected: Success, new user created
   ```

4. **Valid Token (Existing User)**
   ```bash
   grpcurl -plaintext -H "Authorization: Bearer <valid_firebase_token>" localhost:8080 user.v1.UserService/GetUserProfile
   # Expected: Success, existing user retrieved
   ```

### Integration with Existing Services

- **Health Service**: Remains unprotected for monitoring
- **OpenAPI/REST**: Continues to work for health endpoints
- **UserService**: All RPCs now require authentication

## Implementation Notes

### Clean Architecture Compliance

- Firebase logic isolated in infrastructure layer
- Domain entities remain pure (no Firebase dependencies)
- Repository pattern maintained for user data access

### Performance Considerations

- Firebase Admin SDK connection pooling
- Efficient user lookup by Firebase UID
- Minimal overhead for token verification

### Logging and Monitoring

- Structured logging for all authentication events
- Security-conscious logging (no sensitive data)
- Request tracing with correlation IDs

## Future Enhancements

- Support for additional OAuth providers (Apple, Facebook)
- Token refresh mechanism
- Role-based access control (RBAC)
- Rate limiting for authentication attempts
