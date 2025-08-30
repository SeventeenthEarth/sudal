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
		tc := tc // Capture range variable to ensure it's unique for each parallel test.
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			token, err := ExtractBearerToken(tc.authHeader)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("ExtractBearerToken() error = %v, wantErr %v", err, tc.expectedErr)
			}
			if tc.expectedErr == nil && token != tc.expectedToken {
				t.Errorf("ExtractBearerToken() token = %q, want %q", token, tc.expectedToken)
			}
		})
	}
}
