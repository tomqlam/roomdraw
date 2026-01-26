# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Digital Room Draw application for Harvey Mudd College that encodes all room draw rules and facilitates the room selection process for students.

## Development Commands

### Frontend (React)
```bash
cd frontend
npm install          # Install dependencies
npm start            # Run dev server on localhost:3000
npm test             # Run tests
npm run build        # Production build
```

### Backend (Go + Docker)
```bash
cd backend
# Development with live reload:
docker build -t roomdraw-backend .
docker run -it -p 8080:8080 -v $(pwd):/app roomdraw-backend

# Production (uses podman):
podman build -t roomdraw-backend --build-arg ENV=production .
podman run -it -p 8080:8080 -v $(pwd):/app roomdraw-backend
```

### Database Setup
```bash
cd database
# Requires .env with SQL_PASS, SQL_IP, SQL_DB_NAME, SQL_USER
python createAllTables.py   # Creates all tables (drops existing first)
python createDorms.py       # Populates dorm/room data from JSON files
```

## Architecture

### Backend Structure (Go/Gin)
- `cmd/server/main.go` - Entry point, route definitions, middleware setup
- `pkg/handlers/` - Request handlers organized by domain:
  - `room_handler.go` - Core room operations (largest file, ~138K)
  - `suite_handler.go` - Suite management
  - `user_handler.go` - User operations
  - `room_draw_logic.go` - Priority comparison and sorting algorithms
  - `frosh_handler.go` - Freshman room management
  - `admin_handler.go` - Admin-only operations
- `pkg/middleware/` - JWT auth, request queue, blocklist checking
- `pkg/models/types.go` - All domain types and database models
- `pkg/config/config.go` - Environment configuration

### Frontend Structure (React)
- `src/App.js` - Main component with dorm selection, user search
- `src/MyContext.js` - Global state (rooms, users, credentials, settings)
- `src/FloorGrid.js` - Room grid visualization
- `src/BumpModal.js` - Room pull/bump interaction modal
- `src/Search/` - Search functionality
- `src/Admin/` - Admin-only components

### Database (PostgreSQL)
Tables: Users, Rooms, Suites, SuiteGroups, user_rate_limits, transaction_logs

Key relationships:
- Rooms belong to Suites (suite_uuid foreign key)
- Users can belong to SuiteGroups (sgroup_uuid)
- Rooms track occupants as integer array of user IDs
- PullPriority stored as JSONB on Rooms

## Domain Concepts

### Pull Priority System
Priority determines who can claim or bump rooms. Hierarchy (highest to lowest):
1. Lock pull (PullType=3) - highest effective year (6)
2. In-dorm seniors (effective year 5)
3. Seniors (year 4) → Juniors (year 3) → Sophomores (year 2)
4. Within same year: lower draw number wins
5. Preplaced users have special handling

### Pull Types
- 0: Undefined
- 1: Self (pulling for yourself)
- 2: Normal pull (pulling others into your room)
- 3: Lock pull (guaranteed room from previous year)
- 4: Alternative pull

### API Route Groups
Routes are split by authorization requirements:
- **Read routes**: GET endpoints, optional JWT
- **Write routes**: POST endpoints with request queue (serializes writes), JWT required, blocklist check
- **Admin routes**: POST endpoints requiring admin JWT (checked via `adminList` in MyContext.js)

### Environment Variables
Backend requires `.env` file with:
- `SQL_PASS`, `SQL_IP`, `SQL_DB_NAME`, `SQL_USER`, `SQL_PORT`, `USE_SSL`
- `REQUIRE_AUTH` - set to "True" to enable authentication
- `BUNNYNET_*` - CDN configuration for images
- `EMAIL_USERNAME`, `EMAIL_PASSWORD` - for notifications
