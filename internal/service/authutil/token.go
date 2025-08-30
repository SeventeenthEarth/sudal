package authutil

import (
	"errors"
	"strings"
)

var (
	// ErrMissingHeader is returned when the Authorization header is missing.
	ErrMissingHeader = errors.New("missing authorization header")
	// ErrInvalidFormat is returned when the Authorization header doesn't start with the Bearer scheme.
	ErrInvalidFormat = errors.New("authorization header format must be 'Bearer <token>'")
	// ErrEmptyToken is returned when the bearer token is empty.
	ErrEmptyToken = errors.New("bearer token is empty")
)

// ExtractBearerToken extracts a bearer token from an Authorization header.
// It expects the format "Bearer <token>", with a case-insensitive "Bearer" prefix.
// Returns the token string or an error when the header is missing, invalid, or the token is empty.
func ExtractBearerToken(authHeader string) (string, error) {
	if strings.TrimSpace(authHeader) == "" {
		return "", ErrMissingHeader
	}

	// Accept case-insensitive scheme per RFC 6750, while preserving existing messages
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", ErrInvalidFormat
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", ErrEmptyToken
	}
	return token, nil
}
