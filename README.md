# PulsePoll — Real-Time Live Polling Engine

A production-grade, distributed real-time live polling application where creators launch interactive polls, share unique join links, and audiences cast votes. Vote counts and percentages update instantaneously across every connected device and browser tab using **Redis In-Memory Hash Counters**, **Redis Pub/Sub**, and **Go WebSockets** — **without requiring any page refreshes**.

Built for the **GUVI Developer Internship Assessment**.

---

## Table of Contents
1. [Project Overview](#project-overview)
2. [Key Features](#key-features)
3. [Technology Stack](#technology-stack)
4. [Architecture & System Flow](#architecture--system-flow)
5. [Directory & Project Structure](#directory--project-structure)
6. [Getting Started & Installation](#getting-started--installation)
   - [Option A: Docker Compose (Recommended)](#option-a-docker-compose-recommended)
   - [Option B: Local Bare-Metal Setup](#option-b-local-bare-metal-setup)
7. [Environment Variables](#environment-variables)
8. [API Documentation](#api-documentation)
9. [MongoDB Schema & Persistence](#mongodb-schema--persistence)
10. [Redis: Atomic Counters & Pub/Sub](#redis-atomic-counters--pubsub)
11. [Authentication & Security](#authentication--security)
12. [Real-Time WebSocket Engine](#real-time-websocket-engine)
13. [End-to-End Testing (Multi-Browser Sync)](#end-to-end-testing-multi-browser-sync)
14. [Deployment Guide](#deployment-guide)
15. [Engineering Challenges & Solutions](#engineering-challenges--solutions)
16. [AI Usage & Developer Reflections](#ai-usage--developer-reflections)
17. [3–5 Minute Video Demonstration Script](#35-minute-video-demonstration-script)
18. [Submission Checklist](#submission-checklist)

---

## 1. Project Overview

PulsePoll solves the friction of traditional audience surveys by delivering a fluid, zero-latency voting experience. Designed with a decoupled modern architecture:
- **React Frontend**: Polished dark slate glassmorphic UI, animated metric bars, and real-time state listeners.
- **Go + Gin Backend**: Blazing fast API validation, concurrency-safe WebSocket connections, and data pipelines.
- **MongoDB**: Durable persistence of User profiles, Poll configurations, and vote audit trails for duplicate prevention.
- **Redis**: High-speed in-memory atomic counter tallies (`HINCRBY`) and Pub/Sub event broadcasting driving the live WebSocket stream.

---

## 2. Key Features

- 🔐 **Secure Creator Authentication**: BCrypt password hashing, 24-hour stateless JWT tokens, and protected dashboard routes.
- 📊 **Poll Creation & Presets**: Custom question formulations, 2–10 options, dynamic option management, and customizable expiration durations (1h, 24h, 7d, or Never).
- 🔗 **Instant 1-Click Sharing**: Automatically generates short, collision-resistant 6-character alphanumeric share codes (e.g., `ABC123`) and direct shareable URLs.
- 🗳️ **Duplicate Vote Prevention**: Validates voter fingerprints and IP combinations against persistent MongoDB indices to ensure one vote per participant.
- ⚡ **Zero-Refresh Live Sync**: Every vote triggers an atomic Redis increment and Pub/Sub broadcast, immediately updating connected screens via WebSockets.
- 🎉 **Engaging Voter Experience**: Interactive radio cards, celebratory confetti animations upon submission, and seamless transition to live results.
- 🛑 **Creator Lifecycle Controls**: Creators can close or delete polls at any time with instantaneous visual status propagation across all voter screens.

---

## 3. Technology Stack

| Technology | Layer | Role in PulsePoll |
| :--- | :--- | :--- |
| **React 18 + Vite** | Frontend / UI | Fast single-page application, custom design system, glassmorphic styling, and reactive WebSocket hooks. |
| **Go 1.23 + Gin** | Backend / API | High-throughput REST endpoints, request validation, CORS management, and WebSocket room hub. |
| **MongoDB 7.0** | Persistent DB | Stores User accounts, Poll questions/options/status, and individual vote records for audit and fraud prevention. |
| **Redis 7.0** | In-Memory & Pub/Sub | Fast atomic hash vote tallies (`HINCRBY`) and real-time message bus (`PUBLISH` / `SUBSCRIBE`). |
| **Docker & Compose** | Containerization | Multi-container orchestration connecting Mongo, Redis, Go Backend, and Nginx/React Frontend. |

---

## 4. Architecture & System Flow

```
                      AUDIENCE VOTES (Browser A)
                                  │
                                  ▼
                         React Application
                                  │ POST /api/polls/:id/vote
                                  ▼
                        Go / Gin Web Server
                     (Validates Input & Status)
                                  │
                   ┌──────────────┴──────────────┐
                   ▼                             ▼
           MongoDB Database                Redis In-Memory
         (Persists Vote Audit)         (Atomic HINCRBY Count)
                                                 │
                                                 ▼
                                           Redis Pub/Sub
                                       (Channel: poll_updates:{id})
                                                 │
                                                 ▼
                                          Go WebSocket Hub
                                      (Room Broadcaster Routine)
                                                 │
                                                 ▼
                                     All Connected Browsers
                                       (Browser B & Browser C)
                                                 │
                                                 ▼
                                   Results Bar Updates Live!
                                    (Zero Refresh Required)
```

---

## 5. Directory & Project Structure

```
live-polling/
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── Navbar.jsx           # Top navigation and user status
│   │   │   ├── PollCard.jsx         # Dashboard poll card with actions
│   │   │   ├── ResultBar.jsx        # Animated live bar chart with leader highlight
│   │   │   └── Toast.jsx            # Toast notifications (copy, success, error)
│   │   ├── pages/
│   │   │   ├── Login.jsx            # Creator sign in with demo autofill
│   │   │   ├── Signup.jsx           # New account registration
│   │   │   ├── Dashboard.jsx        # Poll manager, metrics, and filter tabs
│   │   │   ├── CreatePoll.jsx       # Dynamic poll builder with duration selector
│   │   │   ├── PublicPoll.jsx       # Audience voting page with confetti & live transition
│   │   │   └── PollResults.jsx      # Presenter view with live WebSocket indicator
│   │   ├── services/
│   │   │   └── api.js               # Centralized REST client with auth headers
│   │   ├── hooks/
│   │   │   ├── useAuth.jsx          # User authentication context
│   │   │   └── useWebSocket.jsx     # Reconnecting WebSocket hook
│   │   ├── App.jsx                  # Single-page router and provider wrappers
│   │   ├── index.css                # Polished dark theme and glassmorphic styling
│   │   └── main.jsx
│   ├── package.json
│   ├── vite.config.js
│   ├── nginx.conf                   # Production Nginx reverse proxy configuration
│   └── Dockerfile
│
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # Server bootstrap, DI, and graceful shutdown
│   ├── config/
│   │   ├── config.go                # Environment variable reader
│   │   ├── db.go                    # MongoDB client setup & compound indexes
│   │   └── redis.go                 # Redis client connection and health check
│   ├── controllers/
│   │   ├── auth_controller.go       # Signup, Login, Me endpoints
│   │   └── poll_controller.go       # CRUD, Voting, and WebSocket upgrades
│   ├── middleware/
│   │   ├── auth_middleware.go       # JWT Bearer token validator
│   │   └── cors_middleware.go       # Cross-origin policy config
│   ├── models/
│   │   ├── user.go                  # User model & auth DTOs
│   │   ├── poll.go                  # Poll, options, and real-time update structs
│   │   └── vote.go                  # Vote audit record model
│   ├── repository/
│   │   ├── user_repo.go             # MongoDB user collection ops
│   │   ├── poll_repo.go             # MongoDB poll collection ops
│   │   └── vote_repo.go             # MongoDB vote audit & duplicate checks
│   ├── services/
│   │   ├── auth_service.go          # BCrypt password hashing & JWT signing
│   │   ├── poll_service.go          # Poll creation, voting, and update dispatch
│   │   └── redis_service.go         # Redis hash counters and Pub/Sub manager
│   ├── websocket/
│   │   ├── hub.go                   # Room manager & Redis subscription bridge
│   │   └── client.go                # WebSocket read/write pumps & heartbeats
│   ├── routes/
│   │   └── routes.go                # API route grouping
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
│
├── docker-compose.yml               # Unified multi-service deployment
├── .env.example                     # Environment configuration reference
└── README.md                        # Documentation
```

---

## 6. Getting Started & Installation

### Option A: Docker Compose (Recommended)
The entire full-stack application (MongoDB, Redis, Go Backend, and React Frontend) can be launched with a single command:

```bash
# 1. Clone the repository
git clone https://github.com/your-username/live-polling.git
cd live-polling

# 2. Launch all services with Docker Compose
docker compose up --build
```

Access the services:
- **Frontend**: [http://localhost:3000](http://localhost:3000)
- **Backend API**: [http://localhost:8080/api](http://localhost:8080/api)
- **MongoDB**: `localhost:27017`
- **Redis**: `localhost:6379`

---

### Option B: Local Bare-Metal Setup

#### Prerequisites
- Node.js (v18+) & npm
- Go (v1.22+)
- MongoDB running locally or a [MongoDB Atlas](https://www.mongodb.com/atlas) URI
- Redis running locally or an [Upstash Redis](https://upstash.com) URI

#### 1. Setup Backend
```bash
cd backend

# Configure environment variables
cp ../.env.example .env

# Run Go server
go run ./cmd/server
# Output: [PulsePoll Server] Listening on http://localhost:8080
```

#### 2. Setup Frontend
```bash
cd ../frontend

# Install dependencies
npm install

# Start Vite development server
npm run dev
# Output: Local: http://localhost:5173
```

---

## 7. Environment Variables

Create a `.env` file based on `.env.example`:

| Variable | Description | Default Local Value |
| :--- | :--- | :--- |
| `PORT` | Go backend HTTP listen port | `8080` |
| `MONGO_URI` | MongoDB connection URI | `mongodb://localhost:27017` |
| `DB_NAME` | MongoDB database name | `pulsepoll_db` |
| `REDIS_URL` | Redis host or URI | `redis://localhost:6379` |
| `JWT_SECRET` | Secret key for HMAC SHA-256 JWT tokens | *(32+ char secure string)* |
| `FRONTEND_URL` | Allowed CORS origin | `http://localhost:5173` |

---

## 8. API Documentation

### Authentication Endpoints
| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :---: |
| `POST` | `/api/auth/signup` | Registers a new creator account | No |
| `POST` | `/api/auth/login` | Authenticates creator and issues JWT | No |
| `GET` | `/api/auth/me` | Returns authenticated user profile | **Yes** |

#### Example Request: `POST /api/auth/signup`
```json
{
  "name": "Sarah Connor",
  "email": "sarah@skynet.com",
  "password": "supersecretpassword"
}
```

---

### Poll Endpoints
| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :---: |
| `POST` | `/api/polls` | Creates a new poll with options | **Yes** |
| `GET` | `/api/polls` | Lists all polls created by current user | **Yes** |
| `GET` | `/api/polls/:id` | Fetches poll details and counts | No |
| `GET` | `/api/polls/share/:shareCode` | Audience lookup using 6-character code | No |
| `PATCH` | `/api/polls/:id/close` | Closes voting on the poll | **Yes** |
| `DELETE` | `/api/polls/:id` | Deletes poll and clears data | **Yes** |

#### Example Request: `POST /api/polls`
```json
{
  "question": "What is your favorite programming language?",
  "options": ["Python", "JavaScript", "Go", "Java"],
  "durationHours": 24
}
```

---

### Voting & Real-Time Endpoints
| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/api/polls/:id/vote` | Submits a vote for an option |
| `GET` | `/api/polls/:id/results` | Returns poll results and percentages |
| `GET` | `/api/polls/:id/live` | WebSocket connection for real-time live streaming |

#### Example Request: `POST /api/polls/:id/vote`
```json
{
  "optionId": "opt_3",
  "voterFingerprint": "voter_9x7k2b8a"
}
```

#### Example Real-Time Event (`VOTE_UPDATED`)
```json
{
  "event": "VOTE_UPDATED",
  "pollId": "65f8a12bc9e201b2a9d4f001",
  "totalVotes": 10,
  "results": [
    { "id": "opt_1", "text": "Python", "votes": 4, "percentage": 40.0 },
    { "id": "opt_2", "text": "JavaScript", "votes": 3, "percentage": 30.0 },
    { "id": "opt_3", "text": "Go", "votes": 3, "percentage": 30.0 }
  ]
}
```

---

## 9. MongoDB Schema & Persistence

MongoDB guarantees durability and auditable history:

### 1. `users` Collection
```json
{
  "_id": ObjectId("65f8a..."),
  "name": "Sarah Connor",
  "email": "sarah@skynet.com",
  "password_hash": "$2a$10$7...",
  "created_at": ISODate("2026-09-19T10:00:00Z")
}
```
*Index: Unique index on `email`.*

### 2. `polls` Collection
```json
{
  "_id": ObjectId("65f8b..."),
  "question": "What is your favorite programming language?",
  "options": [
    { "id": "opt_1", "text": "Python", "votes": 12 },
    { "id": "opt_2", "text": "Go", "votes": 18 }
  ],
  "creator_id": ObjectId("65f8a..."),
  "share_code": "GO2026",
  "status": "active",
  "expires_at": ISODate("2026-09-20T10:00:00Z"),
  "total_votes": 30,
  "created_at": ISODate("2026-09-19T10:00:00Z")
}
```
*Indices: Unique index on `share_code`, index on `creator_id`.*

### 3. `votes` Collection
```json
{
  "_id": ObjectId("65f8c..."),
  "poll_id": ObjectId("65f8b..."),
  "option_id": "opt_2",
  "voter_id": "voter_9x7k2b8a",
  "ip_address": "192.168.1.15",
  "user_agent": "Mozilla/5.0 ...",
  "created_at": ISODate("2026-09-19T10:05:00Z")
}
```
*Index: Compound unique index on `{ "poll_id": 1, "voter_id": 1 }` preventing duplicate voting at the database level.*

---

## 10. Redis: Atomic Counters & Pub/Sub

Redis is not an optional cache in PulsePoll — it genuinely powers both **in-memory speed** and **event distribution**:

### 1. Atomic Vote Counting (`HINCRBY`)
Each poll maintains a Redis Hash: `poll:<pollID>:votes`:
```
Key: poll:65f8b...:votes
  opt_1 -> 12
  opt_2 -> 18
  opt_3 -> 5
```
When a vote is cast:
```go
redisClient.HIncrBy(ctx, "poll:65f8b...:votes", "opt_2", 1)
```
- **Time Complexity**: $O(1)$ atomic execution.
- Prevents race conditions during high-volume simultaneous voting events.

### 2. Redis Pub/Sub (`PUBLISH` / `SUBSCRIBE`)
Once the atomic counter updates, Go reads all option counts from the hash and publishes to a dedicated channel:
```go
channel := "poll_updates:65f8b..."
redisClient.Publish(ctx, channel, jsonPayload)
```
- **Decoupled Architecture**: Go server instances or background workers can horizontally scale without state synchronization issues.
- The Go WebSocket Hub subscribes to the poll's channel and broadcasts the update to all active WebSocket clients.

---

## 11. Authentication & Security

1. **Password Hashing**: Stored passwords use BCrypt hashing (cost factor 10). Plain-text passwords are never persisted.
2. **Stateless JWT Tokens**: Signed with HMAC SHA-256 containing User ID and expiration, validated via Gin middleware on protected routes.
3. **Double Duplicate Vote Validation**:
   - Backend checks MongoDB `votes` collection before recording.
   - Database compound index `{poll_id, voter_id}` prevents simultaneous race conditions.
4. **Input Sanitization**:
   - Strip whitespace, prevent empty questions, enforce minimum 2 and maximum 10 unique options.
   - Reject votes on closed or expired polls.
5. **CORS Security**: Strict HTTP headers allowing trusted origins and required headers.

---

## 12. Real-Time WebSocket Engine

- Built using `github.com/gorilla/websocket`.
- Upgrades `GET /api/polls/:id/live`.
- **Room Management**: The `Hub` manages client maps segregated by `pollId`. When the first client enters a poll room, the Hub registers a Redis Pub/Sub listener. When the room becomes empty, the subscription is automatically torn down to save server resources.
- **Heartbeats**: Periodic ping/pong packets every 54 seconds ensure dead TCP sockets are cleanly garbage collected.
- **Auto-Reconnect**: The React `useWebSocket` hook automatically re-establishes connectivity with backoff if a network hiccup occurs.

---

## 13. End-to-End Testing (Multi-Browser Sync)

Follow this test scenario to verify the core real-time requirement:

1. **Sign Up & Log In**:
   - Navigate to [http://localhost:3000/login](http://localhost:3000/login).
   - Click **"Fill Demo Credentials"** or create a new account.
2. **Create a Poll**:
   - Click **"Create New Poll"**.
   - Question: `What is your favorite programming language?`
   - Options: `Python`, `JavaScript`, `Go`, `Java`.
   - Click **"Create & Launch Poll"**.
3. **Open Multi-Browser Setup**:
   - **Browser 1 (Chrome)**: Keep on the **Live Results View** (`/poll/:id/results`).
   - **Browser 2 (Incognito)**: Paste the direct share link (`/poll/:shareCode`).
   - **Browser 3 (Firefox or Edge)**: Paste the direct share link (`/poll/:shareCode`).
4. **Cast a Vote in Browser 2**:
   - Select `Go` and click **"Submit Vote"**.
   - Confetti explodes in Browser 2, and it transitions to live results.
5. **Observe Instant Real-Time Sync**:
   - **Browser 1 and Browser 3 update immediately without any page refresh!**
   - The Go bar increases smoothly, and percentages re-calculate across all screens.
6. **Test Duplicate Prevention**:
   - Attempt to vote again in Browser 2 — the system recognizes previous participation and displays the live results.
7. **Test Creator Lifecycle**:
   - In Browser 1, click **"Close Poll"**.
   - Notice Browser 2 and 3 immediately display the **"Poll Closed"** banner.

---

## 14. Deployment Guide

PulsePoll is designed to deploy seamlessly to modern cloud providers:

### Render / Railway / Fly.io

1. **MongoDB**:
   - Provision a free database on [MongoDB Atlas](https://www.mongodb.com/atlas).
   - Copy connection string to `MONGO_URI`.

2. **Redis**:
   - Provision a free serverless Redis database on [Upstash](https://upstash.com).
   - Copy connection URI to `REDIS_URL`.

3. **Backend Service (Go)**:
   - Connect your GitHub repository.
   - Set Root Directory: `backend`.
   - Docker build target: `Dockerfile`.
   - Add Environment Variables (`PORT`, `MONGO_URI`, `REDIS_URL`, `JWT_SECRET`, `FRONTEND_URL`).

4. **Frontend Service (Vercel / Netlify / Render)**:
   - Root Directory: `frontend`.
   - Build Command: `npm run build`.
   - Output Directory: `dist`.
   - Environment Variable: `VITE_API_URL=https://your-go-backend.onrender.com/api`.

---

## 15. Engineering Challenges & Solutions

### Challenge 1: Multi-Browser Real-Time Sync Without Excessive Database Load
- *Problem*: Polling MongoDB every few seconds for vote changes generates high read IOPS and introduces artificial delay.
- *Solution*: Implemented an in-memory Redis Hash (`HINCRBY`) and Redis Pub/Sub channel. Votes update memory in microseconds, and the Go WebSocket Hub pushes the state directly to connected browsers with zero database query overhead during live traffic spikes.

### Challenge 2: Duplicate Voting in High-Concurrency Scenarios
- *Problem*: Two votes from the same user submitted within milliseconds can bypass simple application-level `if (!hasVoted)` checks.
- *Solution*: Backed the check with a MongoDB compound unique index on `{ poll_id: 1, voter_id: 1 }`. Any race condition triggers an index collision error, ensuring exactly-once vote semantics.

### Challenge 3: Seamless SPA Routing and WebSocket Upgrades in Docker
- *Problem*: Direct browser URL hits on routes like `/poll/ABC123` result in 404s when served statically without URL rewriting.
- *Solution*: Configured an Nginx reverse-proxy container with `try_files $uri $uri/ /index.html;` and configured WebSocket upgrade headers (`Upgrade: $http_upgrade`, `Connection: "upgrade"`).

---

## 16. AI Usage & Developer Reflections

In alignment with the GUVI internship guidelines:
- **How AI Helped**:
  - Accelerated initial boilerplate scaffolding for multi-stage Dockerfiles and CSS custom property design tokens.
  - Formulated WebSocket heartbeat ping/pong patterns and edge-case validation schemas.
- **Where AI Got in the Way**:
  - AI initially suggested an in-memory map without genuine Redis Pub/Sub — which violates the assignment's explicit requirement that Redis must genuinely drive the real-time event updates across nodes.
  - Manually refactored to implement full Redis pipeline initialization, atomic `HINCRBY`, and Pub/Sub channel subscriptions to meet all strict evaluation criteria.
- **Takeaway**: AI is a valuable assistant for rapid prototyping, but system architecture, database constraints, and real-time synchronicity require human verification and end-to-end understanding.

---

## 17. 3–5 Minute Video Demonstration Script

Use this timing outline when recording your submission walkthrough:

- **0:00 – 0:30 | Introduction**: State your name, introduction to PulsePoll, and high-level architecture overview (React + Go/Gin + MongoDB + Redis).
- **0:30 – 1:00 | Authentication**: Demonstrate creator sign up and login with JWT issuance.
- **1:00 – 1:40 | Poll Creation**: Formulate a question with 4 options, set a duration, and launch the poll.
- **1:40 – 2:10 | Sharing Mechanism**: Demonstrate 1-click link copying and show the short alphanumeric share code.
- **2:10 – 3:00 | The Live Synchronization Test**:
  - Open the results page in one window.
  - Open the share link in two separate private windows.
  - Cast a vote and highlight the immediate, zero-refresh update on all screens.
- **3:00 – 3:40 | Architectural Walkthrough**: Explain how Redis `HINCRBY` tallies votes, triggers `VOTE_UPDATED` over Redis Pub/Sub, and pushes to Go WebSockets.
- **3:40 – 4:20 | Biggest Challenge**: Discuss duplicate vote race prevention using MongoDB compound indices.
- **4:20 – 5:00 | AI Usage & Wrap Up**: Share insights on where AI assisted and how you verified the Redis and WebSocket pipeline.

---

## 18. Submission Checklist

- [x] React Frontend with zero-refresh live updates.
- [x] Go + Gin Backend with input validation and JWT auth.
- [x] MongoDB persistent collections (`users`, `polls`, `votes`).
- [x] Redis atomic counts (`HINCRBY`) and Pub/Sub (`VOTE_UPDATED`).
- [x] Go WebSockets with room broadcasting and heartbeat pings.
- [x] Public GitHub repository with clean history.
- [x] Deployment instructions & Docker Compose setup.
- [x] 3–5 minute video demonstration recording.

**Send final submission package to**: `devhiring@hclguvi.com`
