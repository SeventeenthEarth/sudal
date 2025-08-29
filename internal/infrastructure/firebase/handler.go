package firebase

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/seventeenthearth/sudal/internal/feature/user/domain/entity"
	"github.com/seventeenthearth/sudal/internal/feature/user/domain/repo"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

// AuthVerifier abstracts Firebase ID token verification and user retrieval/creation.
// Production uses FirebaseHandler; tests can mock this interface.
type AuthVerifier interface {
	VerifyIDToken(ctx context.Context, idToken string) (*entity.User, error)
}

// FirebaseHandler handles Firebase Admin SDK operations
// This handler is responsible for initializing the Firebase Admin SDK
// and providing token verification functionality for authentication middleware
type FirebaseHandler struct {
	app        *firebase.App
	authClient *auth.Client
	userRepo   repo.UserRepository
	logger     *zap.Logger
}

// NewFirebaseHandler creates a new Firebase handler instance
// It initializes the Firebase Admin SDK using the provided credentials file
// and sets up the authentication client for token verification
//
// Parameters:
//   - credentialsFile: Path to the Firebase service account credentials JSON file
//   - userRepo: User repository for database operations
//   - logger: Structured logger for recording operations and errors
//
// Returns:
//   - *FirebaseHandler: Initialized Firebase handler
//   - error: Any initialization error
func NewFirebaseHandler(credentialsFile string, userRepo repo.UserRepository, logger *zap.Logger) (*FirebaseHandler, error) {
	ctx := context.Background()

	// Initialize Firebase Admin SDK with service account credentials
	opt := option.WithCredentialsFile(credentialsFile)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		logger.Error("Failed to initialize Firebase app",
			zap.String("credentials_file", credentialsFile),
			zap.Error(err))
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	// Get Firebase Auth client
	authClient, err := app.Auth(ctx)
	if err != nil {
		logger.Error("Failed to get Firebase Auth client", zap.Error(err))
		return nil, fmt.Errorf("failed to get Firebase Auth client: %w", err)
	}

	logger.Info("Firebase Admin SDK initialized successfully",
		zap.String("credentials_file", credentialsFile))

	return &FirebaseHandler{
		app:        app,
		authClient: authClient,
		userRepo:   userRepo,
		logger:     logger,
	}, nil
}

// VerifyIDToken verifies a Firebase ID token and returns the associated user
// This method performs the following operations:
// 1. Verifies the Firebase ID token using Firebase Admin SDK
// 2. Extracts the Firebase UID from the verified token
// 3. Queries the database for an existing user with the Firebase UID
// 4. If user doesn't exist, creates a new user record
// 5. Returns the complete user entity
//
// Parameters:
//   - ctx: Request context for cancellation and tracing
//   - idToken: Firebase ID token to verify
//
// Returns:
//   - *entity.User: The authenticated user entity
//   - error: Authentication or database error
func (h *FirebaseHandler) VerifyIDToken(ctx context.Context, idToken string) (*entity.User, error) {
	// Verify the ID token using Firebase Admin SDK
	token, err := h.authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		h.logger.Warn("Failed to verify Firebase ID token",
			zap.Error(err))
		return nil, fmt.Errorf("invalid or expired ID token: %w", err)
	}

	// Extract Firebase UID from the verified token
	firebaseUID := token.UID
	if firebaseUID == "" {
		h.logger.Error("Firebase UID is empty in verified token")
		return nil, fmt.Errorf("firebase UID is empty in verified token")
	}

	// Try to get existing user by Firebase UID
	user, err := h.userRepo.GetByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		// Check if it's a "user not found" error
		if err == entity.ErrUserNotFound {
			// User doesn't exist, create a new one
			h.logger.Info("User not found, creating new user",
				zap.String("firebase_uid", firebaseUID))

			// Determine auth provider from token claims
			authProvider := h.extractAuthProvider(token)

			// Create new user entity
			newUser := entity.NewUser(firebaseUID, authProvider)

			// Save the new user to database
			createdUser, createErr := h.userRepo.Create(ctx, newUser)
			if createErr != nil {
				h.logger.Error("Failed to create new user",
					zap.String("firebase_uid", firebaseUID),
					zap.String("auth_provider", authProvider),
					zap.Error(createErr))
				return nil, fmt.Errorf("failed to create new user: %w", createErr)
			}

			h.logger.Info("New user created successfully",
				zap.String("user_id", createdUser.ID.String()),
				zap.String("firebase_uid", firebaseUID),
				zap.String("auth_provider", authProvider))

			return createdUser, nil
		}

		// Other database error
		h.logger.Error("Failed to query user by Firebase UID",
			zap.String("firebase_uid", firebaseUID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	// User exists, return it
	h.logger.Debug("User found and authenticated",
		zap.String("user_id", user.ID.String()),
		zap.String("firebase_uid", firebaseUID))

	return user, nil
}

// extractAuthProvider determines the authentication provider from Firebase token claims
// This method examines the token's Firebase claims to identify which OAuth provider
// was used for authentication (Google, Email/Password, etc.)
//
// Parameters:
//   - token: Verified Firebase token containing user claims
//
// Returns:
//   - string: Authentication provider identifier
func (h *FirebaseHandler) extractAuthProvider(token *auth.Token) string {
	// Check Firebase claims for provider information
	if firebase, ok := token.Claims["firebase"].(map[string]interface{}); ok {
		if identities, ok := firebase["identities"].(map[string]interface{}); ok {
			// Check for Google provider
			if _, hasGoogle := identities["google.com"]; hasGoogle {
				return "google"
			}
			// Check for email provider
			if _, hasEmail := identities["email"]; hasEmail {
				return "email"
			}
		}

		// Check sign_in_provider field
		if provider, ok := firebase["sign_in_provider"].(string); ok {
			switch provider {
			case "google.com":
				return "google"
			case "password":
				return "email"
			default:
				return provider
			}
		}
	}

	// Default to email if provider cannot be determined
	h.logger.Warn("Could not determine auth provider from token, defaulting to email")
	return "email"
}
