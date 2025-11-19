package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ikampus/user-service/internal/models"
	"github.com/lib/pq"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// ==============================================================================
// Get Operations
// ==============================================================================

// GetProfileByUserID retrieves a user profile by user ID
func (r *UserRepository) GetProfileByUserID(ctx context.Context, userID string) (*models.Profile, error) {
	query := `
		SELECT user_id, username, display_name, bio, avatar_url, cover_photo_url,
		       course, year_of_study, graduation_year, university_id,
		       interests, website_url, linkedin_url, github_url,
		       is_public, is_searchable, show_email,
		       follower_count, following_count, post_count,
		       created_at, updated_at
		FROM profiles
		WHERE user_id = $1
	`

	profile := &models.Profile{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&profile.UserID,
		&profile.Username,
		&profile.DisplayName,
		&profile.Bio,
		&profile.AvatarURL,
		&profile.CoverPhotoURL,
		&profile.Course,
		&profile.YearOfStudy,
		&profile.GraduationYear,
		&profile.UniversityID,
		pq.Array(&profile.Interests),
		&profile.WebsiteURL,
		&profile.LinkedInURL,
		&profile.GithubURL,
		&profile.IsPublic,
		&profile.IsSearchable,
		&profile.ShowEmail,
		&profile.FollowerCount,
		&profile.FollowingCount,
		&profile.PostCount,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return profile, nil
}

// GetProfileByUsername retrieves a user profile by username
func (r *UserRepository) GetProfileByUsername(ctx context.Context, username string) (*models.Profile, error) {
	query := `
		SELECT user_id, username, display_name, bio, avatar_url, cover_photo_url,
		       course, year_of_study, graduation_year, university_id,
		       interests, website_url, linkedin_url, github_url,
		       is_public, is_searchable, show_email,
		       follower_count, following_count, post_count,
		       created_at, updated_at
		FROM profiles
		WHERE username = $1
	`

	profile := &models.Profile{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&profile.UserID,
		&profile.Username,
		&profile.DisplayName,
		&profile.Bio,
		&profile.AvatarURL,
		&profile.CoverPhotoURL,
		&profile.Course,
		&profile.YearOfStudy,
		&profile.GraduationYear,
		&profile.UniversityID,
		pq.Array(&profile.Interests),
		&profile.WebsiteURL,
		&profile.LinkedInURL,
		&profile.GithubURL,
		&profile.IsPublic,
		&profile.IsSearchable,
		&profile.ShowEmail,
		&profile.FollowerCount,
		&profile.FollowingCount,
		&profile.PostCount,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return profile, nil
}

// GetUserByID retrieves basic user information
func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	query := `
		SELECT id, email, email_verified, role, status, university_id,
		       created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.EmailVerified,
		&user.Role,
		&user.Status,
		&user.UniversityID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetProfilesByUserIDs retrieves multiple profiles by user IDs (batch operation)
func (r *UserRepository) GetProfilesByUserIDs(ctx context.Context, userIDs []string) ([]*models.Profile, error) {
	if len(userIDs) == 0 {
		return []*models.Profile{}, nil
	}

	query := `
		SELECT user_id, username, display_name, bio, avatar_url, cover_photo_url,
		       course, year_of_study, graduation_year, university_id,
		       interests, website_url, linkedin_url, github_url,
		       is_public, is_searchable, show_email,
		       follower_count, following_count, post_count,
		       created_at, updated_at
		FROM profiles
		WHERE user_id = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(userIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profiles := make([]*models.Profile, 0, len(userIDs))
	for rows.Next() {
		profile := &models.Profile{}
		err := rows.Scan(
			&profile.UserID,
			&profile.Username,
			&profile.DisplayName,
			&profile.Bio,
			&profile.AvatarURL,
			&profile.CoverPhotoURL,
			&profile.Course,
			&profile.YearOfStudy,
			&profile.GraduationYear,
			&profile.UniversityID,
			pq.Array(&profile.Interests),
			&profile.WebsiteURL,
			&profile.LinkedInURL,
			&profile.GithubURL,
			&profile.IsPublic,
			&profile.IsSearchable,
			&profile.ShowEmail,
			&profile.FollowerCount,
			&profile.FollowingCount,
			&profile.PostCount,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}

	return profiles, rows.Err()
}

// ==============================================================================
// Update Operations
// ==============================================================================

// UpdateProfile updates a user's profile
func (r *UserRepository) UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) (*models.Profile, error) {
	if len(updates) == 0 {
		return r.GetProfileByUserID(ctx, userID)
	}

	// Build dynamic UPDATE query
	setClauses := []string{}
	args := []interface{}{}
	argPos := 1

	for key, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", key, argPos))
		args = append(args, value)
		argPos++
	}

	// Always update updated_at
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argPos))
	args = append(args, time.Now())
	argPos++

	// Add user_id for WHERE clause
	args = append(args, userID)

	query := fmt.Sprintf(`
		UPDATE profiles
		SET %s
		WHERE user_id = $%d
	`, strings.Join(setClauses, ", "), argPos)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return r.GetProfileByUserID(ctx, userID)
}

// UpdateAvatar updates a user's avatar URL
func (r *UserRepository) UpdateAvatar(ctx context.Context, userID, avatarURL string) error {
	query := `
		UPDATE profiles
		SET avatar_url = $1, updated_at = $2
		WHERE user_id = $3
	`

	_, err := r.db.ExecContext(ctx, query, avatarURL, time.Now(), userID)
	return err
}

// UpdateCoverPhoto updates a user's cover photo URL
func (r *UserRepository) UpdateCoverPhoto(ctx context.Context, userID, coverPhotoURL string) error {
	query := `
		UPDATE profiles
		SET cover_photo_url = $1, updated_at = $2
		WHERE user_id = $3
	`

	_, err := r.db.ExecContext(ctx, query, coverPhotoURL, time.Now(), userID)
	return err
}

// ==============================================================================
// Search Operations
// ==============================================================================

// SearchUsers searches for users based on query and filters
func (r *UserRepository) SearchUsers(ctx context.Context, query string, filters *models.SearchFilters, limit, offset int) ([]*models.Profile, int, error) {
	// Build WHERE clauses
	whereClauses := []string{"is_searchable = true"}
	args := []interface{}{}
	argPos := 1

	// Text search
	if query != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(username ILIKE $%d OR display_name ILIKE $%d)", argPos, argPos))
		args = append(args, "%"+query+"%")
		argPos++
	}

	// University filter
	if filters != nil && filters.UniversityID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("university_id = $%d", argPos))
		args = append(args, filters.UniversityID)
		argPos++
	}

	// Course filter
	if filters != nil && filters.Course != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("course ILIKE $%d", argPos))
		args = append(args, "%"+filters.Course+"%")
		argPos++
	}

	// Year of study filter
	if filters != nil && filters.YearOfStudy > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("year_of_study = $%d", argPos))
		args = append(args, filters.YearOfStudy)
		argPos++
	}

	// Interests filter
	if filters != nil && len(filters.Interests) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("interests && $%d", argPos))
		args = append(args, pq.Array(filters.Interests))
		argPos++
	}

	// Verified only filter
	if filters != nil && filters.VerifiedOnly {
		whereClauses = append(whereClauses, "EXISTS (SELECT 1 FROM users WHERE users.id = profiles.user_id AND users.student_verified = true)")
	}

	whereClause := strings.Join(whereClauses, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM profiles WHERE %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get profiles
	args = append(args, limit, offset)
	searchQuery := fmt.Sprintf(`
		SELECT user_id, username, display_name, bio, avatar_url, cover_photo_url,
		       course, year_of_study, graduation_year, university_id,
		       interests, website_url, linkedin_url, github_url,
		       is_public, is_searchable, show_email,
		       follower_count, following_count, post_count,
		       created_at, updated_at
		FROM profiles
		WHERE %s
		ORDER BY follower_count DESC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argPos, argPos+1)

	rows, err := r.db.QueryContext(ctx, searchQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	profiles := make([]*models.Profile, 0)
	for rows.Next() {
		profile := &models.Profile{}
		err := rows.Scan(
			&profile.UserID,
			&profile.Username,
			&profile.DisplayName,
			&profile.Bio,
			&profile.AvatarURL,
			&profile.CoverPhotoURL,
			&profile.Course,
			&profile.YearOfStudy,
			&profile.GraduationYear,
			&profile.UniversityID,
			pq.Array(&profile.Interests),
			&profile.WebsiteURL,
			&profile.LinkedInURL,
			&profile.GithubURL,
			&profile.IsPublic,
			&profile.IsSearchable,
			&profile.ShowEmail,
			&profile.FollowerCount,
			&profile.FollowingCount,
			&profile.PostCount,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		profiles = append(profiles, profile)
	}

	return profiles, total, rows.Err()
}

// ==============================================================================
// Follow/Unfollow Operations
// ==============================================================================

// FollowUser creates a follow relationship
func (r *UserRepository) FollowUser(ctx context.Context, followerID, followingID string) error {
	// Check for existing follow
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2)"
	err := r.db.QueryRowContext(ctx, checkQuery, followerID, followingID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return nil // Already following
	}

	// Create follow relationship
	query := `
		INSERT INTO follows (id, follower_id, following_id, created_at)
		VALUES ($1, $2, $3, $4)
	`

	followID := fmt.Sprintf("follow_%s", uuid.New().String()[:8])
	_, err = r.db.ExecContext(ctx, query, followID, followerID, followingID, time.Now())
	return err
}

// UnfollowUser removes a follow relationship
func (r *UserRepository) UnfollowUser(ctx context.Context, followerID, followingID string) error {
	query := `
		DELETE FROM follows
		WHERE follower_id = $1 AND following_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, followerID, followingID)
	return err
}

// GetFollowers retrieves a user's followers
func (r *UserRepository) GetFollowers(ctx context.Context, userID string, limit, offset int) ([]*models.Profile, int, error) {
	// Count total
	countQuery := "SELECT COUNT(*) FROM follows WHERE following_id = $1"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get followers
	query := `
		SELECT p.user_id, p.username, p.display_name, p.bio, p.avatar_url, p.cover_photo_url,
		       p.course, p.year_of_study, p.graduation_year, p.university_id,
		       p.interests, p.website_url, p.linkedin_url, p.github_url,
		       p.is_public, p.is_searchable, p.show_email,
		       p.follower_count, p.following_count, p.post_count,
		       p.created_at, p.updated_at
		FROM profiles p
		INNER JOIN follows f ON p.user_id = f.follower_id
		WHERE f.following_id = $1
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	profiles := make([]*models.Profile, 0)
	for rows.Next() {
		profile := &models.Profile{}
		err := rows.Scan(
			&profile.UserID,
			&profile.Username,
			&profile.DisplayName,
			&profile.Bio,
			&profile.AvatarURL,
			&profile.CoverPhotoURL,
			&profile.Course,
			&profile.YearOfStudy,
			&profile.GraduationYear,
			&profile.UniversityID,
			pq.Array(&profile.Interests),
			&profile.WebsiteURL,
			&profile.LinkedInURL,
			&profile.GithubURL,
			&profile.IsPublic,
			&profile.IsSearchable,
			&profile.ShowEmail,
			&profile.FollowerCount,
			&profile.FollowingCount,
			&profile.PostCount,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		profiles = append(profiles, profile)
	}

	return profiles, total, rows.Err()
}

// GetFollowing retrieves users that a user is following
func (r *UserRepository) GetFollowing(ctx context.Context, userID string, limit, offset int) ([]*models.Profile, int, error) {
	// Count total
	countQuery := "SELECT COUNT(*) FROM follows WHERE follower_id = $1"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get following
	query := `
		SELECT p.user_id, p.username, p.display_name, p.bio, p.avatar_url, p.cover_photo_url,
		       p.course, p.year_of_study, p.graduation_year, p.university_id,
		       p.interests, p.website_url, p.linkedin_url, p.github_url,
		       p.is_public, p.is_searchable, p.show_email,
		       p.follower_count, p.following_count, p.post_count,
		       p.created_at, p.updated_at
		FROM profiles p
		INNER JOIN follows f ON p.user_id = f.following_id
		WHERE f.follower_id = $1
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	profiles := make([]*models.Profile, 0)
	for rows.Next() {
		profile := &models.Profile{}
		err := rows.Scan(
			&profile.UserID,
			&profile.Username,
			&profile.DisplayName,
			&profile.Bio,
			&profile.AvatarURL,
			&profile.CoverPhotoURL,
			&profile.Course,
			&profile.YearOfStudy,
			&profile.GraduationYear,
			&profile.UniversityID,
			pq.Array(&profile.Interests),
			&profile.WebsiteURL,
			&profile.LinkedInURL,
			&profile.GithubURL,
			&profile.IsPublic,
			&profile.IsSearchable,
			&profile.ShowEmail,
			&profile.FollowerCount,
			&profile.FollowingCount,
			&profile.PostCount,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		profiles = append(profiles, profile)
	}

	return profiles, total, rows.Err()
}

// IsFollowing checks if a user is following another user
func (r *UserRepository) IsFollowing(ctx context.Context, followerID, followingID string) (bool, *time.Time, error) {
	query := `
		SELECT created_at
		FROM follows
		WHERE follower_id = $1 AND following_id = $2
	`

	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, query, followerID, followingID).Scan(&createdAt)
	if err == sql.ErrNoRows {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}

	return true, &createdAt, nil
}

// ==============================================================================
// Block/Unblock Operations
// ==============================================================================

// BlockUser creates a block relationship
func (r *UserRepository) BlockUser(ctx context.Context, blockerID, blockedID, reason string) error {
	// Check for existing block
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM blocks WHERE blocker_id = $1 AND blocked_id = $2)"
	err := r.db.QueryRowContext(ctx, checkQuery, blockerID, blockedID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return nil // Already blocked
	}

	// Remove any follow relationships
	r.UnfollowUser(ctx, blockerID, blockedID)
	r.UnfollowUser(ctx, blockedID, blockerID)

	// Create block
	query := `
		INSERT INTO blocks (id, blocker_id, blocked_id, reason, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	blockID := fmt.Sprintf("block_%s", uuid.New().String()[:8])
	var reasonPtr *string
	if reason != "" {
		reasonPtr = &reason
	}

	_, err = r.db.ExecContext(ctx, query, blockID, blockerID, blockedID, reasonPtr, time.Now())
	return err
}

// UnblockUser removes a block relationship
func (r *UserRepository) UnblockUser(ctx context.Context, blockerID, blockedID string) error {
	query := `
		DELETE FROM blocks
		WHERE blocker_id = $1 AND blocked_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, blockerID, blockedID)
	return err
}

// GetBlockedUsers retrieves users that a user has blocked
func (r *UserRepository) GetBlockedUsers(ctx context.Context, userID string, limit, offset int) ([]*models.Block, int, error) {
	// Count total
	countQuery := "SELECT COUNT(*) FROM blocks WHERE blocker_id = $1"
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get blocks
	query := `
		SELECT id, blocker_id, blocked_id, reason, created_at
		FROM blocks
		WHERE blocker_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	blocks := make([]*models.Block, 0)
	for rows.Next() {
		block := &models.Block{}
		err := rows.Scan(
			&block.ID,
			&block.BlockerID,
			&block.BlockedID,
			&block.Reason,
			&block.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		blocks = append(blocks, block)
	}

	return blocks, total, rows.Err()
}

// IsBlocked checks if a user has blocked another user
func (r *UserRepository) IsBlocked(ctx context.Context, blockerID, blockedID string) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM blocks WHERE blocker_id = $1 AND blocked_id = $2)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, blockerID, blockedID).Scan(&exists)
	return exists, err
}

// GetUserRelationship gets the relationship between two users
func (r *UserRepository) GetUserRelationship(ctx context.Context, userID, otherUserID string) (*models.UserRelationship, error) {
	rel := &models.UserRelationship{}

	// Check if following
	isFollowing, _, err := r.IsFollowing(ctx, userID, otherUserID)
	if err != nil {
		return nil, err
	}
	rel.IsFollowing = isFollowing

	// Check if follower
	isFollower, _, err := r.IsFollowing(ctx, otherUserID, userID)
	if err != nil {
		return nil, err
	}
	rel.IsFollower = isFollower

	// Check if blocked
	isBlocked, err := r.IsBlocked(ctx, userID, otherUserID)
	if err != nil {
		return nil, err
	}
	rel.IsBlocked = isBlocked

	// Check if blocking
	isBlocking, err := r.IsBlocked(ctx, otherUserID, userID)
	if err != nil {
		return nil, err
	}
	rel.IsBlocking = isBlocking

	return rel, nil
}

// ==============================================================================
// Statistics and Activity Operations
// ==============================================================================

// GetUserStats retrieves user statistics
func (r *UserRepository) GetUserStats(ctx context.Context, userID string) (*models.UserStats, error) {
	query := `
		SELECT follower_count, following_count, post_count
		FROM profiles
		WHERE user_id = $1
	`

	stats := &models.UserStats{UserID: userID}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.FollowerCount,
		&stats.FollowingCount,
		&stats.PostCount,
	)
	if err != nil {
		return nil, err
	}

	// TODO: Get other stats from analytics tables when implemented
	stats.CommentCount = 0
	stats.ReactionCount = 0
	stats.CommunityCount = 0
	stats.EventCount = 0
	stats.StartupCount = 0
	stats.EngagementScore = 0.0

	return stats, nil
}

// ==============================================================================
// Settings Operations
// ==============================================================================

// GetUserSettings retrieves user settings (placeholder)
func (r *UserRepository) GetUserSettings(ctx context.Context, userID string) (*models.UserSettings, error) {
	// TODO: Implement when settings table is created
	// For now, return default settings based on profile
	profile, err := r.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &models.UserSettings{
		UserID:               userID,
		ProfileIsPublic:      profile.IsPublic,
		ProfileIsSearchable:  profile.IsSearchable,
		ShowEmail:            profile.ShowEmail,
		ShowOnlineStatus:     true,
		NotificationsEnabled: true,
		EmailNotifications:   true,
		PushNotifications:    true,
		ShowNSFWContent:      false,
		Language:             "en",
		Timezone:             "UTC",
		AllowMessageRequests: true,
		AllowTags:            true,
		AllowMentions:        true,
		UpdatedAt:            time.Now(),
	}, nil
}

// UpdateUserSettings updates user settings (placeholder)
func (r *UserRepository) UpdateUserSettings(ctx context.Context, userID string, settings *models.UserSettings) error {
	// TODO: Implement when settings table is created
	// For now, just update profile privacy settings
	updates := map[string]interface{}{
		"is_public":      settings.ProfileIsPublic,
		"is_searchable":  settings.ProfileIsSearchable,
		"show_email":     settings.ShowEmail,
	}

	_, err := r.UpdateProfile(ctx, userID, updates)
	return err
}

// ==============================================================================
// Account Management Operations
// ==============================================================================

// DeactivateAccount deactivates a user account
func (r *UserRepository) DeactivateAccount(ctx context.Context, userID string) error {
	query := `
		UPDATE users
		SET status = 'inactive', updated_at = $1
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}

// DeleteAccount soft deletes a user account
func (r *UserRepository) DeleteAccount(ctx context.Context, userID string) error {
	query := `
		UPDATE users
		SET status = 'deleted', updated_at = $1
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}
