package authutil

import "testing"

func TestExtractBearerToken(t *testing.T) {
	testCases := []struct {
		name          string
		authHeader    string
		expectedToken string
		expectError   bool
	}{
		{
			name:          "Success",
			authHeader:    "Bearer abc.def.ghi",
			expectedToken: "abc.def.ghi",
			expectError:   false,
		},
		{
			name:          "Success with case-insensitive scheme",
			authHeader:    "bearer abc.def.ghi",
			expectedToken: "abc.def.ghi",
			expectError:   false,
		},
		{
			name:          "Success with extra space",
			authHeader:    "Bearer   abc.def.ghi",
			expectedToken: "abc.def.ghi",
			expectError:   false,
		},
		{
			name:        "Missing header",
			authHeader:  "",
			expectError: true,
		},
		{
			name:        "Whitespace header",
			authHeader:  "   ",
			expectError: true,
		},
		{
			name:        "Invalid prefix",
			authHeader:  "Token abc",
			expectError: true,
		},
		{
			name:        "Empty token",
			authHeader:  "Bearer   ",
			expectError: true,
		},
		{
			name:        "Header with only Bearer",
			authHeader:  "Bearer",
			expectError: true,
		},
		{
			name:          "Token contains spaces (compat)",
			authHeader:    "Bearer abc def ghi",
			expectedToken: "abc def ghi",
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := ExtractBearerToken(tc.authHeader)
			if tc.expectError {
				if err == nil {
					t.Errorf("expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got: %v", err)
				}
				if token != tc.expectedToken {
					t.Errorf("expected token '%s', but got '%s'", tc.expectedToken, token)
				}
			}
		})
	}
}
