package authutil

import "testing"

func TestExtractBearerToken_Success(t *testing.T) {
	token, err := ExtractBearerToken("Bearer abc.def.ghi")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token != "abc.def.ghi" {
		t.Fatalf("unexpected token: %s", token)
	}
}

func TestExtractBearerToken_Success_CaseInsensitiveScheme(t *testing.T) {
	token, err := ExtractBearerToken("bearer abc.def.ghi")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token != "abc.def.ghi" {
		t.Fatalf("unexpected token: %s", token)
	}
}

func TestExtractBearerToken_MissingHeader(t *testing.T) {
	if _, err := ExtractBearerToken(""); err == nil {
		t.Fatalf("expected error for missing header")
	}
}

func TestExtractBearerToken_InvalidPrefix(t *testing.T) {
	if _, err := ExtractBearerToken("Token abc"); err == nil {
		t.Fatalf("expected error for invalid prefix")
	}
}

func TestExtractBearerToken_EmptyToken(t *testing.T) {
	if _, err := ExtractBearerToken("Bearer   "); err == nil {
		t.Fatalf("expected error for empty token")
	}
}
