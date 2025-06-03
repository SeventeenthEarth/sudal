package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/seventeenthearth/sudal/internal/feature/user/domain/entity"
	"github.com/seventeenthearth/sudal/internal/infrastructure/firebase"
	"go.uber.org/zap"
)

// UserContextKey is the context key for storing authenticated user
type UserContextKey string

const (
	// AuthenticatedUserKey is the context key for the authenticated user
	AuthenticatedUserKey UserContextKey = "authenticated_user"
)

// AuthenticationInterceptor creates a Connect-go interceptor for Firebase token authentication
// This interceptor performs the following operations:
// 1. Extracts the Bearer token from the Authorization header
// 2. Verifies the token using Firebase Admin SDK
// 3. Retrieves or creates the user in the database
// 4. Injects the authenticated user into the request context
//
// Parameters:
//   - firebaseHandler: Firebase handler for token verification
//   - logger: Structured logger for recording authentication events
//
// Returns:
//   - connect.UnaryInterceptorFunc: Connect-go interceptor function
func AuthenticationInterceptor(firebaseHandler *firebase.FirebaseHandler, logger *zap.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Extract Authorization header
			authHeader := req.Header().Get("Authorization")
			if authHeader == "" {
				logger.Warn("Missing Authorization header",
					zap.String("procedure", req.Spec().Procedure))
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("missing authorization header"))
			}

			// Extract Bearer token
			token, err := extractBearerToken(authHeader)
			if err != nil {
				logger.Warn("Invalid Authorization header format",
					zap.String("procedure", req.Spec().Procedure),
					zap.Error(err))
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid authorization header format"))
			}

			// Verify token and get user
			user, err := firebaseHandler.VerifyIDToken(ctx, token)
			if err != nil {
				logger.Warn("Token verification failed",
					zap.String("procedure", req.Spec().Procedure),
					zap.Error(err))
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authentication failed: %w", err))
			}

			// Add user to context
			ctxWithUser := context.WithValue(ctx, AuthenticatedUserKey, user)

			logger.Info("User authenticated successfully",
				zap.String("procedure", req.Spec().Procedure),
				zap.String("user_id", user.ID.String()),
				zap.String("firebase_uid", user.FirebaseUID))

			// Continue with the authenticated context
			return next(ctxWithUser, req)
		}
	}
}

// SelectiveAuthenticationInterceptor creates a Connect-go interceptor that applies authentication
// only to specified procedures, while allowing others to pass through without authentication
//
// Parameters:
//   - firebaseHandler: Firebase handler for token verification
//   - logger: Structured logger for recording authentication events
//   - protectedProcedures: List of procedure names that require authentication
//
// Returns:
//   - connect.UnaryInterceptorFunc: Connect-go interceptor function
func SelectiveAuthenticationInterceptor(firebaseHandler *firebase.FirebaseHandler, logger *zap.Logger, protectedProcedures []string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			procedure := req.Spec().Procedure

			// Check if this procedure requires authentication
			requiresAuth := false
			for _, protected := range protectedProcedures {
				if procedure == protected {
					requiresAuth = true
					break
				}
			}

			// If authentication is not required, proceed without authentication
			if !requiresAuth {
				logger.Debug("Procedure does not require authentication, proceeding",
					zap.String("procedure", procedure))
				return next(ctx, req)
			}

			// Apply authentication for protected procedures
			logger.Debug("Applying authentication to protected procedure",
				zap.String("procedure", procedure))

			// Extract Authorization header
			authHeader := req.Header().Get("Authorization")
			if authHeader == "" {
				logger.Warn("Missing Authorization header",
					zap.String("procedure", procedure))
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("missing authorization header"))
			}

			// Extract Bearer token
			token, err := extractBearerToken(authHeader)
			if err != nil {
				logger.Warn("Invalid Authorization header format",
					zap.String("procedure", procedure),
					zap.Error(err))
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid authorization header format"))
			}

			// Verify token and get user
			user, err := firebaseHandler.VerifyIDToken(ctx, token)
			if err != nil {
				logger.Warn("Token verification failed",
					zap.String("procedure", procedure),
					zap.Error(err))
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authentication failed: %w", err))
			}

			// Add user to context
			ctxWithUser := context.WithValue(ctx, AuthenticatedUserKey, user)

			logger.Info("User authenticated successfully",
				zap.String("procedure", procedure),
				zap.String("user_id", user.ID.String()),
				zap.String("firebase_uid", user.FirebaseUID))

			// Continue with the authenticated context
			return next(ctxWithUser, req)
		}
	}
}

// extractBearerToken extracts the token from the Authorization header
// Expected format: "Bearer <token>"
//
// Parameters:
//   - authHeader: Authorization header value
//
// Returns:
//   - string: Extracted token
//   - error: Extraction error if format is invalid
func extractBearerToken(authHeader string) (string, error) {
	const bearerPrefix = "Bearer "

	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", fmt.Errorf("authorization header must start with 'Bearer '")
	}

	token := strings.TrimPrefix(authHeader, bearerPrefix)
	token = strings.TrimSpace(token)

	if token == "" {
		return "", fmt.Errorf("bearer token is empty")
	}

	return token, nil
}

// GetAuthenticatedUser retrieves the authenticated user from the request context
// This is a utility function for handlers to easily access the authenticated user
//
// Parameters:
//   - ctx: Request context containing the authenticated user
//
// Returns:
//   - *entity.User: The authenticated user entity
//   - error: Error if user is not found in context
func GetAuthenticatedUser(ctx context.Context) (*entity.User, error) {
	user, ok := ctx.Value(AuthenticatedUserKey).(*entity.User)
	if !ok || user == nil {
		return nil, fmt.Errorf("authenticated user not found in context")
	}
	return user, nil
}

// AuthenticationMiddleware creates an HTTP middleware for Firebase token authentication
// This middleware is designed for HTTP handlers that need authentication
// It follows the same authentication flow as the Connect-go interceptor
//
// Parameters:
//   - firebaseHandler: Firebase handler for token verification
//   - logger: Structured logger for recording authentication events
//
// Returns:
//   - func(http.Handler) http.Handler: HTTP middleware function
func AuthenticationMiddleware(firebaseHandler *firebase.FirebaseHandler, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Warn("Missing Authorization header",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method))
				writeUnauthenticatedError(w, "missing authorization header")
				return
			}

			// Extract Bearer token
			token, err := extractBearerToken(authHeader)
			if err != nil {
				logger.Warn("Invalid Authorization header format",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
					zap.Error(err))
				writeUnauthenticatedError(w, "invalid authorization header format")
				return
			}

			// Verify token and get user
			user, err := firebaseHandler.VerifyIDToken(r.Context(), token)
			if err != nil {
				logger.Warn("Token verification failed",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
					zap.Error(err))
				writeUnauthenticatedError(w, fmt.Sprintf("authentication failed: %v", err))
				return
			}

			// Add user to context
			ctxWithUser := context.WithValue(r.Context(), AuthenticatedUserKey, user)

			logger.Info("User authenticated successfully",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
				zap.String("user_id", user.ID.String()),
				zap.String("firebase_uid", user.FirebaseUID))

			// Continue with the authenticated context
			next.ServeHTTP(w, r.WithContext(ctxWithUser))
		})
	}
}

// writeUnauthenticatedError writes a standardized unauthenticated error response
// This ensures consistent error format for HTTP/JSON clients
//
// Parameters:
//   - w: HTTP response writer
//   - message: Error message to include in response
func writeUnauthenticatedError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	// Write standardized error response as specified in the requirements
	errorResponse := fmt.Sprintf(`{"code":"unauthenticated","message":"%s"}`, message)
	w.Write([]byte(errorResponse))
}
