# Recommended Microservices for iKampus

**Based on Google Cloud microservices best practices**

---

## Service Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         Mobile/Web Clients                       │
│                     (iOS, Android, React Web)                   │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTPS/WSS
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                        API Gateway Service                       │
│              (Authentication, Rate Limiting, Routing)           │
└────────────────────────────┬────────────────────────────────────┘
                             │ gRPC
            ┌────────────────┼────────────────┐
            ▼                ▼                ▼
┌──────────────────┬──────────────────┬──────────────────────┐
│  Auth & User     │  Social Layer    │  AI & Intelligence   │
├──────────────────┼──────────────────┼──────────────────────┤
│ AuthService      │ FeedService      │ AIAssistantService   │
│ UserService      │ PostService      │ RecommendationSvc    │
│ UniversityService│ CommunityService │ ModerationService    │
│                  │ MessagingService │                      │
│                  │ EventService     │                      │
└──────────────────┴──────────────────┴──────────────────────┘
            │                │                │
            └────────────────┼────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Infrastructure Services                       │
│  NotificationService │ SearchService │ AnalyticsService │        │
│  MediaService        │ StartupService│ MatchmakingService│       │
└─────────────────────────────────────────────────────────────────┘
```

---

## Core Services (Priority 1)

### 1. GatewayService
**Language:** Go
**Port:** 8080 (HTTP/REST), 8443 (HTTPS)
**Responsibility:** API Gateway, authentication, rate limiting

**Key Features:**
- JWT token validation
- Rate limiting per user/IP
- Request routing to backend services
- CORS handling
- API versioning
- WebSocket upgrade for real-time features

**Dependencies:**
- AuthService (for token validation)
- All backend services

**Resources:**
```yaml
requests:
  cpu: 200m
  memory: 128Mi
limits:
  cpu: 500m
  memory: 256Mi
```

---

### 2. AuthService
**Language:** Go / Node.js
**Port:** 9000 (gRPC)
**Responsibility:** Authentication, authorization, student verification

**API Methods:**
```protobuf
service AuthService {
    rpc SignUp(SignUpRequest) returns (SignUpResponse);
    rpc VerifyEmail(VerifyEmailRequest) returns (Empty);
    rpc Login(LoginRequest) returns (LoginResponse);
    rpc RefreshToken(RefreshTokenRequest) returns (TokenResponse);
    rpc VerifyStudentID(VerifyStudentIDRequest) returns (VerificationResponse);
    rpc ValidateToken(ValidateTokenRequest) returns (TokenValidationResponse);
}
```

**Database Tables:**
- `users` (id, email, password_hash, email_verified, created_at)
- `verification_tokens` (token, user_id, expires_at)
- `student_verifications` (user_id, student_id, verification_status, document_url)
- `sessions` (session_id, user_id, refresh_token, expires_at)

**Cache (Redis):**
- `session:{session_id}` → user_data (TTL: 24h)
- `verification_token:{token}` → user_id (TTL: 1h)
- `rate_limit:signup:{ip}` → count (TTL: 1h)

**Resources:**
```yaml
requests:
  cpu: 100m
  memory: 64Mi
limits:
  cpu: 200m
  memory: 128Mi
```

---

### 3. UserService
**Language:** Go
**Port:** 9001 (gRPC)
**Responsibility:** User profiles, preferences, settings

**API Methods:**
```protobuf
service UserService {
    rpc GetUser(GetUserRequest) returns (User);
    rpc UpdateProfile(UpdateProfileRequest) returns (User);
    rpc GetUsersByIds(GetUsersByIdsRequest) returns (GetUsersResponse);
    rpc SearchUsers(SearchUsersRequest) returns (SearchUsersResponse);
    rpc FollowUser(FollowUserRequest) returns (Empty);
    rpc UnfollowUser(UnfollowUserRequest) returns (Empty);
    rpc GetFollowers(GetFollowersRequest) returns (GetFollowersResponse);
    rpc GetFollowing(GetFollowingRequest) returns (GetFollowingResponse);
}
```

**Database Tables:**
- `profiles` (user_id, username, course, year, university_id, bio, avatar_url, interests)
- `user_settings` (user_id, notification_prefs, privacy_settings)
- `follows` (follower_id, following_id, created_at)
- `universities` (id, name, domain, country, verified)

**Cache (Redis):**
- `user:{user_id}` → profile (TTL: 1h)
- `followers:{user_id}` → count (no TTL)
- `following:{user_id}` → count (no TTL)

**Resources:**
```yaml
requests:
  cpu: 100m
  memory: 64Mi
limits:
  cpu: 200m
  memory: 128Mi
```

---

### 4. PostService
**Language:** Go
**Port:** 9002 (gRPC)
**Responsibility:** Posts, comments, reactions

**API Methods:**
```protobuf
service PostService {
    rpc CreatePost(CreatePostRequest) returns (Post);
    rpc GetPost(GetPostRequest) returns (Post);
    rpc UpdatePost(UpdatePostRequest) returns (Post);
    rpc DeletePost(DeletePostRequest) returns (Empty);
    rpc ListPosts(ListPostsRequest) returns (ListPostsResponse);

    rpc AddComment(AddCommentRequest) returns (Comment);
    rpc GetComments(GetCommentsRequest) returns (GetCommentsResponse);
    rpc DeleteComment(DeleteCommentRequest) returns (Empty);

    rpc AddReaction(AddReactionRequest) returns (Empty);
    rpc RemoveReaction(RemoveReactionRequest) returns (Empty);
    rpc GetReactions(GetReactionsRequest) returns (GetReactionsResponse);
}
```

**Database Tables:**
- `posts` (id, user_id, content, image_urls, community_id, is_anonymous, created_at, updated_at)
- `comments` (id, post_id, user_id, content, parent_comment_id, created_at)
- `reactions` (post_id, user_id, reaction_type, created_at)
- `hashtags` (post_id, tag, created_at)

**Cache (Redis):**
- `post:{post_id}` → post_data (TTL: 5min)
- `post_likes:{post_id}` → count (no TTL)
- `post_comments:{post_id}` → count (no TTL)

**Storage (S3/GCS):**
- Post images/videos

**Resources:**
```yaml
requests:
  cpu: 150m
  memory: 128Mi
limits:
  cpu: 300m
  memory: 256Mi
```

---

### 5. FeedService
**Language:** Go
**Port:** 9003 (gRPC)
**Responsibility:** Feed aggregation, ranking, personalization

**API Methods:**
```protobuf
service FeedService {
    rpc GetHomeFeed(GetHomeFeedRequest) returns (GetFeedResponse);
    rpc GetCommunityFeed(GetCommunityFeedRequest) returns (GetFeedResponse);
    rpc GetTrendingFeed(GetTrendingFeedRequest) returns (GetFeedResponse);
    rpc GetUserFeed(GetUserFeedRequest) returns (GetFeedResponse);
    rpc RefreshFeed(RefreshFeedRequest) returns (Empty);
}
```

**Feed Algorithm:**
1. Fetch posts from followed users (UserService → PostService)
2. Fetch posts from joined communities (CommunityService → PostService)
3. Fetch trending posts (Redis cache)
4. Rank by score: `recency × engagement × relevance`
5. Cache feed for 5 minutes

**Cache (Redis):**
- `feed:{user_id}` → post_ids[] (TTL: 5min)
- `trending:global` → post_ids[] (TTL: 10min)
- `trending:{university_id}` → post_ids[] (TTL: 10min)

**Resources:**
```yaml
requests:
  cpu: 150m
  memory: 128Mi
limits:
  cpu: 300m
  memory: 256Mi
```

---

### 6. MessagingService
**Language:** Go (with WebSocket support)
**Port:** 9004 (gRPC), 9005 (WebSocket)
**Responsibility:** Direct messaging, group chats

**API Methods:**
```protobuf
service MessagingService {
    rpc SendMessage(SendMessageRequest) returns (Message);
    rpc GetMessages(GetMessagesRequest) returns (GetMessagesResponse);
    rpc GetConversations(GetConversationsRequest) returns (GetConversationsResponse);
    rpc MarkAsRead(MarkAsReadRequest) returns (Empty);
    rpc DeleteMessage(DeleteMessageRequest) returns (Empty);
    rpc CreateGroupChat(CreateGroupChatRequest) returns (GroupChat);
    rpc AddMemberToGroup(AddMemberRequest) returns (Empty);
}
```

**Database Tables:**
- `conversations` (id, type, participants[], created_at)
- `messages` (id, conversation_id, sender_id, content, created_at, read_by[])
- `message_requests` (from_user_id, to_user_id, status, created_at)

**Cache (Redis):**
- `presence:{user_id}` → online_status (TTL: 5min)
- `unread:{user_id}` → count (no TTL)
- `typing:{conversation_id}` → user_ids[] (TTL: 5s)

**WebSocket Events:**
- `message.received`
- `message.read`
- `user.typing`
- `user.online`
- `user.offline`

**Resources:**
```yaml
requests:
  cpu: 200m
  memory: 256Mi
limits:
  cpu: 400m
  memory: 512Mi
```

---

## Advanced Services (Priority 2)

### 7. AIAssistantService
**Language:** Python
**Port:** 9006 (gRPC)
**Responsibility:** AI-powered academic assistance

**API Methods:**
```protobuf
service AIAssistantService {
    rpc Chat(ChatRequest) returns (ChatResponse);
    rpc SummarizeDocument(SummarizeRequest) returns (SummarizeResponse);
    rpc ExplainConcept(ExplainConceptRequest) returns (ExplanationResponse);
    rpc GenerateStudyPlan(StudyPlanRequest) returns (StudyPlanResponse);
    rpc AnswerQuestion(QuestionRequest) returns (AnswerResponse);
}
```

**Integration:**
- OpenAI API / Anthropic Claude API
- Vector database for RAG (Pinecone/Weaviate)
- Context: user's course, year, university

**Resources:**
```yaml
requests:
  cpu: 200m
  memory: 512Mi
limits:
  cpu: 500m
  memory: 1Gi
```

---

### 8. CommunityService
**Language:** Go
**Port:** 9007 (gRPC)
**Responsibility:** Groups, societies, communities

**API Methods:**
```protobuf
service CommunityService {
    rpc CreateCommunity(CreateCommunityRequest) returns (Community);
    rpc GetCommunity(GetCommunityRequest) returns (Community);
    rpc ListCommunities(ListCommunitiesRequest) returns (ListCommunitiesResponse);
    rpc JoinCommunity(JoinCommunityRequest) returns (Empty);
    rpc LeaveCommunity(LeaveCommunityRequest) returns (Empty);
    rpc GetMembers(GetMembersRequest) returns (GetMembersResponse);
}
```

**Database Tables:**
- `communities` (id, name, description, type, university_id, member_count)
- `community_members` (community_id, user_id, role, joined_at)
- `community_rules` (community_id, rule_text, created_at)

---

### 9. StartupService
**Language:** Go
**Port:** 9008 (gRPC)
**Responsibility:** Startup profiles, teams, pitches

**API Methods:**
```protobuf
service StartupService {
    rpc CreateStartup(CreateStartupRequest) returns (Startup);
    rpc GetStartup(GetStartupRequest) returns (Startup);
    rpc UpdateStartup(UpdateStartupRequest) returns (Startup);
    rpc ListStartups(ListStartupsRequest) returns (ListStartupsResponse);
    rpc AddTeamMember(AddTeamMemberRequest) returns (Empty);
    rpc RemoveTeamMember(RemoveTeamMemberRequest) returns (Empty);
    rpc SubmitPitch(SubmitPitchRequest) returns (Pitch);
}
```

**Database Tables:**
- `startups` (id, name, description, stage, industry, demo_url, founded_at)
- `startup_members` (startup_id, user_id, role, equity_percentage)
- `pitches` (id, startup_id, pitch_deck_url, video_url, created_at)
- `investor_connections` (startup_id, investor_id, status, notes)

---

### 10. NotificationService
**Language:** Node.js
**Port:** 9009 (gRPC)
**Responsibility:** Push notifications, emails, SMS

**API Methods:**
```protobuf
service NotificationService {
    rpc SendNotification(SendNotificationRequest) returns (Empty);
    rpc GetNotifications(GetNotificationsRequest) returns (GetNotificationsResponse);
    rpc MarkAsRead(MarkAsReadRequest) returns (Empty);
    rpc UpdatePreferences(UpdatePreferencesRequest) returns (Empty);
}
```

**Notification Types:**
- New follower
- Post reaction
- Comment on post
- Mention in post
- Message received
- Event reminder
- Startup match

**Channels:**
- Push notification (FCM/APNS)
- Email (SendGrid/AWS SES)
- In-app notification
- SMS (Twilio) - optional

**Resources:**
```yaml
requests:
  cpu: 100m
  memory: 128Mi
limits:
  cpu: 200m
  memory: 256Mi
```

---

## Infrastructure Services (Priority 3)

### 11. SearchService
**Language:** Go
**Port:** 9010 (gRPC)
**Responsibility:** Full-text search across platform

**Backend:** Elasticsearch / Algolia

**Search Indices:**
- Users (username, bio, course)
- Posts (content, hashtags)
- Communities (name, description)
- Startups (name, description, industry)

---

### 12. MediaService
**Language:** Go
**Port:** 9011 (gRPC)
**Responsibility:** Image/video upload, processing, CDN

**Features:**
- Image compression and resizing
- Video transcoding
- Thumbnail generation
- CDN integration (CloudFront/CloudFlare)
- Signed URLs for secure uploads

---

### 13. AnalyticsService
**Language:** Go
**Port:** 9012 (gRPC)
**Responsibility:** Event tracking, insights

**Events:**
- User signup
- Post created
- Post viewed
- Message sent
- Search performed
- Startup created

---

### 14. ModerationService
**Language:** Python
**Port:** 9013 (gRPC)
**Responsibility:** Content moderation (AI-powered)

**Features:**
- Toxic content detection
- Spam filtering
- Profanity detection
- Image moderation
- User reports handling

---

## Deployment Order

### Week 1-2: Foundation
1. GatewayService
2. AuthService
3. UserService

### Week 3-4: Core Social
4. PostService
5. FeedService
6. NotificationService

### Week 5-6: Communication
7. MessagingService
8. CommunityService
9. EventService

### Week 7-8: Advanced Features
10. AIAssistantService
11. StartupService
12. SearchService

### Week 9-10: Infrastructure
13. MediaService
14. AnalyticsService
15. ModerationService

---

## Technology Stack Summary

| Component | Technology | Reasoning |
|-----------|------------|-----------|
| **Services** | Go, Python, Node.js | Go for performance, Python for AI, Node.js for real-time |
| **Communication** | gRPC + Protocol Buffers | Type-safe, efficient, polyglot |
| **API Gateway** | Go with Gin/Echo | High performance HTTP server |
| **Database** | PostgreSQL | ACID compliance, mature, scalable |
| **Cache** | Redis | In-memory speed, pub/sub support |
| **Search** | Elasticsearch | Full-text search, scalable |
| **Storage** | S3/GCS | Object storage for media |
| **Message Queue** | Google Pub/Sub / AWS SQS | Async event processing |
| **Container Orchestration** | Kubernetes (GKE/EKS) | Industry standard, auto-scaling |
| **Service Mesh** | Istio | mTLS, traffic management, observability |
| **Monitoring** | Prometheus + Grafana | Metrics and dashboards |
| **Tracing** | OpenTelemetry + Jaeger | Distributed tracing |
| **Logging** | Fluentd + Elasticsearch | Centralized logging |
| **CI/CD** | GitHub Actions + ArgoCD | Automated deployments |

---

## Estimated Resource Requirements

### Development Environment
- **Nodes:** 3 × n1-standard-4 (4 vCPU, 15GB RAM)
- **Total:** 12 vCPU, 45GB RAM
- **Monthly Cost:** ~$300/month

### Staging Environment
- **Nodes:** 3 × n1-standard-8 (8 vCPU, 30GB RAM)
- **Total:** 24 vCPU, 90GB RAM
- **Monthly Cost:** ~$600/month

### Production Environment (Launch)
- **Nodes:** 5 × n1-standard-16 (16 vCPU, 60GB RAM)
- **Total:** 80 vCPU, 300GB RAM
- **Database:** Cloud SQL (db-n1-standard-4)
- **Redis:** Memorystore (5GB)
- **Monthly Cost:** ~$2,500/month

### Production Environment (Growth - 10K users)
- **Nodes:** 10 × n1-standard-16
- **Total:** 160 vCPU, 600GB RAM
- **Database:** Cloud SQL (db-n1-standard-8)
- **Redis:** Memorystore (10GB)
- **Monthly Cost:** ~$5,000/month

---

**Document Version:** 1.0
**Last Updated:** 2025-11-18
