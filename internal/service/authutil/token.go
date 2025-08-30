package authutil

import (
	"fmt"
	"strings"
)

// ExtractBearerToken extracts a bearer token from an Authorization header.
// It expects the format "Bearer <token>", with a case-insensitive "Bearer" prefix.
// Returns the token string or an error when the header is missing, invalid, or the token is empty.
func ExtractBearerToken(authHeader string) (string, error) {
	if strings.TrimSpace(authHeader) == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	// Accept case-insensitive scheme per RFC 6750, while preserving existing messages
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", fmt.Errorf("authorization header must be in 'Bearer <token>' format")
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", fmt.Errorf("bearer token is empty")
	}
	return token, nil
}
