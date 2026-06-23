# Backend Development Progress

This document tracks the progress of backend feature scaffolding and development, establishing a roadmap and logging what has been implemented.

## Implemented So Far

The backend architecture is structured around standard Go conventions (`models/`, `services/`, and `handlers/`) without the use of an ORM. We use pure SQL to ensure query efficiency and type safety.

### Infrastructure & Setup
- **Docker Compose:** Configured and running PostgreSQL, Redis, and Meilisearch locally.
- **Server:** Running `chi` router on port `8080`.
- **Authentication Middleware:** Built `ClerkMiddleware` and `RequireAuth` which parse Clerk JWTs from the `Authorization: Bearer <token>` header, verifying them against the Clerk secret key, and injecting the Clerk User ID into the request context.
- **Redis-Backed Rate Limiting:** Implemented `go-chi/httprate` and `go-chi/httprate-redis` middleware to protect routes. Tiered limits: 100 requests/minute for reads (Tier 1), and 10 requests/minute for mutations/writes (Tier 1 writes, Tiers 2 & 3).

### 1. Users Feature
**Endpoints:**
- `POST /users/sync`: Upserts the user record. This allows our backend to stay in sync with the frontend when a user signs up/in via Clerk.
- `GET /users/me`: Retrieves the current user profile.
- `PATCH /users/me`: Applies a partial update to the user profile.

**Key Components:**
- `models/user.go`: Maps the database schema and structures HTTP requests. Supports user-specific search coordinates (`latitude`, `longitude`) and preferences (`search_radius`).
- `services/users.go`: SQL statements utilizing `ON CONFLICT` for robust upserts.
- `handlers/users.go`: Handler mapping utilizing helper functions (`respondJSON`, `respondError`).

### 2. Events Feature
**Endpoints:**
- `POST /events`: Creates a new event. The system automatically associates the created event with the organizer's database UUID by mapping it from the Clerk ID.
- `GET /events`: Lists upcoming, published events with filtering, pagination, and sorting.
- `GET /events/{id}`: Fetches event details.
- `PATCH /events/{id}`: Partially updates event details (restricted to organizer or admin).
- `DELETE /events/{id}`: Soft deletes/cancels an event (restricted to organizer or admin).

**Key Components:**
- `models/event.go`: Full mapping of the Events table and properties, including `banner_url` and filters.
- `services/events.go`: Core domain logic including ownership checks (`GetEventOwner`) and dynamic query-building.
- **Geo-location Proximity:** Employs the Haversine formula in pure SQL to fetch events within a given radius. Integrates with the calling user's profile location preferences as fallback defaults.
- **Dynamic Filters:** Pagination (`page`/`limit`/`offset`), age restriction boundaries (`age_min`/`age_max`), date ranges (`from`/`to`), and interests tags. Enforces a maximum page limit of 50.
- **Sorting:** Supports `upcoming` (starts_at ASC), `distance` (nearest first), and `trending` (rsvp_count DESC).

### 3. Communities Feature
**Endpoints:**
- `POST /communities`: Creates a new community. The creator is automatically added as an `'admin'` member.
- `GET /communities`: Lists public, active communities with filtering, pagination, and sorting.
- `GET /communities/{id}`: Fetches details of a specific community.
- `PATCH /communities/{id}`: Partially updates community details (restricted to owner or admin).
- `DELETE /communities/{id}`: Soft deactivates a community (restricted to owner or admin).
- `POST /communities/{id}/join`: Joins a community (atomically manages transactions and updates membership count).
- `DELETE /communities/{id}/leave`: Leaves a community.
- `GET /communities/{id}/members`: Lists community members with joining history and roles.

**Key Components:**
- `models/community.go`: Defines the community models, including `image_url` and community members.
- `services/communities.go`: Database operations for community creation, membership joins/leaves, and ownership validation (`GetCommunityOwner`).
- **Geo-location Proximity:** Supports filtering communities by latitude/longitude and radius.
- **Filters & Sorting:** Supports pagination, age-restrictions, and sorting (`newest`, `distance`, `trending`).

### 4. RSVPs Feature
**Endpoints:**
- `POST /events/{id}/rsvp`: Joins/RSVPs to an event (accepts status: `pending`, `confirmed`, `cancelled`, or `waitlisted`).
- `DELETE /events/{id}/rsvp`: Cancels an existing RSVP (sets status to `cancelled`).
- `GET /events/{id}/rsvps`: Lists all active/confirmed RSVPs for a given event, including attendee profiles.
- `GET /users/me/rsvps`: Lists all RSVPs the current user has created, including event summaries.

**Key Components:**
- `models/rsvp.go`: Mirror definitions of the `rsvps` table and joint representations.
- `services/rsvps.go`: Database operations for creating, editing, and querying event/user RSVPs. Includes database transaction mapping to update event `rsvp_count` values dynamically and broadcast changes to Redis. Triggers organizer push notifications upon RSVP confirmation.
- `handlers/rsvps.go`: Exposes RSVP operations and handles response envelopes.

### 5. Interests Feature
**Endpoints:**
- `GET /interests`: Returns all system-wide available topics and categories.
- `POST /interests`: Adds a new topic/category to the system.
- `GET /users/me/interests`: Returns interests associated with the current user's profile.
- `POST /users/me/interests`: Replaces/updates the list of interests associated with the user, including proficiency levels (`beginner`, `intermediate`, `regular`, `expert`).
- `GET /events/{id}/interests`: Gets interests tagged to a specific event.
- `POST /events/{id}/interests`: Replaces/tags interests to an event.

**Key Components:**
- `models/interest.go`: Models mapping Interests, User Interests, and Event Interests tables.
- `services/interests.go`: Uses SQL transactions to clear and replace pivot relations safely.
- `handlers/interests.go`: Handles user and event interest updates.

### 6. Meilisearch Sync & Search
**Endpoints:**
- `GET /search?q=query&type=all|events|communities&interest_id=uuid&lat=x&lng=y&radius=km`: Multi-index search query with filtering and proximity sorting.
- `GET /search/autocomplete?q=query&type=all|events|communities`: Low-latency suggestion matches returning top 5 recommendations.

**Key Components:**
- `services/search.go`: Native HTTP Meilisearch client wrapper that manages index settings configuration (`ConfigureSettings`), queries (`Search`, `Autocomplete`), document upserts (`SyncEvent`, `SyncCommunity`), and deletions.
- **Asynchronous Syncing:** Created, updated, or deleted events/communities automatically sync to Meilisearch in a background goroutine.
- **Advanced Filtering and Proximity Sorting:**
  - Index settings are configured dynamically on boot to enable filtering/sorting on `interests`, `status`, `visibility`, and `_geo`.
  - Supports category matching using `interest_id` and spatial queries using `_geoRadius` and `_geoPoint:asc` proximity sorting.

### 7. Media Uploads (Cloudflare R2 / S3)
**Endpoints:**
- `POST /upload/avatar`: Generates a presigned Cloudflare R2 / S3 upload PUT URL for profile avatars.
- `POST /upload/event-banner`: Generates a presigned upload URL for event banners (restricted to organizer/admin).
- `POST /upload/community-image`: Generates a presigned upload URL for community cover images (restricted to organizer/admin).

**Key Components:**
- `models/upload.go`: Structures request metadata (`content_type`) and response outputs (`upload_url`, `public_url`).
- `services/storage.go`: S3-compatible service wrapper using standard AWS SDK v2, supporting configuration overrides for Cloudflare R2, AWS S3, or local MinIO.
- `handlers/upload.go`: Authenticates and generates unique object storage keys (with extensions mapped from the content-type) returning presigned PUT URLs, preventing file upload bottlenecks on the backend app server.

### 8. WebSocket Server (Realtime)
**Endpoints:**
- `WS /ws?token=<jwt>`: Secure WebSocket handshake upgrade connection endpoint (listens on separate port `8081`).

**Key Components:**
- `cmd/websocket/main.go`: Separate deployable Go WebSocket compute executable.
- **Presence Tracking:** Client connections automatically set and refresh Redis presence keys (`presence:user:<user_id>`). They also periodically run heartbeats to write `last_seen_at = NOW()` to PostgreSQL.
- **RSVP Broadcaster:** Subscribes to Redis Pub/Sub `event:rsvp_updates` to broadcast live rsvp counts to clients that subscribed (`subscribe_event`).
- **Community presence:** Allows users to subscribe (`subscribe_community`) to get the initial list of online community members (using optimized `MGet` sets) and receive real-time connection status change broadcasts (`presence_change`).

### 9. Push Notifications (FCM)
**Endpoints:**
- `POST /users/me/fcm-token`: Registers/updates the user's active FCM device token.
- `POST /communities/{id}/announcements`: Broadcasts community announcements via push notification to all members (restricted to owner/admin).

**Key Components:**
- `services/notification.go`: Integrates with Firebase Admin Go SDK to manage multicast push alerts. Supports automated token cleanup on client failures and defaults to a mock console logger if Firebase account credentials are unset.
- **RSVP Alerting:** Sends a push notification to event organizers whenever a new user confirms their RSVP.
- **Nearby activity alerts:** Triggers a geo-targeted spatial SQL query (using Haversine calculations) when events are published, pushing alerts to all local matched users who share tagged interests within their custom search radius.
- **Background Reminders Worker:** Runs a background loop ticking every 5 minutes in a separate goroutine to send reminders to confirmed RSVP attendees 24h and 1h before the event's starts_at time, logging completions in the database to prevent duplicate alerts.

---

## Complete API Route Map

All endpoints (except `/health`) are protected via Clerk JWT verification (using Clerk Bearer tokens).

```
GET    /health

# Users
POST   /users/sync
GET    /users/me
PATCH  /users/me
GET    /users/me/rsvps
GET    /users/me/interests
POST   /users/me/interests
POST   /users/me/fcm-token

# Events
POST   /events
GET    /events
GET    /events/{id}
PATCH  /events/{id}
DELETE /events/{id}
POST   /events/{id}/rsvp
DELETE /events/{id}/rsvp
GET    /events/{id}/rsvps
GET    /events/{id}/interests
POST   /events/{id}/interests

# Communities
POST   /communities
GET    /communities
GET    /communities/{id}
PATCH  /communities/{id}
DELETE /communities/{id}
POST   /communities/{id}/join
DELETE /communities/{id}/leave
GET    /communities/{id}/members
POST   /communities/{id}/announcements

# Interests
GET    /interests
POST   /interests

# Search
GET    /search
GET    /search/autocomplete

# Media Uploads
POST   /upload/avatar
POST   /upload/event-banner
POST   /upload/community-image

# WebSocket Realtime (Separate Port 8081)
WS     /ws?token=<token>
```
