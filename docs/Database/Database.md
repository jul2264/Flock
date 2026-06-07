#### Database Design

a.	users 
Central to everything. Stores Clerk's external clerk_id alongside profile data, age, location (lat/lng + city/neighborhood), search radius preference, and a role enum (user, organizer, admin) for RBAC. last_seen_at drives online presence tracking.

b. interests 
A self-referencing taxonomy (e.g. "Sports" → "Football", "Arts" → "Photography"). The parent_id FK enables two-level category/subcategory filtering without duplicating data.

c. user_interests 
Junction table linking users to their interest tags, with an optional proficiency_level for the matching algorithm (e.g. beginner vs. regular).

d. communities 
Groups/recurring squads. Has its own geolocation, age range, and visibility (public, private, invite_only). is_recurring + recurrence_rule (iCal RRULE format) handles recurring meetups like weekly football groups. max_members enforces limits.

e. community_members 
Junction with roles (owner, moderator, member) and status (active, banned, pending). left_at is nullable so you can track churn.

f. events 
Links to both an organizer and optionally a community. Stores full venue info including google_place_id for Maps integration. rsvp_count is a denormalized counter (updated via trigger or app logic) so you avoid expensive COUNT queries on hot event feeds. status enum: draft, published, cancelled, completed.

g. event_interests 
Many-to-many between events and interests, enabling multi-tag filtering (e.g. "show events tagged Photography AND Outdoor").

h. rsvps 
Tracks RSVP lifecycle: pending, confirmed, cancelled, waitlisted. attended bool is set post-event for analytics.

i. messages 
Community chat. reply_to_id is a self-referencing FK for threaded replies. media_urls is a text array for multi-image messages. Soft-delete via is_deleted rather than physical deletion, to preserve thread continuity.

j. notifications 
Covers all push/in-app notification types (RSVP confirmed, event reminder, community announcement, nearby alert). ref_entity_id + ref_entity_type is a polymorphic pointer (e.g. event, community) so you can deep-link from the notification.

k. reports 
Polymorphic moderation table. target_type + target_id handles reporting users, events, or communities from a single table. resolved_by FK points back to an admin user.


#### Database related documentation

Geolocation on both users and events/communities. Distance filtering happens at query time using PostGIS ST_DWithin — store lat/lng as float and add a generated geography column, or use PostGIS POINT type directly.

Meilisearch as the search layer. Keep Postgres as source of truth; sync events/communities/users to Meilisearch on write. Don't try to do fuzzy full-text search in Postgres.

Redis for the hot path. Cache: nearby event feed per user (TTL ~2 min), live RSVP counts, session data from Clerk, online presence bitmaps.

RSVP count denormalization. The rsvp_count column on events avoids a full scan on every feed render. Increment/decrement it transactionally when RSVPs change.

Age filtering. Store date_of_birth on users, compute age at query time — never store age directly. Events/communities store age_min/age_max for range filtering.