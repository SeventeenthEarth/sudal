package authutil

import (
	"fmt"
	"strings"
)

// ExtractBearerToken extracts a bearer token from an Authorization header.
// Expected format: "Bearer <token>" (case-sensitive prefix per current usage).
// Returns the token string or an error when the header is missing/invalid/empty.
func ExtractBearerToken(authHeader string) (string, error) {
	if strings.TrimSpace(authHeader) == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	// Accept case-insensitive scheme per RFC 6750, while preserving existing messages
	parts := strings.Fields(authHeader)
	if len(parts) < 2 {
		return "", fmt.Errorf("authorization header must start with 'Bearer '")
	}
	if strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("authorization header must start with 'Bearer '")
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", fmt.Errorf("bearer token is empty")
	}
	return token, nil
}
