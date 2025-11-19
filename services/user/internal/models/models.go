package models

import "time"

// Profile represents a user profile in the database
type Profile struct {
	UserID         string
	Username       string
	DisplayName    *string
	Bio            *string
	AvatarURL      *string
	CoverPhotoURL  *string
	Course         *string
	YearOfStudy    *int
	GraduationYear *int
	UniversityID   *string
	Interests      []string
	WebsiteURL     *string
	LinkedInURL    *string
	GithubURL      *string
	IsPublic       bool
	IsSearchable   bool
	ShowEmail      bool
	FollowerCount  int
	FollowingCount int
	PostCount      int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// User represents basic user information from the users table
type User struct {
	ID            string
	Email         string
	EmailVerified bool
	Role          string
	Status        string
	UniversityID  *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Follow represents a follow relationship
type Follow struct {
	ID          string
	FollowerID  string
	FollowingID string
	CreatedAt   time.Time
}

// Block represents a block relationship
type Block struct {
	ID        string
	BlockerID string
	BlockedID string
	Reason    *string
	CreatedAt time.Time
}

// UserStats represents aggregated user statistics
type UserStats struct {
	UserID          string
	FollowerCount   int
	FollowingCount  int
	PostCount       int
	CommentCount    int
	ReactionCount   int
	CommunityCount  int
	EventCount      int
	StartupCount    int
	EngagementScore float64
	LastActiveAt    *time.Time
}

// UserSettings represents user preferences and settings
type UserSettings struct {
	UserID                 string
	ProfileIsPublic        bool
	ProfileIsSearchable    bool
	ShowEmail              bool
	ShowOnlineStatus       bool
	NotificationsEnabled   bool
	EmailNotifications     bool
	PushNotifications      bool
	ShowNSFWContent        bool
	Language               string
	Timezone               string
	AllowMessageRequests   bool
	AllowTags              bool
	AllowMentions          bool
	UpdatedAt              time.Time
}

// UserRelationship represents the relationship between two users
type UserRelationship struct {
	IsFollowing bool
	IsFollower  bool
	IsBlocked   bool
	IsBlocking  bool
}

// ActivitySummary represents user activity summary
type ActivitySummary struct {
	PostsCreated      int
	CommentsMade      int
	ReactionsGiven    int
	EventsAttended    int
	CommunitiesJoined int
}

// DailyActivity represents daily activity metrics
type DailyActivity struct {
	Date          string // YYYY-MM-DD
	PostCount     int
	CommentCount  int
	ReactionCount int
}

// SearchFilters represents filters for user search
type SearchFilters struct {
	UniversityID  string
	Course        string
	YearOfStudy   int
	Interests     []string
	VerifiedOnly  bool
}
