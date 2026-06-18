# Voxlab Backend Architecture

## Overview

Voxlab Backend is a public speaking education platform built with Go, using a hexagonal/clean architecture pattern. It provides RESTful APIs for educational content management, user progression tracking, and media storage.

## System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                          CLIENTS                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────────┐ │
│  │   Teacher    │  │   Student    │  │   Admin Frontend       │ │
└──┴──────┬───────┴──┴──────┬───────┴──┴────────┬───────────────┴─┘
          │                  │                    │
          └──────────────────┴────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                         API LAYER                                │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │                    Gin Router                             │ │
│  │  /api/v1/                                                 │ │
│  │    /health      ───┐                                     │ │
│  │    /auth/*      ───┼──► AuthController                   │ │
│  │    /tracks/*    ───┤    TrackController                  │ │
│  │    /modules/*   ───┤    ModuleController                 │ │
│  │    /lessons/*   ───┤    LessonController                 │ │
│  │    /exercises/* ───┤    ExerciseController               │ │
│  │    /progress/*  ───┤    ProgressController               │ │
│  │    /users/*     ───┤    UserController (admin)           │ │
│  │    /upload/*    ───┤    UploadController                 │ │
│  └─────────────────┬──┴────────────────────────────────────┘ │
└─────────────────────┼───────────────────────────────────────┘
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                     MIDDLEWARE                                   │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │  CORSMiddleware    → CORS headers by environment             │ │
│  │  AuthMiddleware    → JWT validation + token blacklist check  │ │
│  │  AdminMiddleware   → Role check (user.role == "admin")       │ │
│  └───────────────────────────────────────────────────────────┘ │
└─────────────────────┼───────────────────────────────────────┘
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                    SERVICES LAYER                                │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │  AuthService     → JWT, bcrypt, logout blacklist           │ │
│  │  TrackService    → Business logic for courses               │ │
│  │  ModuleService   → Business logic for modules                │ │
│  │  LessonService   → Business logic for lessons                │ │
│  │  ExerciseService → Exercise CRUD + JSONB handling           │ │
│  │  ProgressService → XP calculation, lesson completion         │ │
│  │  UserService     → Admin user management                    │ │
│  │  UploadService   → Image processing (WebP) + MinIO upload    │ │
│  └───────────────────────────────────────────────────────────┘ │
└─────────────────────┼───────────────────────────────────────┘
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                  REPOSITORIES LAYER                              │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │  UserRepository    → User DB operations                     │ │
│  │  TrackRepository   → Track DB operations                    │ │
│  │  ModuleRepository  → Module DB operations                   │ │
│  │  LessonRepository  → Lesson DB operations                    │ │
│  │  ExerciseRepository→ Exercise DB operations (JSONB)          │ │
│  │  ProgressRepository→ Progress DB operations                 │ │
│  │  CommunityRepository→ Reaction DB operations                 │ │
│  └───────────────────────────────────────────────────────────┘ │
└─────────────────────┼───────────────────────────────────────┘
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│              INFRASTRUCTURE (External Services)                  │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │                    PostgreSQL                              │ │
│  │  Tables: users, tracks, modules, lessons, exercises,       │ │
│  │          module_lessons (pivot), user_progress,             │ │
│  │          progress_statuses, gamified_titles,                │ │
│  │          user_titles, user_reactions                         │ │
│  │                                                            │ │
│  │  JSONB fields: exercises.content                           │ │
│  │  Features: AutoMigrate, seed data, indexes                  │ │
│  └───────────────────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │                         Redis                              │ │
│  │  Keys:                                                     │ │
│  │    live_sessions:{room_id}     → 2-hour TTL                │ │
│  │    auth:blacklist:{sha256}      → 25-hour TTL              │ │
│  │    user:streak:active:{user_id}→ 24-hour TTL               │ │
│  │    user:progress:{user_id}     → cache XP/streak           │ │
│  │    logs:errors                 → last 1000 errors (LRU)   │ │
│  └───────────────────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │                         MinIO                              │ │
│  │  Bucket: voxlab-media                                      │ │
│  │  Path structure:                                             │ │
│  │    tracks/{uuid}.webp      → course cover images            │ │
│  │    modules/{uuid}.webp     → module images                  │ │
│  │    lessons/{uuid}.webp     → lesson images                  │ │
│  │    avatars/{uuid}.webp     → user profile pictures          │ │
│  │  Processing: Resize to fit, convert to WebP (quality 80)    │ │
│  └───────────────────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │                      LiveKit                            │ │
│  │  Signaling: ws://localhost:7880                           │ │
│  │  Uses: WebRTC calls for oratory practice                    │ │
│  └───────────────────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │                    Python Analyzer                         │ │
│  │  Endpoint: http://analyzer:8000/analyze/text               │ │
│  │  Function: NLP analysis of writing exercises                 │ │
│  │  Timeout: 180 seconds                                     │ │
│  └───────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Data Flow

### User Registration/Login Flow
```
1. POST /auth/register or POST /auth/login
2. AuthService.ValidateCredentials()
3. Query UserRepository.FindByEmail()
4. If valid: create JWT with claims (user_id, email, role)
5. Return { token, expires_at, user: { id, name, email, role } }
```

### Content Creation Flow (Admin)
```
1. POST /tracks → TrackController.CreateTrack()
2. POST /modules → ModuleController.CreateModule()
3. POST /lessons → LessonController.CreateLesson()
4. POST /modules/:id/lessons → ModuleController.LinkLesson()
5. POST /exercises → ExerciseController.CreateExercise(type=quiz|writing|...)
6. Schema validation in ExerciseController based on type
```

### Lesson Completion Flow
```
1. POST /progress → ProgressController.CompleteLesson()
2. ProgressService.CompleteLesson()
3. Calculate XP: base (10) + exercise_count * 5 + score / 10
4. Upsert UserProgress record
5. Increment User.XP via UserRepository.AddXP()
6. Cache update in Redis: user:progress:{user_id}
```

### File Upload Flow
```
1. POST /upload/avatar → UploadController.UploadAvatar()
2. Read multipart file → io.LimitReader(2MB)
3. UploadService.processAndUpload():
   a. Decode image (JPEG/PNG → image.Image)
   b. Resize to fit (400x400 for avatar, 1920x1080 for covers)
   c. Encode to WebP (quality 80)
   d. Upload via go-storage to MinIO
   e. Return public URL
4. Update User.AvatarURL or Track.IconURL
```

## Database Schema Overview

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    password_hash VARCHAR(255) NOT NULL,
    xp INT DEFAULT 0,
    streak_days INT DEFAULT 0,
    lives INT DEFAULT 5,
    avatar_url VARCHAR(512),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
```

### Exercises JSONB Schema
The `exercises.content` field is polymorphic JSONB. See exercise types below.

## Error Handling Pattern

All endpoints return standardized JSON responses:

```json
{
  "success": true,
  "data": { ... }
}
```

```json
{
  "error": "Error message"
}
```

Common HTTP status codes:
- `200` OK
- `201` Created
- `400` Bad Request (validation error)
- `401` Unauthorized (missing/invalid token)
- `403` Forbidden (admin-only endpoint)
- `404` Not Found
- `409` Conflict (duplicate email)
- `500` Internal Server Error
- `502` Bad Gateway (analyzer unavailable)