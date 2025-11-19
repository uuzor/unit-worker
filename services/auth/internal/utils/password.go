package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost is the cost factor for bcrypt hashing (10-12 recommended for production)
	BcryptCost = 12
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	// Minimum password length
	if len(password) < 8 {
		return "", fmt.Errorf("password must be at least 8 characters")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// ComparePassword compares a hashed password with a plaintext password
func ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	if len(password) > 128 {
		return fmt.Errorf("password must be less than 128 characters")
	}

	// Check for at least one number
	hasNumber := false
	hasLetter := false
	for _, char := range password {
		if char >= '0' && char <= '9' {
			hasNumber = true
		}
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			hasLetter = true
		}
	}

	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}

	if !hasLetter {
		return fmt.Errorf("password must contain at least one letter")
	}

	return nil
}
