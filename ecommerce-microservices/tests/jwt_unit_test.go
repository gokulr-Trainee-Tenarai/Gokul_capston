package tests

import (
	"testing"

	"ecommerce-microservices/pkg/auth"
)

func TestJWT_GenerateAndValidate(t *testing.T) {
	token, err := auth.GenerateToken("testuser", "USER")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate valid token: %v", err)
	}

	if claims.Username != "testuser" || claims.Role != "USER" {
		t.Errorf("unexpected claims: %+v", claims)
	}
}

func TestJWT_InvalidToken(t *testing.T) {
	_, err := auth.ValidateToken("invalid.token.string")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}
