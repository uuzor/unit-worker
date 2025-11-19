# iKampus Protocol Buffer Definitions

**gRPC API Contracts for iKampus Microservices**

Version: 1.0
Last Updated: 2025-11-18

---

## Overview

This directory contains Protocol Buffer (`.proto`) definitions for all iKampus microservices. These contracts define type-safe APIs for service-to-service communication using gRPC.

---

## Directory Structure

```
protos/
├── common/              # Shared types and messages
│   └── common.proto
├── auth/                # Authentication service
│   └── auth.proto
├── user/                # User profile service
│   └── user.proto
├── social/              # Social features
│   ├── post.proto
│   ├── feed.proto       # (to be created)
│   ├── community.proto  # (to be created)
│   └── event.proto      # (to be created)
├── messaging/           # Messaging service
│   └── messaging.proto  # (to be created)
├── startup/             # Startup ecosystem
│   ├── startup.proto    # (to be created)
│   ├── matching.proto   # (to be created)
│   └── mentor.proto     # (to be created)
└── infrastructure/      # Platform services
    ├── notification.proto # (to be created)
    ├── search.proto      # (to be created)
    ├── analytics.proto   # (to be created)
    └── media.proto       # (to be created)
```

---

## Services Overview

### ✅ Implemented

| Service | Proto File | Description | RPCs |
|---------|-----------|-------------|------|
| **AuthService** | `auth/auth.proto` | Authentication, authorization, verification | 16 |
| **UserService** | `user/user.proto` | User profiles, follows, blocks | 25 |
| **PostService** | `social/post.proto` | Posts, comments, reactions | 24 |

### 🚧 To Be Implemented

| Service | Proto File | Description |
|---------|-----------|-------------|
| **FeedService** | `social/feed.proto` | Personalized feeds, trending |
| **CommunityService** | `social/community.proto` | Groups, societies, membership |
| **EventService** | `social/event.proto` | Campus events, RSVPs |
| **MessagingService** | `messaging/messaging.proto` | Direct messages, group chats |
| **StartupService** | `startup/startup.proto` | Startup profiles, teams |
| **MatchmakingService** | `startup/matching.proto` | Co-founder matching |
| **MentorService** | `startup/mentor.proto` | Mentorship connections |
| **NotificationService** | `infrastructure/notification.proto` | Multi-channel notifications |
| **SearchService** | `infrastructure/search.proto` | Full-text search |
| **AnalyticsService** | `infrastructure/analytics.proto` | Event tracking |
| **MediaService** | `infrastructure/media.proto` | Media upload/processing |

---

## Common Types

The `common/common.proto` file provides shared types used across all services:

**Core Types:**
- `Empty` - Empty message
- `SuccessResponse` - Generic success message
- `PaginationRequest` / `PaginationResponse` - Pagination
- `UUID` - UUID wrapper
- `TimeRange` - Time filtering
- `Location` - Geographic location (PostGIS)
- `MediaAttachment` - Media files

**Reference Types:**
- `UserRef` - Lightweight user reference
- `UniversityRef` - University reference
- `CommunityRef` - Community reference
- `StartupRef` - Startup reference

**Enums:**
- `UserRole` - student, alumni, moderator, admin
- `AccountStatus` - active, suspended, deleted, banned
- `VerificationStatus` - pending, approved, rejected, expired
- `Visibility` - public, university, community, followers, private
- `SortOrder` - asc, desc

---

## Code Generation

### Prerequisites

Install Protocol Buffer compiler and plugins:

```bash
# Protocol Buffer Compiler
# macOS
brew install protobuf

# Ubuntu/Debian
apt-get install -y protobuf-compiler

# Go plugin
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Python plugin
pip install grpcio-tools

# Node.js plugin
npm install -g grpc-tools

# C# plugin (included with .NET SDK)
```

### Generate Code

**Go:**
```bash
./scripts/generate-go.sh
```

**Python:**
```bash
./scripts/generate-python.sh
```

**Node.js:**
```bash
./scripts/generate-nodejs.sh
```

**C#:**
```bash
./scripts/generate-csharp.sh
```

**All Languages:**
```bash
./scripts/generate-all.sh
```

---

## Usage Examples

### Go

```go
import (
    "context"
    "google.golang.org/grpc"
    authpb "github.com/ikampus/protos/auth"
)

// Create client
conn, err := grpc.Dial("auth-service:9000", grpc.WithInsecure())
client := authpb.NewAuthServiceClient(conn)

// Make RPC call
resp, err := client.Login(context.Background(), &authpb.LoginRequest{
    Email:    "student@uni.ac.uk",
    Password: "password123",
})
```

### Python

```python
import grpc
from ikampus.protos.auth import auth_pb2, auth_pb2_grpc

# Create channel and stub
channel = grpc.insecure_channel('auth-service:9000')
stub = auth_pb2_grpc.AuthServiceStub(channel)

# Make RPC call
response = stub.Login(auth_pb2.LoginRequest(
    email='student@uni.ac.uk',
    password='password123'
))
```

### Node.js

```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

// Load proto
const packageDefinition = protoLoader.loadSync('auth/auth.proto');
const authProto = grpc.loadPackageDefinition(packageDefinition).ikampus.auth;

// Create client
const client = new authProto.AuthService(
    'auth-service:9000',
    grpc.credentials.createInsecure()
);

// Make RPC call
client.Login({
    email: 'student@uni.ac.uk',
    password: 'password123'
}, (err, response) => {
    console.log(response);
});
```

---

## API Documentation

### AuthService

**Authentication and Authorization**

```protobuf
service AuthService {
  rpc SignUp(SignUpRequest) returns (SignUpResponse);
  rpc VerifyEmail(VerifyEmailRequest) returns (VerifyEmailResponse);
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
  rpc Logout(LogoutRequest) returns (SuccessResponse);
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc RequestPasswordReset(PasswordResetRequest) returns (SuccessResponse);
  rpc ResetPassword(ResetPasswordRequest) returns (SuccessResponse);
  rpc ChangePassword(ChangePasswordRequest) returns (SuccessResponse);
  rpc SubmitStudentVerification(StudentVerificationRequest) returns (StudentVerificationResponse);
  rpc GetVerificationStatus(GetVerificationStatusRequest) returns (VerificationStatusResponse);
  rpc GetActiveSessions(GetActiveSessionsRequest) returns (GetActiveSessionsResponse);
  rpc RevokeSession(RevokeSessionRequest) returns (SuccessResponse);
}
```

**Key Endpoints:**
- `SignUp` - Register new student with university email
- `Login` - Authenticate and get JWT tokens
- `ValidateToken` - Validate JWT (used by API Gateway)
- `SubmitStudentVerification` - Upload student ID for verification

---

### UserService

**User Profile Management**

```protobuf
service UserService {
  rpc GetUser(GetUserRequest) returns (UserProfile);
  rpc GetUserByUsername(GetUserByUsernameRequest) returns (UserProfile);
  rpc GetUsers(GetUsersRequest) returns (GetUsersResponse);
  rpc UpdateProfile(UpdateProfileRequest) returns (UserProfile);
  rpc SearchUsers(SearchUsersRequest) returns (SearchUsersResponse);
  rpc FollowUser(FollowUserRequest) returns (SuccessResponse);
  rpc UnfollowUser(UnfollowUserRequest) returns (SuccessResponse);
  rpc GetFollowers(GetFollowersRequest) returns (GetFollowersResponse);
  rpc GetFollowing(GetFollowingRequest) returns (GetFollowingResponse);
  rpc BlockUser(BlockUserRequest) returns (SuccessResponse);
  rpc GetUserStats(GetUserStatsRequest) returns (UserStats);
  rpc UpdateSettings(UpdateSettingsRequest) returns (UserSettings);
}
```

**Key Endpoints:**
- `GetUser` - Get user profile by ID
- `SearchUsers` - Search users by name, course, interests
- `FollowUser` - Follow another user
- `GetUserStats` - Get engagement statistics

---

### PostService

**Posts, Comments, and Reactions**

```protobuf
service PostService {
  rpc CreatePost(CreatePostRequest) returns (Post);
  rpc GetPost(GetPostRequest) returns (Post);
  rpc UpdatePost(UpdatePostRequest) returns (Post);
  rpc DeletePost(DeletePostRequest) returns (SuccessResponse);
  rpc ListPosts(ListPostsRequest) returns (ListPostsResponse);
  rpc AddComment(AddCommentRequest) returns (Comment);
  rpc GetComments(GetCommentsRequest) returns (GetCommentsResponse);
  rpc AddPostReaction(AddPostReactionRequest) returns (SuccessResponse);
  rpc RemovePostReaction(RemovePostReactionRequest) returns (SuccessResponse);
  rpc SavePost(SavePostRequest) returns (SuccessResponse);
  rpc VotePoll(VotePollRequest) returns (PollResults);
  rpc ReportPost(ReportPostRequest) returns (SuccessResponse);
}
```

**Key Endpoints:**
- `CreatePost` - Create text, image, video, poll, or confession post
- `AddComment` - Comment on a post (supports nesting)
- `AddPostReaction` - React with like, love, celebrate, etc.
- `VotePoll` - Vote on a poll

---

## Best Practices

### 1. Versioning

Version your protos in the package name:
```protobuf
package ikampus.auth.v1;
```

### 2. Backward Compatibility

- Never change field numbers
- Only add new fields with optional or repeated
- Use `reserved` for removed fields

```protobuf
message User {
  reserved 4, 5;  // Removed fields
  reserved "old_field_name";

  string id = 1;
  string name = 2;
  string email = 3;
  // Field 4 and 5 are reserved
  string new_field = 6;
}
```

### 3. Field Naming

- Use `snake_case` for field names
- Use descriptive names
- Use `_id` suffix for IDs
- Use `_url` suffix for URLs
- Use `_at` suffix for timestamps

### 4. Optional Fields

Use `optional` for fields that may not be set:
```protobuf
message UpdateProfileRequest {
  string user_id = 1;
  optional string display_name = 2;
  optional string bio = 3;
}
```

### 5. Pagination

Always use pagination for list endpoints:
```protobuf
message ListPostsRequest {
  common.PaginationRequest pagination = 1;
}

message ListPostsResponse {
  repeated Post posts = 1;
  common.PaginationResponse pagination = 2;
}
```

### 6. Error Handling

Use gRPC status codes:
- `OK` (0) - Success
- `INVALID_ARGUMENT` (3) - Bad request
- `UNAUTHENTICATED` (16) - Not authenticated
- `PERMISSION_DENIED` (7) - Not authorized
- `NOT_FOUND` (5) - Resource not found
- `ALREADY_EXISTS` (6) - Duplicate resource
- `INTERNAL` (13) - Server error

### 7. Request Metadata

Include `requester_id` for authorization:
```protobuf
message GetUserRequest {
  string user_id = 1;
  string requester_id = 2;  // For permission checks
}
```

---

## Testing

### grpcurl

Test RPCs using `grpcurl`:

```bash
# List services
grpcurl -plaintext localhost:9000 list

# List methods
grpcurl -plaintext localhost:9000 list ikampus.auth.AuthService

# Call method
grpcurl -plaintext -d '{"email":"test@uni.ac.uk","password":"test123"}' \
  localhost:9000 ikampus.auth.AuthService/Login
```

### Postman

Postman supports gRPC testing:
1. Create new gRPC request
2. Import proto files
3. Select service and method
4. Fill in request message
5. Send

---

## CI/CD Integration

### GitHub Actions

```yaml
name: Generate Protos

on:
  push:
    paths:
      - 'protos/**/*.proto'

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install protoc
        run: |
          apt-get update
          apt-get install -y protobuf-compiler

      - name: Generate Go code
        run: ./scripts/generate-go.sh

      - name: Commit generated code
        run: |
          git add generated/
          git commit -m "chore: regenerate proto code"
          git push
```

---

## Troubleshooting

### Common Issues

**1. Import not found**

Ensure proto path is correct:
```bash
protoc --proto_path=protos --go_out=. protos/auth/auth.proto
```

**2. Plugin not found**

Install the protoc plugin:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

**3. Package conflicts**

Use unique package names:
```protobuf
option go_package = "github.com/ikampus/protos/auth;auth";
```

---

## Resources

**Official Documentation:**
- Protocol Buffers: https://protobuf.dev/
- gRPC: https://grpc.io/
- gRPC Go: https://grpc.io/docs/languages/go/
- gRPC Python: https://grpc.io/docs/languages/python/
- gRPC Node.js: https://grpc.io/docs/languages/node/

**Style Guide:**
- Google Protocol Buffer Style Guide: https://protobuf.dev/programming-guides/style/

**Tools:**
- grpcurl: https://github.com/fullstorydev/grpcurl
- Buf: https://buf.build/ (Proto linting and breaking change detection)

---

**Version:** 1.0
**Last Updated:** 2025-11-18
**Maintained by:** iKampus Engineering Team
