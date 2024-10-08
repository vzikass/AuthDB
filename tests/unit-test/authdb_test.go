// Simple tests of my project functionality
package unittest

import (
	"AuthDB/utils"
	"testing"
	"time"
)

// Utils tests
func TestCompareHashPassword(t *testing.T) {
	password := "secretpassword"

	hashedPassword, err := utils.GenerateHash(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	if !utils.CompareHashPassword(password, hashedPassword) {
		t.Errorf("CompareHashPassword failed for valid password")
	}

	wrongPassword := "wrongpassword"
	if utils.CompareHashPassword(wrongPassword, hashedPassword) {
		t.Errorf("CompareHashPassword succeeded for invalid password")
	}
}

// Test jwt token
func TestGenerateJWT(t *testing.T) {
	userID := "testUser"

	token, err := utils.GenerateJWT(userID)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	if len(token) == 0 {
		t.Errorf("Generated JWT is empty")
	}
}

func TestParseJWT(t *testing.T) {
	userID := "testUser"
	tokenString, err := utils.GenerateJWT(userID)

	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	token, claims, err := utils.ParseJWT(tokenString)
	if err != nil {
		t.Fatalf("Failed to ParseJWT: %v", err)
	}

	if !token.Valid {
		t.Errorf("Token is not valid: %v", err)
	}

	if (*claims)["userID"] != userID {
		t.Errorf("Expected userID %v, got %v", userID, (*claims)["userID"])
	}

	exp, ok := (*claims)["exp"].(float64)
	if !ok || time.Now().Unix() > int64(exp) {
		t.Errorf("JWT token has expired")
	}
}
