package utils

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantError bool
	}{
		{
			name:      "Valid password",
			password:  "Password123",
			wantError: false,
		},
		{
			name:      "Empty password",
			password:  "",
			wantError: true,
		},
		{
			name:      "Password too short",
			password:  "Pass1",
			wantError: true,
		},
		{
			name:      "Long valid password",
			password:  "ThisIsAVeryLongPassword123456789",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if tt.wantError {
				if err == nil {
					t.Errorf("HashPassword() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("HashPassword() unexpected error: %v", err)
				}
				if hash == "" {
					t.Errorf("HashPassword() returned empty hash")
				}
				if hash == tt.password {
					t.Errorf("HashPassword() returned plaintext password")
				}
				if !strings.HasPrefix(hash, "$2a$") {
					t.Errorf("HashPassword() hash doesn't look like bcrypt")
				}
			}
		})
	}
}

func TestComparePassword(t *testing.T) {
	password := "TestPassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password for test: %v", err)
	}

	tests := []struct {
		name         string
		hashedPass   string
		plainPass    string
		expectMatch  bool
	}{
		{
			name:        "Correct password",
			hashedPass:  hash,
			plainPass:   password,
			expectMatch: true,
		},
		{
			name:        "Incorrect password",
			hashedPass:  hash,
			plainPass:   "WrongPassword",
			expectMatch: false,
		},
		{
			name:        "Empty password",
			hashedPass:  hash,
			plainPass:   "",
			expectMatch: false,
		},
		{
			name:        "Case sensitive",
			hashedPass:  hash,
			plainPass:   "testpassword123",
			expectMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ComparePassword(tt.hashedPass, tt.plainPass)

			if tt.expectMatch {
				if err != nil {
					t.Errorf("ComparePassword() expected match, got error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("ComparePassword() expected mismatch, got match")
				}
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "Valid password with letter and number",
			password:  "Password123",
			wantError: false,
		},
		{
			name:      "Valid password minimum length",
			password:  "Pass123a",
			wantError: false,
		},
		{
			name:      "Password too short",
			password:  "Pass12",
			wantError: true,
			errorMsg:  "at least 8 characters",
		},
		{
			name:      "Password too long",
			password:  strings.Repeat("a", 129) + "1",
			wantError: true,
			errorMsg:  "less than 128 characters",
		},
		{
			name:      "Password without number",
			password:  "PasswordOnly",
			wantError: true,
			errorMsg:  "at least one number",
		},
		{
			name:      "Password without letter",
			password:  "12345678",
			wantError: true,
			errorMsg:  "at least one letter",
		},
		{
			name:      "Password with special characters",
			password:  "P@ssw0rd!",
			wantError: false,
		},
		{
			name:      "Password with spaces",
			password:  "Pass word 123",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)

			if tt.wantError {
				if err == nil {
					t.Errorf("ValidatePassword() expected error, got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("ValidatePassword() error message = %v, want to contain %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePassword() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestHashPassword_Consistency(t *testing.T) {
	password := "TestPassword123"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	// Bcrypt should produce different hashes due to random salt
	if hash1 == hash2 {
		t.Errorf("HashPassword() produced identical hashes, expected different due to salt")
	}

	// But both should verify against the original password
	if err := ComparePassword(hash1, password); err != nil {
		t.Errorf("ComparePassword() failed for hash1: %v", err)
	}

	if err := ComparePassword(hash2, password); err != nil {
		t.Errorf("ComparePassword() failed for hash2: %v", err)
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "BenchmarkPassword123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = HashPassword(password)
	}
}

func BenchmarkComparePassword(b *testing.B) {
	password := "BenchmarkPassword123"
	hash, _ := HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ComparePassword(hash, password)
	}
}

func BenchmarkValidatePassword(b *testing.B) {
	password := "BenchmarkPassword123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidatePassword(password)
	}
}
