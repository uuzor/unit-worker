# 🚀 iKampus — Platform Overview for Developers

## 1. Core Purpose

iKampus is a university-exclusive digital platform designed to connect students, support academic life, and help young entrepreneurs build startups.

**Only verified students** (and optionally alumni) can join, making it a safe, authentic campus ecosystem.

---

## 🧩 2. Core Modules / Components

### A. Authentication & Verification

**Sign-up Process:**
- Sign-up using `.ac.uk` university email
- Email verification + optional student ID verification
- Create custom username, profile with:
  - Course
  - Year of study
  - University
  - Profile photo
  - Interests

**Goal:** Only real students → trusted environment.

---

### B. Social Interaction Layer

A lightweight social network for campus communities.

**Features:**
- **Campus Feed** (posts, images, events)
- **Communities & Groups** (by course, societies, halls, interests)
- **Direct Messaging** with request system
- **Anonymous Confessions / Posts** (optional module)
- **Campus "Near Me" mode** (see active users in your university area)
- **Events Board** (societies & clubs can post events)

---

### C. AI 24/7 Assistant

An integrated AI system that helps students with:
- Academic support (summaries, explanations, notes)
- University life guidance
- Mental load reduction (reminders, schedules)
- Personalized support based on course/year
- Campus knowledge (locations, study spots, deadlines)

---

### D. Startup & Entrepreneur Hub

A dedicated space for student entrepreneurs.

**Features:**
- **Register a startup** with:
  - Name
  - Description
  - Team members
  - Stage (idea, MVP, scaling)
  - Demo links
- **AI startup mentor** (pitch feedback, branding help, plans)
- **Matchmaking:**
  - Students looking for co-founders
  - Alumni who can mentor or invest
- Display startup profiles publicly within the university network

---

### E. Alumni Integration (Optional but Powerful)

Verified alumni can:
- Offer mentoring
- Post internships & job openings
- Support startups
- Join discussion groups
- Share career stories

**Benefit:** Helps build a lifelong community.

---

## 🏗️ 3. Technical Architecture (High-Level)

### Frontend
- **Mobile-first** (React Native / Flutter recommended)
- Clean Gen-Z-friendly UI
- Dark mode default

### Backend
- Auth & user management
- Post/feed system
- Real-time chat & notifications (WebSockets / Firebase)
- AI integration API
- Startup hub database
- Event & community interaction module
- Moderation tools (user reports, auto-flagging)

### Database

**Core Data Models:**
- User profiles
- University list & verification states
- Posts & activity feed
- Messages
- Startup profiles
- Community groups & memberships

### Security
- OAuth-style email verification
- Rate limiting
- Content moderation (AI + human review)
- End-to-end or server-side encrypted messaging (depending on design needs)

---

## 🎨 4. Key UX Principles for Gen Z

Programmers must build with:
- **Speed** (fast loading, minimal steps)
- **Simplicity** (modern, clean UI)
- **Social presence** (real-time, interactive features)
- **Visually engaging components**
- **Safe and non-toxic environment**

---

## 💡 5. What Makes iKampus Unique

Unlike normal student apps, iKampus combines:

✅ Social network
✅ AI academic assistant
✅ Entrepreneur/startup hub
✅ Campus-specific ecosystem
✅ Verified students only

This creates an **all-in-one student platform**.

---

## 📦 6. Next Steps for Developers

This platform overview can be expanded into:
- ✅ Full Developer Requirements Document (DRD)
- ✅ Wireframes / UI layouts
- ✅ Feature roadmap
- ✅ API structure and endpoints
- ✅ Database schema

---

## 📚 Additional Documentation

For detailed technical specifications, please refer to the `/docs` directory:
- `/docs/architecture/` - System architecture and design patterns
- `/docs/api/` - API documentation and endpoints
- `/docs/database/` - Database schema and models
- `/docs/features/` - Detailed feature specifications
- `/docs/security/` - Security guidelines and best practices
