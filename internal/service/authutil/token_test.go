package authutil

import (
	"errors"
	"testing"
)

func TestExtractBearerToken(t *testing.T) {
	testCases := []struct {
		name          string
		authHeader    string
		expectedToken string
		expectedErr   error
	}{
		{
			name:          "Success",
			authHeader:    "Bearer abc.def.ghi",
			expectedToken: "abc.def.ghi",
			expectedErr:   nil,
		},
		{
			name:          "Success with case-insensitive scheme",
			authHeader:    "bearer abc.def.ghi",
			expectedToken: "abc.def.ghi",
			expectedErr:   nil,
		},
		{
			name:          "Success with extra space",
			authHeader:    "Bearer   abc.def.ghi",
			expectedToken: "abc.def.ghi",
			expectedErr:   nil,
		},
		{
			name:        "Missing header",
			authHeader:  "",
			expectedErr: ErrMissingHeader,
		},
		{
			name:        "Whitespace header",
			authHeader:  "   ",
			expectedErr: ErrMissingHeader,
		},
		{
			name:        "Invalid prefix",
			authHeader:  "Token abc",
			expectedErr: ErrInvalidFormat,
		},
		{
			name:        "Empty token",
			authHeader:  "Bearer   ",
			expectedErr: ErrEmptyToken,
		},
		{
			name:        "Header with only Bearer",
			authHeader:  "Bearer",
			expectedErr: ErrInvalidFormat,
		},
		{
			name:          "Token contains spaces (compat)",
			authHeader:    "Bearer abc def ghi",
			expectedToken: "abc def ghi",
			expectedErr:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := ExtractBearerToken(tc.authHeader)
			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, but got: %v", err)
				}
				if token != tc.expectedToken {
					t.Errorf("expected token '%s', but got '%s'", tc.expectedToken, token)
				}
			}
		})
	}
}
