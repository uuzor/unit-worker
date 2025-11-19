package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ikampus/auth-service/internal/models"
	"github.com/lib/pq"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user and profile
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User, profile *models.Profile) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert user
	userQuery := `
		INSERT INTO users (
			id, email, email_verified, email_verification_token,
			email_verification_expires_at, password_hash, role, status,
			university_id, failed_login_attempts, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = tx.ExecContext(ctx, userQuery,
		user.ID, user.Email, user.EmailVerified, user.EmailVerificationToken,
		user.EmailVerificationExpiresAt, user.PasswordHash, user.Role, user.Status,
		user.UniversityID, user.FailedLoginAttempts, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return fmt.Errorf("email already exists")
			}
		}
		return fmt.Errorf("failed to insert user: %w", err)
	}

	// Insert profile
	profileQuery := `
		INSERT INTO profiles (
			user_id, username, display_name, course, year_of_study,
			graduation_year, interests, is_public, is_searchable,
			show_email, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = tx.ExecContext(ctx, profileQuery,
		profile.UserID, profile.Username, profile.DisplayName, profile.Course,
		profile.YearOfStudy, profile.GraduationYear, pq.Array(profile.Interests),
		profile.IsPublic, profile.IsSearchable, profile.ShowEmail,
		profile.CreatedAt, profile.UpdatedAt,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return fmt.Errorf("username already exists")
			}
		}
		return fmt.Errorf("failed to insert profile: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}

	query := `
		SELECT id, email, email_verified, email_verification_token,
			   email_verification_expires_at, password_hash, role, status,
			   university_id, last_login_at, last_login_ip,
			   failed_login_attempts, locked_until, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.EmailVerified, &user.EmailVerificationToken,
		&user.EmailVerificationExpiresAt, &user.PasswordHash, &user.Role, &user.Status,
		&user.UniversityID, &user.LastLoginAt, &user.LastLoginIP,
		&user.FailedLoginAttempts, &user.LockedUntil, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	user := &models.User{}

	query := `
		SELECT id, email, email_verified, email_verification_token,
			   email_verification_expires_at, password_hash, role, status,
			   university_id, last_login_at, last_login_ip,
			   failed_login_attempts, locked_until, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Email, &user.EmailVerified, &user.EmailVerificationToken,
		&user.EmailVerificationExpiresAt, &user.PasswordHash, &user.Role, &user.Status,
		&user.UniversityID, &user.LastLoginAt, &user.LastLoginIP,
		&user.FailedLoginAttempts, &user.LockedUntil, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetProfileByUserID retrieves a profile by user ID
func (r *UserRepository) GetProfileByUserID(ctx context.Context, userID string) (*models.Profile, error) {
	profile := &models.Profile{}

	query := `
		SELECT user_id, username, display_name, bio, course, year_of_study,
			   graduation_year, avatar_url, cover_photo_url, interests,
			   website_url, linkedin_url, github_url, is_public, is_searchable,
			   show_email, follower_count, following_count, post_count,
			   created_at, updated_at
		FROM profiles
		WHERE user_id = $1
	`

	var interests pq.StringArray
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&profile.UserID, &profile.Username, &profile.DisplayName, &profile.Bio,
		&profile.Course, &profile.YearOfStudy, &profile.GraduationYear,
		&profile.AvatarURL, &profile.CoverPhotoURL, &interests,
		&profile.WebsiteURL, &profile.LinkedInURL, &profile.GitHubURL,
		&profile.IsPublic, &profile.IsSearchable, &profile.ShowEmail,
		&profile.FollowerCount, &profile.FollowingCount, &profile.PostCount,
		&profile.CreatedAt, &profile.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("profile not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	profile.Interests = interests

	return profile, nil
}

// VerifyEmail marks a user's email as verified
func (r *UserRepository) VerifyEmail(ctx context.Context, token string) error {
	query := `
		UPDATE users
		SET email_verified = true,
			email_verification_token = NULL,
			email_verification_expires_at = NULL,
			updated_at = $1
		WHERE email_verification_token = $2
		  AND email_verification_expires_at > $1
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), token)
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("invalid or expired verification token")
	}

	return nil
}

// UpdateLastLogin updates the user's last login information
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID, ipAddress string) error {
	query := `
		UPDATE users
		SET last_login_at = $1,
			last_login_ip = $2,
			failed_login_attempts = 0,
			locked_until = NULL,
			updated_at = $1
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), ipAddress, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// IncrementFailedLoginAttempts increments failed login attempts
func (r *UserRepository) IncrementFailedLoginAttempts(ctx context.Context, userID string) error {
	query := `
		UPDATE users
		SET failed_login_attempts = failed_login_attempts + 1,
			locked_until = CASE
				WHEN failed_login_attempts >= 4 THEN $1
				ELSE NULL
			END,
			updated_at = $2
		WHERE id = $3
	`

	lockUntil := time.Now().Add(30 * time.Minute) // Lock for 30 minutes after 5 failed attempts
	_, err := r.db.ExecContext(ctx, query, lockUntil, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to increment failed login attempts: %w", err)
	}

	return nil
}

// CreateRefreshToken creates a new refresh token
func (r *UserRepository) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (
			id, user_id, token_hash, device_id, device_name,
			device_type, user_agent, ip_address, expires_at,
			last_used_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		token.ID, token.UserID, token.TokenHash, token.DeviceID,
		token.DeviceName, token.DeviceType, token.UserAgent,
		token.IPAddress, token.ExpiresAt, token.LastUsedAt, token.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

// GetRefreshToken retrieves a refresh token by hash
func (r *UserRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	token := &models.RefreshToken{}

	query := `
		SELECT id, user_id, token_hash, device_id, device_name,
			   device_type, user_agent, ip_address, expires_at,
			   revoked, revoked_at, last_used_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.DeviceID,
		&token.DeviceName, &token.DeviceType, &token.UserAgent,
		&token.IPAddress, &token.ExpiresAt, &token.Revoked,
		&token.RevokedAt, &token.LastUsedAt, &token.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("refresh token not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return token, nil
}

// RevokeRefreshToken revokes a refresh token
func (r *UserRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = true,
			revoked_at = $1
		WHERE token_hash = $2
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), tokenHash)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (r *UserRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = true,
			revoked_at = $1
		WHERE user_id = $2 AND revoked = false
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all tokens: %w", err)
	}

	return nil
}

// HashToken creates a SHA-256 hash of a token
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GenerateVerificationToken generates a random verification token
func GenerateVerificationToken() (string, error) {
	token := uuid.New().String()
	return token, nil
}

// UpdateVerificationToken updates the email verification token
func (r *UserRepository) UpdateVerificationToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	query := `
		UPDATE users
		SET email_verification_token = $1,
			email_verification_expires_at = $2,
			updated_at = $3
		WHERE id = $4
	`

	_, err := r.db.ExecContext(ctx, query, token, expiresAt, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update verification token: %w", err)
	}

	return nil
}
