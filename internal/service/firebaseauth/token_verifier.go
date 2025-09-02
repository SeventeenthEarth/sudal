package firebaseauth

//go:generate go run go.uber.org/mock/mockgen -destination=../../mocks/mock_token_verifier.go -package=mocks github.com/seventeenthearth/sudal/internal/service/firebaseauth TokenVerifier

import (
	"context"
	"fmt"

	"firebase.google.com/go/v4/auth"
	"go.uber.org/zap"
)

// TokenVerifier defines a minimal interface to verify an ID token
// and extract only authentication claims needed by upper layers.
// It must not depend on any internal repositories or domain services.
type TokenVerifier interface {
	// Verify validates the given ID token and returns the Firebase UID
	// and the inferred auth provider (e.g., "google", "email").
	Verify(ctx context.Context, idToken string) (uid string, provider string, err error)
}

// firebaseTokenVerifier is a thin wrapper over Firebase Admin SDK's auth.Client.
type firebaseTokenVerifier struct {
	client *auth.Client
	logger *zap.Logger
}

// NewFirebaseTokenVerifier creates a new TokenVerifier using a Firebase auth client.
func NewFirebaseTokenVerifier(client *auth.Client, logger *zap.Logger) TokenVerifier {
	return &firebaseTokenVerifier{client: client, logger: logger}
}

// Verify implements TokenVerifier using Firebase Admin SDK.
func (v *firebaseTokenVerifier) Verify(ctx context.Context, idToken string) (string, string, error) {
	token, err := v.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		if v.logger != nil {
			v.logger.Warn("Failed to verify Firebase ID token", zap.Error(err))
		}
		return "", "", fmt.Errorf("invalid or expired ID token: %w", err)
	}

	uid := token.UID
	if uid == "" {
		if v.logger != nil {
			v.logger.Error("Firebase UID is empty in verified token")
		}
		return "", "", fmt.Errorf("firebase UID is empty in verified token")
	}

	provider := extractAuthProvider(token, v.logger)
	return uid, provider, nil
}

// extractAuthProvider determines the authentication provider from Firebase token claims.
func extractAuthProvider(token *auth.Token, logger *zap.Logger) string {
	if token == nil {
		return "email"
	}

	if firebase, ok := token.Claims["firebase"].(map[string]interface{}); ok {
		if identities, ok := firebase["identities"].(map[string]interface{}); ok {
			if _, hasGoogle := identities["google.com"]; hasGoogle {
				return "google"
			}
			if _, hasEmail := identities["email"]; hasEmail {
				return "email"
			}
		}

		if p, ok := firebase["sign_in_provider"].(string); ok {
			switch p {
			case "google.com":
				return "google"
			case "password":
				return "email"
			default:
				return p
			}
		}
	}

	if logger != nil {
		logger.Warn("Could not determine auth provider from token, defaulting to email")
	}
	return "email"
}
