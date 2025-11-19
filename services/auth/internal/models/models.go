package models

import (
	"time"
)

// User represents a user in the database
type User struct {
	ID                        string
	Email                     string
	EmailVerified             bool
	EmailVerificationToken    *string
	EmailVerificationExpiresAt *time.Time
	PasswordHash              string
	Role                      string
	Status                    string
	UniversityID              *string
	LastLoginAt               *time.Time
	LastLoginIP               *string
	FailedLoginAttempts       int
	LockedUntil               *time.Time
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

// Profile represents a user profile
type Profile struct {
	UserID          string
	Username        string
	DisplayName     *string
	Bio             *string
	Course          *string
	YearOfStudy     *int
	GraduationYear  *int
	AvatarURL       *string
	CoverPhotoURL   *string
	Interests       []string
	WebsiteURL      *string
	LinkedInURL     *string
	GitHubURL       *string
	IsPublic        bool
	IsSearchable    bool
	ShowEmail       bool
	FollowerCount   int
	FollowingCount  int
	PostCount       int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	ID         string
	UserID     string
	TokenHash  string
	DeviceID   *string
	DeviceName *string
	DeviceType *string
	UserAgent  *string
	IPAddress  *string
	ExpiresAt  time.Time
	Revoked    bool
	RevokedAt  *time.Time
	LastUsedAt time.Time
	CreatedAt  time.Time
}

// StudentVerification represents a student verification record
type StudentVerification struct {
	ID                  string
	UserID              string
	StudentID           *string
	DocumentURL         *string
	VerificationMethod  string
	Status              string
	SubmittedAt         time.Time
	ReviewedAt          *time.Time
	ReviewedBy          *string
	RejectionReason     *string
	ExpiresAt           *time.Time
	Notes               *string
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	Used      bool
	UsedAt    *time.Time
	CreatedAt time.Time
}

// Session represents an active user session
type Session struct {
	ID         string
	DeviceID   *string
	DeviceName *string
	DeviceType *string
	IPAddress  *string
	CreatedAt  time.Time
	LastUsedAt time.Time
	ExpiresAt  time.Time
	IsCurrent  bool
}
