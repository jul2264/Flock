# Backend Development Progress

This document tracks the progress of backend feature scaffolding and development, establishing a roadmap and logging what has been implemented.

## Implemented So Far

The backend architecture is structured around standard Go conventions (`models/`, `services/`, and `handlers/`) without the use of an ORM. We use pure SQL to ensure query efficiency and type safety.

### Infrastructure & Setup
- **Docker Compose:** Configured and running PostgreSQL, Redis, and Meilisearch locally.
- **Server:** Running `chi` router on port `8080`.
- **Authentication Middleware:** Built `ClerkMiddleware` and `RequireAuth` which parse Clerk JWTs from the `Authorization: Bearer <token>` header, verifying them against the Clerk secret key, and injecting the Clerk User ID into the request context.

### 1. Users Feature
**Endpoints:**
- `POST /users/sync`: Upserts the user record. This allows our backend to stay in sync with the frontend when a user signs up/in via Clerk.
- `GET /users/me`: Retrieves the current user profile.
- `PATCH /users/me`: Applies a partial update to the user profile.

**Key Components:**
- `models/user.go`: Maps the database schema and structures HTTP requests.
- `services/users.go`: SQL statements utilizing `ON CONFLICT` for robust upserts.
- `handlers/users.go`: Handler mapping utilizing helper functions (`respondJSON`, `respondError`).

### 2. Events Feature
**Endpoints:**
- `POST /events`: Creates a new event. The system automatically associates the created event with the organizer's database UUID by mapping it from the Clerk ID found in the request context.
- `GET /events`: Lists upcoming, published events.
- `GET /events/{id}`: Fetches event details.

**Key Components:**
- `models/event.go`: Full mapping of the Events table and properties.
- `services/events.go`: Core domain logic converting the Clerk context into a proper relational link.
- `handlers/events.go`: API handlers enforcing required parameters.

### 3. Communities Feature
**Endpoints:**
- `POST /communities`: Creates a new community.
- `GET /communities`: Lists public, active communities.
- `GET /communities/{id}`: Fetches details of a specific community.

**Key Components:**
- `models/community.go`: Defines the community models.
- `services/communities.go`: Core domain logic for fetching and persisting communities.
- `handlers/communities.go`: Handles request payloads, validating properties such as community name.

### 4. RSVPs Feature
**Endpoints:**
- `POST /events/{id}/rsvp`: Joins/RSVPs to an event (accepts status: `pending`, `confirmed`, `cancelled`, or `waitlisted`).
- `DELETE /events/{id}/rsvp`: Cancels an existing RSVP (sets status to `cancelled`).
- `GET /events/{id}/rsvps`: Lists all active/confirmed RSVPs for a given event, including attendee profiles.
- `GET /users/me/rsvps`: Lists all RSVPs the current user has created, including event summaries.

**Key Components:**
- `models/rsvp.go`: Mirror definitions of the `rsvps` table and joint representations.
- `services/rsvps.go`: Database operations for creating, editing, and querying event/user RSVPs.
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
- `GET /search?q=query&type=all|events|communities`: Multi-index search query utilizing Meilisearch.

**Key Components:**
- `services/search.go`: Native HTTP Meilisearch client wrapper that manages index queries and document upserts.
- **Asynchronous Syncing:** Created events and communities automatically trigger indexing into Meilisearch inside a lightweight goroutine.

---

## Complete API Route Map

All core endpoints are protected via Clerk JWT verification (using Clerk Bearer tokens).

```
GET    /health
POST   /users/sync
GET    /users/me
PATCH  /users/me
GET    /users/me/rsvps
GET    /users/me/interests
POST   /users/me/interests
POST   /events
GET    /events
GET    /events/{id}
POST   /events/{id}/rsvp
DELETE /events/{id}/rsvp
GET    /events/{id}/rsvps
GET    /events/{id}/interests
POST   /events/{id}/interests
POST   /communities
GET    /communities
GET    /communities/{id}
GET    /interests
POST   /interests
GET    /search
```
