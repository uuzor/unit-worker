package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ikampus/auth-service/internal/config"
	"github.com/ikampus/auth-service/internal/models"
	"github.com/ikampus/auth-service/internal/repository"
	"github.com/ikampus/auth-service/internal/utils"
	"github.com/sirupsen/logrus"
)

// AuthService handles authentication business logic
type AuthService struct {
	repo       *repository.UserRepository
	jwtManager *utils.JWTManager
	config     *config.Config
	logger     *logrus.Logger
}

// NewAuthService creates a new AuthService instance
func NewAuthService(db *sql.DB, cfg *config.Config, logger *logrus.Logger) (*AuthService, error) {
	repo := repository.NewUserRepository(db)
	jwtManager := utils.NewJWTManager(
		cfg.JWT.AccessTokenSecret,
		cfg.JWT.RefreshTokenSecret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
		cfg.JWT.Issuer,
	)

	return &AuthService{
		repo:       repo,
		jwtManager: jwtManager,
		config:     cfg,
		logger:     logger,
	}, nil
}

// SignUpRequest represents sign-up parameters
type SignUpRequest struct {
	Email          string
	Password       string
	Username       string
	DisplayName    string
	UniversityID   string
	Course         string
	YearOfStudy    int
	GraduationYear int
	Interests      []string
}

// SignUpResponse represents sign-up result
type SignUpResponse struct {
	UserID               string
	Message              string
	VerificationRequired bool
}

// SignUp creates a new user account
func (s *AuthService) SignUp(ctx context.Context, req *SignUpRequest) (*SignUpResponse, error) {
	s.logger.WithField("email", req.Email).Info("Starting sign-up process")

	// 1. Validate university email
	if !s.isUniversityEmail(req.Email) {
		return nil, fmt.Errorf("email must be a valid university email (.ac.uk domain)")
	}

	// 2. Validate password strength
	if err := utils.ValidatePassword(req.Password); err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	// 3. Validate username (alphanumeric and underscores only, 3-30 chars)
	if !s.isValidUsername(req.Username) {
		return nil, fmt.Errorf("username must be 3-30 characters and contain only letters, numbers, and underscores")
	}

	// 4. Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		s.logger.WithError(err).Error("Failed to hash password")
		return nil, fmt.Errorf("failed to process password")
	}

	// 5. Generate email verification token
	verificationToken, err := s.generateSecureToken()
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate verification token")
		return nil, fmt.Errorf("failed to generate verification token")
	}

	verificationExpiry := time.Now().Add(s.config.Email.VerificationExpiry)

	// 6. Create user and profile
	userID := fmt.Sprintf("user_%s", uuid.New().String()[:8])

	user := &models.User{
		ID:                         userID,
		Email:                      strings.ToLower(req.Email),
		EmailVerified:              false,
		EmailVerificationToken:     &verificationToken,
		EmailVerificationExpiresAt: &verificationExpiry,
		PasswordHash:               passwordHash,
		Role:                       "student",
		Status:                     "pending_verification",
		UniversityID:               &req.UniversityID,
		FailedLoginAttempts:        0,
		CreatedAt:                  time.Now(),
		UpdatedAt:                  time.Now(),
	}

	profile := &models.Profile{
		UserID:         userID,
		Username:       req.Username,
		DisplayName:    &req.DisplayName,
		Course:         &req.Course,
		YearOfStudy:    &req.YearOfStudy,
		GraduationYear: &req.GraduationYear,
		Interests:      req.Interests,
		IsPublic:       true,
		IsSearchable:   true,
		ShowEmail:      false,
		FollowerCount:  0,
		FollowingCount: 0,
		PostCount:      0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 7. Save to database
	if err := s.repo.CreateUser(ctx, user, profile); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			if strings.Contains(err.Error(), "email") {
				return nil, fmt.Errorf("email already registered")
			}
			if strings.Contains(err.Error(), "username") {
				return nil, fmt.Errorf("username already taken")
			}
		}
		s.logger.WithError(err).Error("Failed to create user")
		return nil, fmt.Errorf("failed to create user account")
	}

	// 8. Send verification email (TODO: implement email service)
	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"email":   req.Email,
		"token":   verificationToken,
	}).Info("User created, verification email should be sent")

	return &SignUpResponse{
		UserID:               userID,
		Message:              "Account created successfully. Please check your email to verify your account.",
		VerificationRequired: true,
	}, nil
}

// VerifyEmailRequest represents email verification parameters
type VerifyEmailRequest struct {
	Token string
}

// VerifyEmailResponse represents email verification result
type VerifyEmailResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	Message      string
}

// VerifyEmail verifies user email with token
func (s *AuthService) VerifyEmail(ctx context.Context, req *VerifyEmailRequest) (*VerifyEmailResponse, error) {
	s.logger.WithField("token", req.Token[:10]+"...").Info("Verifying email")

	// 1. Verify email in database
	if err := s.repo.VerifyEmail(ctx, req.Token); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid or expired verification token")
		}
		s.logger.WithError(err).Error("Failed to verify email")
		return nil, fmt.Errorf("failed to verify email")
	}

	// 2. Get user by verification token to generate tokens
	user, err := s.repo.GetUserByVerificationToken(ctx, req.Token)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user after verification")
		return nil, fmt.Errorf("failed to complete verification")
	}

	// 3. Generate JWT tokens
	accessToken, expiresAt, err := s.jwtManager.GenerateAccessToken(user.ID, "", user.Email, user.Role)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate access token")
		return nil, fmt.Errorf("failed to generate access token")
	}

	refreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate refresh token")
		return nil, fmt.Errorf("failed to generate refresh token")
	}

	// 4. Store refresh token in database
	refreshTokenModel := &models.RefreshToken{
		ID:         fmt.Sprintf("rt_%s", uuid.New().String()[:8]),
		UserID:     user.ID,
		TokenHash:  repository.HashToken(refreshToken),
		ExpiresAt:  refreshExpiresAt,
		Revoked:    false,
		LastUsedAt: time.Now(),
		CreatedAt:  time.Now(),
	}

	if err := s.repo.CreateRefreshToken(ctx, refreshTokenModel); err != nil {
		s.logger.WithError(err).Error("Failed to store refresh token")
		return nil, fmt.Errorf("failed to store refresh token")
	}

	s.logger.WithField("user_id", user.ID).Info("Email verified successfully")

	return &VerifyEmailResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		Message:      "Email verified successfully",
	}, nil
}

// LoginRequest represents login parameters
type LoginRequest struct {
	Email      string
	Password   string
	DeviceID   string
	DeviceName string
	DeviceType string
	UserAgent  string
	IPAddress  string
}

// LoginResponse represents login result
type LoginResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	UserID       string
	Username     string
	Email        string
	Role         string
}

// Login authenticates a user
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	s.logger.WithField("email", req.Email).Info("Login attempt")

	// 1. Get user by email
	user, err := s.repo.GetUserByEmail(ctx, strings.ToLower(req.Email))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid email or password")
		}
		s.logger.WithError(err).Error("Failed to get user")
		return nil, fmt.Errorf("login failed")
	}

	// 2. Check if account is locked
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		lockDuration := time.Until(*user.LockedUntil).Round(time.Minute)
		return nil, fmt.Errorf("account locked due to too many failed login attempts. Try again in %v", lockDuration)
	}

	// 3. Check if email is verified
	if !user.EmailVerified {
		return nil, fmt.Errorf("email not verified. Please check your email for verification link")
	}

	// 4. Verify password
	if err := utils.ComparePassword(user.PasswordHash, req.Password); err != nil {
		// Increment failed login attempts
		if err := s.repo.IncrementFailedLoginAttempts(ctx, user.ID); err != nil {
			s.logger.WithError(err).Error("Failed to increment failed login attempts")
		}

		s.logger.WithField("user_id", user.ID).Warn("Failed login attempt - invalid password")
		return nil, fmt.Errorf("invalid email or password")
	}

	// 5. Get username from profile
	username := ""
	profile, err := s.repo.GetProfileByUserID(ctx, user.ID)
	if err == nil && profile != nil {
		username = profile.Username
	}

	// 6. Generate JWT tokens
	accessToken, expiresAt, err := s.jwtManager.GenerateAccessToken(user.ID, username, user.Email, user.Role)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate access token")
		return nil, fmt.Errorf("failed to generate access token")
	}

	refreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate refresh token")
		return nil, fmt.Errorf("failed to generate refresh token")
	}

	// 7. Store refresh token with device information
	refreshTokenModel := &models.RefreshToken{
		ID:         fmt.Sprintf("rt_%s", uuid.New().String()[:8]),
		UserID:     user.ID,
		TokenHash:  repository.HashToken(refreshToken),
		DeviceID:   stringPtr(req.DeviceID),
		DeviceName: stringPtr(req.DeviceName),
		DeviceType: stringPtr(req.DeviceType),
		UserAgent:  stringPtr(req.UserAgent),
		IPAddress:  stringPtr(req.IPAddress),
		ExpiresAt:  refreshExpiresAt,
		Revoked:    false,
		LastUsedAt: time.Now(),
		CreatedAt:  time.Now(),
	}

	if err := s.repo.CreateRefreshToken(ctx, refreshTokenModel); err != nil {
		s.logger.WithError(err).Error("Failed to store refresh token")
		return nil, fmt.Errorf("failed to store refresh token")
	}

	// 8. Update last login timestamp and IP
	if err := s.repo.UpdateLastLogin(ctx, user.ID, req.IPAddress); err != nil {
		s.logger.WithError(err).Error("Failed to update last login")
		// Non-critical error, continue
	}

	s.logger.WithField("user_id", user.ID).Info("Login successful")

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		UserID:       user.ID,
		Username:     username,
		Email:        user.Email,
		Role:         user.Role,
	}, nil
}

// RefreshTokenRequest represents token refresh parameters
type RefreshTokenRequest struct {
	RefreshToken string
}

// RefreshTokenResponse represents token refresh result
type RefreshTokenResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// RefreshToken generates new access and refresh tokens
func (s *AuthService) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	s.logger.Info("Refreshing token")

	// 1. Validate refresh token
	claims, err := s.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// 2. Check if token exists and is not revoked
	tokenHash := repository.HashToken(req.RefreshToken)
	storedToken, err := s.repo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid refresh token")
		}
		s.logger.WithError(err).Error("Failed to get refresh token")
		return nil, fmt.Errorf("failed to validate refresh token")
	}

	if storedToken.Revoked {
		return nil, fmt.Errorf("refresh token has been revoked")
	}

	if storedToken.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("refresh token has expired")
	}

	// 3. Get user information
	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user")
		return nil, fmt.Errorf("user not found")
	}

	// Get username
	username := ""
	profile, err := s.repo.GetProfileByUserID(ctx, user.ID)
	if err == nil && profile != nil {
		username = profile.Username
	}

	// 4. Generate new tokens
	newAccessToken, expiresAt, err := s.jwtManager.GenerateAccessToken(user.ID, username, user.Email, user.Role)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate access token")
		return nil, fmt.Errorf("failed to generate access token")
	}

	newRefreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate refresh token")
		return nil, fmt.Errorf("failed to generate refresh token")
	}

	// 5. Revoke old refresh token
	if err := s.repo.RevokeRefreshToken(ctx, storedToken.ID); err != nil {
		s.logger.WithError(err).Error("Failed to revoke old refresh token")
		// Non-critical, continue
	}

	// 6. Store new refresh token (with same device info)
	newRefreshTokenModel := &models.RefreshToken{
		ID:         fmt.Sprintf("rt_%s", uuid.New().String()[:8]),
		UserID:     user.ID,
		TokenHash:  repository.HashToken(newRefreshToken),
		DeviceID:   storedToken.DeviceID,
		DeviceName: storedToken.DeviceName,
		DeviceType: storedToken.DeviceType,
		UserAgent:  storedToken.UserAgent,
		IPAddress:  storedToken.IPAddress,
		ExpiresAt:  refreshExpiresAt,
		Revoked:    false,
		LastUsedAt: time.Now(),
		CreatedAt:  time.Now(),
	}

	if err := s.repo.CreateRefreshToken(ctx, newRefreshTokenModel); err != nil {
		s.logger.WithError(err).Error("Failed to store new refresh token")
		return nil, fmt.Errorf("failed to store new refresh token")
	}

	s.logger.WithField("user_id", user.ID).Info("Token refreshed successfully")

	return &RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// ValidateTokenRequest represents token validation parameters
type ValidateTokenRequest struct {
	AccessToken string
}

// ValidateTokenResponse represents token validation result
type ValidateTokenResponse struct {
	Valid    bool
	UserID   string
	Username string
	Email    string
	Role     string
}

// ValidateToken validates an access token (used by API Gateway)
func (s *AuthService) ValidateToken(ctx context.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	claims, err := s.jwtManager.ValidateAccessToken(req.AccessToken)
	if err != nil {
		return &ValidateTokenResponse{Valid: false}, nil
	}

	return &ValidateTokenResponse{
		Valid:    true,
		UserID:   claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		Role:     claims.Role,
	}, nil
}

// LogoutRequest represents logout parameters
type LogoutRequest struct {
	RefreshToken string
}

// Logout revokes a refresh token
func (s *AuthService) Logout(ctx context.Context, req *LogoutRequest) error {
	s.logger.Info("Logging out user")

	tokenHash := repository.HashToken(req.RefreshToken)
	storedToken, err := s.repo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // Already logged out
		}
		s.logger.WithError(err).Error("Failed to get refresh token")
		return fmt.Errorf("logout failed")
	}

	if err := s.repo.RevokeRefreshToken(ctx, storedToken.ID); err != nil {
		s.logger.WithError(err).Error("Failed to revoke refresh token")
		return fmt.Errorf("logout failed")
	}

	s.logger.WithField("user_id", storedToken.UserID).Info("User logged out successfully")
	return nil
}

// LogoutAllRequest represents logout from all devices parameters
type LogoutAllRequest struct {
	UserID string
}

// LogoutAll revokes all refresh tokens for a user
func (s *AuthService) LogoutAll(ctx context.Context, req *LogoutAllRequest) error {
	s.logger.WithField("user_id", req.UserID).Info("Logging out from all devices")

	if err := s.repo.RevokeAllRefreshTokens(ctx, req.UserID); err != nil {
		s.logger.WithError(err).Error("Failed to revoke all refresh tokens")
		return fmt.Errorf("failed to logout from all devices")
	}

	s.logger.WithField("user_id", req.UserID).Info("Logged out from all devices successfully")
	return nil
}

// GetActiveSessionsRequest represents active sessions request parameters
type GetActiveSessionsRequest struct {
	UserID string
}

// GetActiveSessions retrieves all active sessions for a user
func (s *AuthService) GetActiveSessions(ctx context.Context, req *GetActiveSessionsRequest) ([]*models.RefreshToken, error) {
	sessions, err := s.repo.GetActiveRefreshTokens(ctx, req.UserID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get active sessions")
		return nil, fmt.Errorf("failed to retrieve active sessions")
	}

	return sessions, nil
}

// RevokeSessionRequest represents session revocation parameters
type RevokeSessionRequest struct {
	UserID    string
	SessionID string
}

// RevokeSession revokes a specific session
func (s *AuthService) RevokeSession(ctx context.Context, req *RevokeSessionRequest) error {
	s.logger.WithFields(logrus.Fields{
		"user_id":    req.UserID,
		"session_id": req.SessionID,
	}).Info("Revoking session")

	// Verify the session belongs to the user
	token, err := s.repo.GetRefreshTokenByID(ctx, req.SessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("session not found")
		}
		s.logger.WithError(err).Error("Failed to get session")
		return fmt.Errorf("failed to revoke session")
	}

	if token.UserID != req.UserID {
		return fmt.Errorf("unauthorized: session does not belong to user")
	}

	if err := s.repo.RevokeRefreshToken(ctx, req.SessionID); err != nil {
		s.logger.WithError(err).Error("Failed to revoke session")
		return fmt.Errorf("failed to revoke session")
	}

	s.logger.Info("Session revoked successfully")
	return nil
}

// Helper functions

// isUniversityEmail checks if email has .ac.uk domain
func (s *AuthService) isUniversityEmail(email string) bool {
	return strings.HasSuffix(strings.ToLower(email), ".ac.uk")
}

// isValidUsername validates username format
func (s *AuthService) isValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 30 {
		return false
	}

	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	return true
}

// generateSecureToken generates a cryptographically secure random token
func (s *AuthService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// GetRefreshTokenExpiry returns the refresh token expiry duration
func (s *AuthService) GetRefreshTokenExpiry() time.Duration {
	return s.config.JWT.RefreshTokenExpiry
}
