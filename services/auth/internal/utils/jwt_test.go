package utils

import (
	"strings"
	"testing"
	"time"
)

func TestNewJWTManager(t *testing.T) {
	manager := NewJWTManager(
		"test-access-secret",
		"test-refresh-secret",
		15*time.Minute,
		720*time.Hour,
		"test-issuer",
	)

	if manager == nil {
		t.Fatal("NewJWTManager() returned nil")
	}
}

func TestGenerateAccessToken(t *testing.T) {
	manager := NewJWTManager(
		"test-access-secret-1234567890",
		"test-refresh-secret-1234567890",
		15*time.Minute,
		720*time.Hour,
		"ikampus-auth-test",
	)

	tests := []struct {
		name      string
		userID    string
		username  string
		email     string
		role      string
		wantError bool
	}{
		{
			name:      "Valid token generation",
			userID:    "user_123",
			username:  "testuser",
			email:     "test@ox.ac.uk",
			role:      "student",
			wantError: false,
		},
		{
			name:      "Empty user ID",
			userID:    "",
			username:  "testuser",
			email:     "test@ox.ac.uk",
			role:      "student",
			wantError: false, // JWT allows empty fields
		},
		{
			name:      "Admin role",
			userID:    "user_456",
			username:  "admin",
			email:     "admin@ox.ac.uk",
			role:      "admin",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, expiresAt, err := manager.GenerateAccessToken(
				tt.userID,
				tt.username,
				tt.email,
				tt.role,
			)

			if tt.wantError {
				if err == nil {
					t.Errorf("GenerateAccessToken() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("GenerateAccessToken() unexpected error: %v", err)
				}
				if token == "" {
					t.Errorf("GenerateAccessToken() returned empty token")
				}
				if expiresAt.IsZero() {
					t.Errorf("GenerateAccessToken() returned zero expiry time")
				}
				if expiresAt.Before(time.Now()) {
					t.Errorf("GenerateAccessToken() returned expired token")
				}
				// Check token has three parts (header.payload.signature)
				parts := strings.Split(token, ".")
				if len(parts) != 3 {
					t.Errorf("GenerateAccessToken() token has %d parts, want 3", len(parts))
				}
			}
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	manager := NewJWTManager(
		"test-access-secret-1234567890",
		"test-refresh-secret-1234567890",
		15*time.Minute,
		720*time.Hour,
		"ikampus-auth-test",
	)

	tests := []struct {
		name      string
		userID    string
		wantError bool
	}{
		{
			name:      "Valid refresh token",
			userID:    "user_123",
			wantError: false,
		},
		{
			name:      "Empty user ID",
			userID:    "",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, expiresAt, err := manager.GenerateRefreshToken(tt.userID)

			if tt.wantError {
				if err == nil {
					t.Errorf("GenerateRefreshToken() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("GenerateRefreshToken() unexpected error: %v", err)
				}
				if token == "" {
					t.Errorf("GenerateRefreshToken() returned empty token")
				}
				if expiresAt.IsZero() {
					t.Errorf("GenerateRefreshToken() returned zero expiry time")
				}
				// Refresh token should have longer expiry
				if expiresAt.Before(time.Now().Add(24 * time.Hour)) {
					t.Errorf("GenerateRefreshToken() expiry too short")
				}
			}
		})
	}
}

func TestValidateAccessToken(t *testing.T) {
	manager := NewJWTManager(
		"test-access-secret-1234567890",
		"test-refresh-secret-1234567890",
		15*time.Minute,
		720*time.Hour,
		"ikampus-auth-test",
	)

	// Generate a valid token
	validToken, _, err := manager.GenerateAccessToken(
		"user_123",
		"testuser",
		"test@ox.ac.uk",
		"student",
	)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	tests := []struct {
		name      string
		token     string
		wantError bool
	}{
		{
			name:      "Valid token",
			token:     validToken,
			wantError: false,
		},
		{
			name:      "Empty token",
			token:     "",
			wantError: true,
		},
		{
			name:      "Invalid token format",
			token:     "invalid.token.format",
			wantError: true,
		},
		{
			name:      "Malformed token",
			token:     "notavalidtoken",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := manager.ValidateAccessToken(tt.token)

			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateAccessToken() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateAccessToken() unexpected error: %v", err)
				}
				if claims == nil {
					t.Errorf("ValidateAccessToken() returned nil claims")
				}
				if claims.UserID != "user_123" {
					t.Errorf("ValidateAccessToken() userID = %v, want user_123", claims.UserID)
				}
				if claims.Username != "testuser" {
					t.Errorf("ValidateAccessToken() username = %v, want testuser", claims.Username)
				}
				if claims.Email != "test@ox.ac.uk" {
					t.Errorf("ValidateAccessToken() email = %v, want test@ox.ac.uk", claims.Email)
				}
				if claims.Role != "student" {
					t.Errorf("ValidateAccessToken() role = %v, want student", claims.Role)
				}
			}
		})
	}
}

func TestValidateRefreshToken(t *testing.T) {
	manager := NewJWTManager(
		"test-access-secret-1234567890",
		"test-refresh-secret-1234567890",
		15*time.Minute,
		720*time.Hour,
		"ikampus-auth-test",
	)

	// Generate a valid refresh token
	validToken, _, err := manager.GenerateRefreshToken("user_123")
	if err != nil {
		t.Fatalf("Failed to generate test refresh token: %v", err)
	}

	tests := []struct {
		name      string
		token     string
		wantError bool
	}{
		{
			name:      "Valid refresh token",
			token:     validToken,
			wantError: false,
		},
		{
			name:      "Empty token",
			token:     "",
			wantError: true,
		},
		{
			name:      "Invalid token",
			token:     "invalid.token",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := manager.ValidateRefreshToken(tt.token)

			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateRefreshToken() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateRefreshToken() unexpected error: %v", err)
				}
				if claims == nil {
					t.Errorf("ValidateRefreshToken() returned nil claims")
				}
				if claims.UserID != "user_123" {
					t.Errorf("ValidateRefreshToken() userID = %v, want user_123", claims.UserID)
				}
			}
		})
	}
}

func TestTokenExpiry(t *testing.T) {
	// Create manager with very short expiry for testing
	manager := NewJWTManager(
		"test-access-secret-1234567890",
		"test-refresh-secret-1234567890",
		1*time.Second,
		2*time.Second,
		"ikampus-auth-test",
	)

	// Generate token
	token, expiresAt, err := manager.GenerateAccessToken(
		"user_123",
		"testuser",
		"test@ox.ac.uk",
		"student",
	)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Token should be valid immediately
	claims, err := manager.ValidateAccessToken(token)
	if err != nil {
		t.Errorf("Token should be valid immediately: %v", err)
	}
	if claims == nil {
		t.Error("Claims should not be nil")
	}

	// Wait for expiry
	time.Sleep(2 * time.Second)

	// Token should be expired
	_, err = manager.ValidateAccessToken(token)
	if err == nil {
		t.Error("Token should be expired")
	}

	// Check expiry time was set correctly
	if expiresAt.After(time.Now()) {
		t.Error("Token should be expired by now")
	}
}

func TestTokenWithDifferentSecrets(t *testing.T) {
	manager1 := NewJWTManager(
		"secret-1",
		"secret-1-refresh",
		15*time.Minute,
		720*time.Hour,
		"issuer-1",
	)

	manager2 := NewJWTManager(
		"secret-2",
		"secret-2-refresh",
		15*time.Minute,
		720*time.Hour,
		"issuer-2",
	)

	// Generate token with manager1
	token, _, err := manager1.GenerateAccessToken(
		"user_123",
		"testuser",
		"test@ox.ac.uk",
		"student",
	)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Should validate with manager1
	_, err = manager1.ValidateAccessToken(token)
	if err != nil {
		t.Errorf("Token should be valid with same manager: %v", err)
	}

	// Should NOT validate with manager2 (different secret)
	_, err = manager2.ValidateAccessToken(token)
	if err == nil {
		t.Error("Token should not be valid with different secret")
	}
}

func BenchmarkGenerateAccessToken(b *testing.B) {
	manager := NewJWTManager(
		"test-access-secret-1234567890",
		"test-refresh-secret-1234567890",
		15*time.Minute,
		720*time.Hour,
		"ikampus-auth-test",
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = manager.GenerateAccessToken(
			"user_123",
			"testuser",
			"test@ox.ac.uk",
			"student",
		)
	}
}

func BenchmarkValidateAccessToken(b *testing.B) {
	manager := NewJWTManager(
		"test-access-secret-1234567890",
		"test-refresh-secret-1234567890",
		15*time.Minute,
		720*time.Hour,
		"ikampus-auth-test",
	)

	token, _, _ := manager.GenerateAccessToken(
		"user_123",
		"testuser",
		"test@ox.ac.uk",
		"student",
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.ValidateAccessToken(token)
	}
}
