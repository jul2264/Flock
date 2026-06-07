#### System Architecture

C4 Level 1 · Context
Flock as a black box: three actor types (mobile users, web users, event organisers) interacting with it, and four external systems it depends on (Clerk, Google Maps, FCM, Meilisearch).

C4 Level 2 · Container 
The major deployable units inside Flock's boundary: the React Native mobile app and Next.js web app as clients, the Go API server and WebSocket server as the compute layer, and PostgreSQL + Redis + Meilisearch as the data tier — with protocols shown.

C4 Level 3 · Component 
The internal structure of the Go API server broken into three columns: request handlers (auth middleware, event/user/search/chat handlers), domain services (auth, event, location, notification, search), and repositories (user, event, community, session cache).

Deployment 
The infrastructure view: a cloud VPC with a public subnet running Dockerised containers (Go API, WebSocket, Meilisearch) and a private subnet for managed data stores, with CDN/edge in front for Next.js static assets and external SaaS services on the right.

Sequence 
The end-to-end flow for the core use case: a user searches for nearby events (cache lookup, Meilisearch search, cache write) and RSVPs (DB insert + FCM push to organiser), showing request vs. response arrows.

Data flow 
How data moves from four source types through five processing stages (Zod validation, auth, geo-filtering, interest matching, Redis cache) into three storage destinations and two external output services.