package handler

import (
	"context"
	"database/sql"

	"github.com/ikampus/auth-service/internal/service"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/ikampus/protos/auth"
	commonpb "github.com/ikampus/protos/common"
)

// AuthHandler implements the gRPC AuthService interface
type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	service *service.AuthService
	logger  *logrus.Logger
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(svc *service.AuthService, logger *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		service: svc,
		logger:  logger,
	}
}

// SignUp handles user registration
func (h *AuthHandler) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	h.logger.WithField("email", req.Email).Info("SignUp request received")

	// Validate request
	if req.Email == "" || req.Password == "" || req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "email, password, and username are required")
	}

	// Convert to service request
	svcReq := &service.SignUpRequest{
		Email:          req.Email,
		Password:       req.Password,
		Username:       req.Username,
		DisplayName:    req.DisplayName,
		UniversityID:   req.UniversityId,
		Course:         req.Course,
		YearOfStudy:    int(req.YearOfStudy),
		GraduationYear: int(req.GraduationYear),
		Interests:      req.Interests,
	}

	// Call service
	resp, err := h.service.SignUp(ctx, svcReq)
	if err != nil {
		h.logger.WithError(err).Error("SignUp failed")
		return nil, h.handleError(err)
	}

	// Convert response
	return &pb.SignUpResponse{
		UserId:        resp.UserID,
		Email:         req.Email,
		Username:      req.Username,
		EmailVerified: false,
		Message:       resp.Message,
	}, nil
}

// VerifyEmail handles email verification
func (h *AuthHandler) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	h.logger.Info("VerifyEmail request received")

	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	// Convert to service request
	svcReq := &service.VerifyEmailRequest{
		Token: req.Token,
	}

	// Call service
	resp, err := h.service.VerifyEmail(ctx, svcReq)
	if err != nil {
		h.logger.WithError(err).Error("VerifyEmail failed")
		return nil, h.handleError(err)
	}

	// Convert response
	return &pb.VerifyEmailResponse{
		Success: true,
		Message: resp.Message,
		Tokens: &pb.TokenPair{
			AccessToken:           resp.AccessToken,
			RefreshToken:          resp.RefreshToken,
			AccessTokenExpiresAt:  timestamppb.New(resp.ExpiresAt),
			RefreshTokenExpiresAt: timestamppb.New(resp.ExpiresAt.Add(h.service.GetRefreshTokenExpiry())),
		},
	}, nil
}

// ResendVerificationEmail resends the verification email
func (h *AuthHandler) ResendVerificationEmail(ctx context.Context, req *pb.ResendVerificationRequest) (*commonpb.SuccessResponse, error) {
	h.logger.WithField("email", req.Email).Info("ResendVerificationEmail request received")

	// TODO: Implement resend verification email logic
	return nil, status.Error(codes.Unimplemented, "resend verification email not yet implemented")
}

// Login handles user authentication
func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	h.logger.WithField("email", req.Email).Info("Login request received")

	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	// Get IP address from context metadata if available
	ipAddress := h.getIPFromContext(ctx)

	// Convert to service request
	svcReq := &service.LoginRequest{
		Email:      req.Email,
		Password:   req.Password,
		DeviceID:   req.DeviceId,
		DeviceName: req.DeviceName,
		DeviceType: req.DeviceType,
		UserAgent:  req.UserAgent,
		IPAddress:  ipAddress,
	}

	// Call service
	resp, err := h.service.Login(ctx, svcReq)
	if err != nil {
		h.logger.WithError(err).Error("Login failed")
		return nil, h.handleError(err)
	}

	// Get user profile for response
	profile, err := h.getUserProfile(ctx, resp.UserID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to get user profile")
	}

	// Convert response
	pbUser := &pb.User{
		Id:            resp.UserID,
		Email:         resp.Email,
		Username:      resp.Username,
		Role:          h.convertRole(resp.Role),
		EmailVerified: true,
	}

	if profile != nil {
		pbUser.DisplayName = stringValue(profile.DisplayName)
		pbUser.AvatarUrl = stringValue(profile.AvatarURL)
		pbUser.UniversityId = stringValue(profile.UniversityID)
	}

	return &pb.LoginResponse{
		Success: true,
		Message: "Login successful",
		Tokens: &pb.TokenPair{
			AccessToken:           resp.AccessToken,
			RefreshToken:          resp.RefreshToken,
			AccessTokenExpiresAt:  timestamppb.New(resp.ExpiresAt),
			RefreshTokenExpiresAt: timestamppb.New(resp.ExpiresAt.Add(h.service.GetRefreshTokenExpiry())),
		},
		User: pbUser,
	}, nil
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	h.logger.Info("RefreshToken request received")

	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	// Convert to service request
	svcReq := &service.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	}

	// Call service
	resp, err := h.service.RefreshToken(ctx, svcReq)
	if err != nil {
		h.logger.WithError(err).Error("RefreshToken failed")
		return nil, h.handleError(err)
	}

	// Convert response
	return &pb.RefreshTokenResponse{
		Tokens: &pb.TokenPair{
			AccessToken:           resp.AccessToken,
			RefreshToken:          resp.RefreshToken,
			AccessTokenExpiresAt:  timestamppb.New(resp.ExpiresAt),
			RefreshTokenExpiresAt: timestamppb.New(resp.ExpiresAt.Add(h.service.GetRefreshTokenExpiry())),
		},
	}, nil
}

// Logout handles user logout
func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*commonpb.SuccessResponse, error) {
	h.logger.Info("Logout request received")

	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	// Convert to service request
	svcReq := &service.LogoutRequest{
		RefreshToken: req.RefreshToken,
	}

	// Call service
	if err := h.service.Logout(ctx, svcReq); err != nil {
		h.logger.WithError(err).Error("Logout failed")
		return nil, h.handleError(err)
	}

	return &commonpb.SuccessResponse{
		Success: true,
		Message: "Logged out successfully",
	}, nil
}

// LogoutAll handles logout from all devices
func (h *AuthHandler) LogoutAll(ctx context.Context, req *pb.LogoutAllRequest) (*commonpb.SuccessResponse, error) {
	h.logger.WithField("user_id", req.UserId).Info("LogoutAll request received")

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}

	// Convert to service request
	svcReq := &service.LogoutAllRequest{
		UserID: req.UserId,
	}

	// Call service
	if err := h.service.LogoutAll(ctx, svcReq); err != nil {
		h.logger.WithError(err).Error("LogoutAll failed")
		return nil, h.handleError(err)
	}

	return &commonpb.SuccessResponse{
		Success: true,
		Message: "Logged out from all devices successfully",
	}, nil
}

// ValidateToken validates an access token (used by API Gateway)
func (h *AuthHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	if req.AccessToken == "" {
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	// Convert to service request
	svcReq := &service.ValidateTokenRequest{
		AccessToken: req.AccessToken,
	}

	// Call service
	resp, err := h.service.ValidateToken(ctx, svcReq)
	if err != nil {
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	// Convert response
	return &pb.ValidateTokenResponse{
		Valid:    resp.Valid,
		UserId:   resp.UserID,
		Username: resp.Username,
		Role:     h.convertRole(resp.Role),
	}, nil
}

// RequestPasswordReset initiates password reset flow
func (h *AuthHandler) RequestPasswordReset(ctx context.Context, req *pb.PasswordResetRequest) (*commonpb.SuccessResponse, error) {
	h.logger.WithField("email", req.Email).Info("RequestPasswordReset request received")

	// TODO: Implement password reset request logic
	return nil, status.Error(codes.Unimplemented, "password reset not yet implemented")
}

// ResetPassword resets password with token
func (h *AuthHandler) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*commonpb.SuccessResponse, error) {
	h.logger.Info("ResetPassword request received")

	// TODO: Implement password reset logic
	return nil, status.Error(codes.Unimplemented, "password reset not yet implemented")
}

// ChangePassword changes user password
func (h *AuthHandler) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*commonpb.SuccessResponse, error) {
	h.logger.WithField("user_id", req.UserId).Info("ChangePassword request received")

	// TODO: Implement change password logic
	return nil, status.Error(codes.Unimplemented, "change password not yet implemented")
}

// SubmitStudentVerification submits student verification
func (h *AuthHandler) SubmitStudentVerification(ctx context.Context, req *pb.StudentVerificationRequest) (*pb.StudentVerificationResponse, error) {
	h.logger.WithField("user_id", req.UserId).Info("SubmitStudentVerification request received")

	// TODO: Implement student verification submission logic
	return nil, status.Error(codes.Unimplemented, "student verification not yet implemented")
}

// GetVerificationStatus gets student verification status
func (h *AuthHandler) GetVerificationStatus(ctx context.Context, req *pb.GetVerificationStatusRequest) (*pb.VerificationStatusResponse, error) {
	h.logger.WithField("user_id", req.UserId).Info("GetVerificationStatus request received")

	// TODO: Implement verification status retrieval logic
	return nil, status.Error(codes.Unimplemented, "verification status not yet implemented")
}

// GetActiveSessions retrieves all active sessions for a user
func (h *AuthHandler) GetActiveSessions(ctx context.Context, req *pb.GetActiveSessionsRequest) (*pb.GetActiveSessionsResponse, error) {
	h.logger.WithField("user_id", req.UserId).Info("GetActiveSessions request received")

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}

	// Convert to service request
	svcReq := &service.GetActiveSessionsRequest{
		UserID: req.UserId,
	}

	// Call service
	sessions, err := h.service.GetActiveSessions(ctx, svcReq)
	if err != nil {
		h.logger.WithError(err).Error("GetActiveSessions failed")
		return nil, h.handleError(err)
	}

	// Convert response
	pbSessions := make([]*pb.Session, 0, len(sessions))
	for _, session := range sessions {
		pbSessions = append(pbSessions, &pb.Session{
			Id:          session.ID,
			DeviceId:    stringValue(session.DeviceID),
			DeviceName:  stringValue(session.DeviceName),
			DeviceType:  stringValue(session.DeviceType),
			IpAddress:   stringValue(session.IPAddress),
			CreatedAt:   timestamppb.New(session.CreatedAt),
			LastUsedAt:  timestamppb.New(session.LastUsedAt),
			ExpiresAt:   timestamppb.New(session.ExpiresAt),
			IsCurrent:   false, // TODO: Determine current session
		})
	}

	return &pb.GetActiveSessionsResponse{
		Sessions: pbSessions,
	}, nil
}

// RevokeSession revokes a specific session
func (h *AuthHandler) RevokeSession(ctx context.Context, req *pb.RevokeSessionRequest) (*commonpb.SuccessResponse, error) {
	h.logger.WithFields(logrus.Fields{
		"user_id":    req.UserId,
		"session_id": req.SessionId,
	}).Info("RevokeSession request received")

	if req.UserId == "" || req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID and session ID are required")
	}

	// Convert to service request
	svcReq := &service.RevokeSessionRequest{
		UserID:    req.UserId,
		SessionID: req.SessionId,
	}

	// Call service
	if err := h.service.RevokeSession(ctx, svcReq); err != nil {
		h.logger.WithError(err).Error("RevokeSession failed")
		return nil, h.handleError(err)
	}

	return &commonpb.SuccessResponse{
		Success: true,
		Message: "Session revoked successfully",
	}, nil
}

// Helper functions

// handleError converts service errors to gRPC status errors
func (h *AuthHandler) handleError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Map common errors to gRPC status codes
	switch {
	case err == sql.ErrNoRows:
		return status.Error(codes.NotFound, "resource not found")
	case errMsg == "invalid email or password":
		return status.Error(codes.Unauthenticated, errMsg)
	case errMsg == "email not verified. Please check your email for verification link":
		return status.Error(codes.FailedPrecondition, errMsg)
	case errMsg == "email already registered":
		return status.Error(codes.AlreadyExists, errMsg)
	case errMsg == "username already taken":
		return status.Error(codes.AlreadyExists, errMsg)
	case errMsg == "invalid or expired verification token":
		return status.Error(codes.InvalidArgument, errMsg)
	case errMsg == "refresh token has been revoked":
		return status.Error(codes.Unauthenticated, errMsg)
	case errMsg == "refresh token has expired":
		return status.Error(codes.Unauthenticated, errMsg)
	case errMsg == "email must be a valid university email (.ac.uk domain)":
		return status.Error(codes.InvalidArgument, errMsg)
	default:
		if len(errMsg) > 0 && (errMsg[0:7] == "invalid" || errMsg[0:4] == "must") {
			return status.Error(codes.InvalidArgument, errMsg)
		}
		if len(errMsg) > 14 && errMsg[0:14] == "account locked" {
			return status.Error(codes.PermissionDenied, errMsg)
		}
		return status.Error(codes.Internal, "internal server error")
	}
}

// convertRole converts string role to protobuf UserRole
func (h *AuthHandler) convertRole(role string) commonpb.UserRole {
	switch role {
	case "student":
		return commonpb.UserRole_STUDENT
	case "alumni":
		return commonpb.UserRole_ALUMNI
	case "moderator":
		return commonpb.UserRole_MODERATOR
	case "admin":
		return commonpb.UserRole_ADMIN
	default:
		return commonpb.UserRole_STUDENT
	}
}

// getIPFromContext extracts IP address from gRPC metadata
func (h *AuthHandler) getIPFromContext(ctx context.Context) string {
	// TODO: Extract IP from gRPC metadata
	return ""
}

// getUserProfile gets user profile (placeholder)
type UserProfile struct {
	DisplayName  *string
	AvatarURL    *string
	UniversityID *string
}

func (h *AuthHandler) getUserProfile(ctx context.Context, userID string) (*UserProfile, error) {
	// TODO: Implement profile retrieval
	// For now, return nil to indicate profile not fetched
	return nil, nil
}

// stringValue safely returns string value from pointer
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
