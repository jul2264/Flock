package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/jul2264/Flock/backend/internal/db"
)

// WebSocket Message Types (Client-to-Server)
type ClientMessage struct {
	Action      string `json:"action"`
	EventID     string `json:"event_id,omitempty"`
	CommunityID string `json:"community_id,omitempty"`
}

// WebSocket Message Types (Server-to-Client)
type ServerMessage struct {
	Type        string   `json:"type"`
	EventID     string   `json:"event_id,omitempty"`
	RSVPCount   *int     `json:"rsvp_count,omitempty"`
	CommunityID string   `json:"community_id,omitempty"`
	UserID      string   `json:"user_id,omitempty"`
	Status      string   `json:"status,omitempty"`
	OnlineUsers []string `json:"online_users,omitempty"`
}

type RedisPresenceMsg struct {
	UserID string `json:"user_id"`
	Status string `json:"status"` // "online" or "offline"
}

type RedisRSVPMsg struct {
	EventID   string `json:"event_id"`
	RSVPCount int    `json:"rsvp_count"`
}

type Client struct {
	conn                  *websocket.Conn
	userID                string
	subscribedEvents      map[string]bool
	subscribedCommunities map[string]bool
	mu                    sync.Mutex
}

func (c *Client) sendJSON(val interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(val)
}

type Server struct {
	db              *sql.DB
	redisClient     *redis.Client
	upgrader        websocket.Upgrader
	clients         map[string]*Client // key: user_id
	clientsMu       sync.RWMutex
	eventSubs       map[string]map[string]*Client // event_id -> user_id -> Client
	eventSubsMu     sync.RWMutex
	communitySubs   map[string]map[string]*Client // community_id -> user_id -> Client
	communitySubsMu sync.RWMutex
}

func NewServer(database *sql.DB, rds *redis.Client) *Server {
	return &Server{
		db:          database,
		redisClient: rds,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients:       make(map[string]*Client),
		eventSubs:     make(map[string]map[string]*Client),
		communitySubs: make(map[string]map[string]*Client),
	}
}

func main() {
	// Try loading .env from various directory depths
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	// Set Clerk Secret Key
	clerk.SetKey(os.Getenv("CLERK_SECRET_KEY"))

	// Connect to database
	database := db.Connect()
	defer database.Close()

	// Connect to Redis
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}
	redisOpts, err := redis.ParseURL(redisURL)
	var redisClient *redis.Client
	if err == nil {
		redisClient = redis.NewClient(redisOpts)
		log.Println("Connected to Redis successfully!")
		defer redisClient.Close()
	} else {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	server := NewServer(database, redisClient)

	// Start Redis Pub/Sub subscribers
	go server.subscribeRSVPUpdates()
	go server.subscribePresenceUpdates()

	http.HandleFunc("/ws", server.handleConnection)

	port := os.Getenv("WEBSOCKET_PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("🚀 WebSocket server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (s *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	// 1. Authenticate with Clerk JWT via Query Parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "unauthorized: missing token", http.StatusUnauthorized)
		return
	}

	claims, err := jwt.Verify(r.Context(), &jwt.VerifyParams{
		Token: token,
	})
	if err != nil {
		http.Error(w, "unauthorized: invalid or expired token", http.StatusUnauthorized)
		return
	}

	clerkID := claims.Subject

	// 2. Fetch User UUID from Postgres
	var userID string
	err = s.db.QueryRowContext(r.Context(), `SELECT id FROM users WHERE clerk_id = $1`, clerkID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "forbidden: user profile not synced", http.StatusForbidden)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// 3. Upgrade HTTP connection
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := &Client{
		conn:                  conn,
		userID:                userID,
		subscribedEvents:      make(map[string]bool),
		subscribedCommunities: make(map[string]bool),
	}

	// 4. Register client
	s.registerClient(client)

	// Create heartbeat context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 5. Handle reader and write loop / heartbeats
	go s.runPresenceHeartbeats(ctx, client)
	s.readPump(client)
}

func (s *Server) registerClient(client *Client) {
	s.clientsMu.Lock()
	if old, exists := s.clients[client.userID]; exists {
		old.conn.Close()
	}
	s.clients[client.userID] = client
	s.clientsMu.Unlock()

	// Set Redis presence key
	ctx := context.Background()
	presenceKey := fmt.Sprintf("presence:user:%s", client.userID)
	if err := s.redisClient.Set(ctx, presenceKey, "online", 70*time.Second).Err(); err != nil {
		log.Printf("Failed to set Redis presence key for user %s: %v", client.userID, err)
	}

	// Publish presence update
	s.publishPresenceChange(client.userID, "online")
}

func (s *Server) deregisterClient(client *Client) {
	s.clientsMu.Lock()
	if current, exists := s.clients[client.userID]; exists && current == client {
		delete(s.clients, client.userID)
	}
	s.clientsMu.Unlock()

	// Unsubscribe from events
	s.eventSubsMu.Lock()
	for eventID := range client.subscribedEvents {
		if subs, ok := s.eventSubs[eventID]; ok {
			delete(subs, client.userID)
			if len(subs) == 0 {
				delete(s.eventSubs, eventID)
			}
		}
	}
	s.eventSubsMu.Unlock()

	// Unsubscribe from communities
	s.communitySubsMu.Lock()
	for communityID := range client.subscribedCommunities {
		if subs, ok := s.communitySubs[communityID]; ok {
			delete(subs, client.userID)
			if len(subs) == 0 {
				delete(s.communitySubs, communityID)
			}
		}
	}
	s.communitySubsMu.Unlock()

	// Delete Redis presence key
	ctx := context.Background()
	presenceKey := fmt.Sprintf("presence:user:%s", client.userID)
	if err := s.redisClient.Del(ctx, presenceKey).Err(); err != nil {
		log.Printf("Failed to delete Redis presence key for user %s: %v", client.userID, err)
	}

	// Publish presence update
	s.publishPresenceChange(client.userID, "offline")
}

func (s *Server) publishPresenceChange(userID string, status string) {
	payload := RedisPresenceMsg{
		UserID: userID,
		Status: status,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	s.redisClient.Publish(context.Background(), "presence:updates", data)
}

func (s *Server) runPresenceHeartbeats(ctx context.Context, client *Client) {
	ticker := time.NewTicker(30 * time.Second)
	dbUpdateTicker := time.NewTicker(60 * time.Second)
	defer func() {
		ticker.Stop()
		dbUpdateTicker.Stop()
	}()

	// Update immediately on connection
	_, _ = s.db.Exec(`UPDATE users SET last_seen_at = NOW() WHERE id = $1`, client.userID)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			presenceKey := fmt.Sprintf("presence:user:%s", client.userID)
			s.redisClient.Expire(context.Background(), presenceKey, 70*time.Second)
		case <-dbUpdateTicker.C:
			_, err := s.db.Exec(`UPDATE users SET last_seen_at = NOW() WHERE id = $1`, client.userID)
			if err != nil {
				log.Printf("Failed to update last_seen_at for user %s: %v", client.userID, err)
			}
		}
	}
}

func (s *Server) readPump(client *Client) {
	defer func() {
		s.connClose(client)
	}()

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg ClientMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshalling client message: %v", err)
			continue
		}

		switch msg.Action {
		case "subscribe_event":
			s.handleSubscribeEvent(client, msg.EventID)
		case "unsubscribe_event":
			s.handleUnsubscribeEvent(client, msg.EventID)
		case "subscribe_community":
			s.handleSubscribeCommunity(client, msg.CommunityID)
		case "unsubscribe_community":
			s.handleUnsubscribeCommunity(client, msg.CommunityID)
		}
	}
}

func (s *Server) connClose(client *Client) {
	client.conn.Close()
	s.deregisterClient(client)
}

func (s *Server) handleSubscribeEvent(client *Client, eventID string) {
	if eventID == "" {
		return
	}
	s.eventSubsMu.Lock()
	if _, ok := s.eventSubs[eventID]; !ok {
		s.eventSubs[eventID] = make(map[string]*Client)
	}
	s.eventSubs[eventID][client.userID] = client
	s.eventSubsMu.Unlock()

	client.mu.Lock()
	client.subscribedEvents[eventID] = true
	client.mu.Unlock()
}

func (s *Server) handleUnsubscribeEvent(client *Client, eventID string) {
	if eventID == "" {
		return
	}
	s.eventSubsMu.Lock()
	if subs, ok := s.eventSubs[eventID]; ok {
		delete(subs, client.userID)
		if len(subs) == 0 {
			delete(s.eventSubs, eventID)
		}
	}
	s.eventSubsMu.Unlock()

	client.mu.Lock()
	delete(client.subscribedEvents, eventID)
	client.mu.Unlock()
}

func (s *Server) handleSubscribeCommunity(client *Client, communityID string) {
	if communityID == "" {
		return
	}
	s.communitySubsMu.Lock()
	if _, ok := s.communitySubs[communityID]; !ok {
		s.communitySubs[communityID] = make(map[string]*Client)
	}
	s.communitySubs[communityID][client.userID] = client
	s.communitySubsMu.Unlock()

	client.mu.Lock()
	client.subscribedCommunities[communityID] = true
	client.mu.Unlock()

	// Send list of current online members of the community to the client
	go s.sendCommunityOnlinePresenceList(client, communityID)
}

func (s *Server) handleUnsubscribeCommunity(client *Client, communityID string) {
	if communityID == "" {
		return
	}
	s.communitySubsMu.Lock()
	if subs, ok := s.communitySubs[communityID]; ok {
		delete(subs, client.userID)
		if len(subs) == 0 {
			delete(s.communitySubs, communityID)
		}
	}
	s.communitySubsMu.Unlock()

	client.mu.Lock()
	delete(client.subscribedCommunities, communityID)
	client.mu.Unlock()
}

func (s *Server) sendCommunityOnlinePresenceList(client *Client, communityID string) {
	rows, err := s.db.Query(`SELECT user_id FROM community_members WHERE community_id = $1`, communityID)
	if err != nil {
		log.Printf("Failed to query community members: %v", err)
		return
	}
	defer rows.Close()

	var memberIDs []string
	for rows.Next() {
		var memberID string
		if err := rows.Scan(&memberID); err == nil {
			memberIDs = append(memberIDs, memberID)
		}
	}

	if len(memberIDs) == 0 {
		client.sendJSON(ServerMessage{
			Type:        "presence_list",
			CommunityID: communityID,
			OnlineUsers: []string{},
		})
		return
	}

	ctx := context.Background()
	var onlineUsers []string

	keys := make([]string, len(memberIDs))
	for i, id := range memberIDs {
		keys[i] = fmt.Sprintf("presence:user:%s", id)
	}

	vals, err := s.redisClient.MGet(ctx, keys...).Result()
	if err != nil {
		log.Printf("Redis MGet failed: %v", err)
		// Fallback to checking local instance's active client map
		s.clientsMu.RLock()
		for _, id := range memberIDs {
			if _, exists := s.clients[id]; exists {
				onlineUsers = append(onlineUsers, id)
			}
		}
		s.clientsMu.RUnlock()
	} else {
		for i, val := range vals {
			if val != nil {
				onlineUsers = append(onlineUsers, memberIDs[i])
			}
		}
	}

	client.sendJSON(ServerMessage{
		Type:        "presence_list",
		CommunityID: communityID,
		OnlineUsers: onlineUsers,
	})
}

func (s *Server) subscribeRSVPUpdates() {
	pubsub := s.redisClient.Subscribe(context.Background(), "event:rsvp_updates")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var payload RedisRSVPMsg
		if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
			continue
		}

		s.eventSubsMu.RLock()
		subs, ok := s.eventSubs[payload.EventID]
		if !ok || len(subs) == 0 {
			s.eventSubsMu.RUnlock()
			continue
		}

		for _, client := range subs {
			go func(c *Client) {
				count := payload.RSVPCount
				c.sendJSON(ServerMessage{
					Type:      "rsvp_update",
					EventID:   payload.EventID,
					RSVPCount: &count,
				})
			}(client)
		}
		s.eventSubsMu.RUnlock()
	}
}

func (s *Server) subscribePresenceUpdates() {
	pubsub := s.redisClient.Subscribe(context.Background(), "presence:updates")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var payload RedisPresenceMsg
		if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
			continue
		}

		rows, err := s.db.Query(`SELECT community_id FROM community_members WHERE user_id = $1`, payload.UserID)
		if err != nil {
			continue
		}

		var communityIDs []string
		for rows.Next() {
			var cid string
			if err := rows.Scan(&cid); err == nil {
				communityIDs = append(communityIDs, cid)
			}
		}
		rows.Close()

		for _, cid := range communityIDs {
			s.communitySubsMu.RLock()
			subs, ok := s.communitySubs[cid]
			if !ok || len(subs) == 0 {
				s.communitySubsMu.RUnlock()
				continue
			}

			for _, client := range subs {
				go func(c *Client) {
					c.sendJSON(ServerMessage{
						Type:        "presence_change",
						CommunityID: cid,
						UserID:      payload.UserID,
						Status:      payload.Status,
					})
				}(client)
			}
			s.communitySubsMu.RUnlock()
		}
	}
}
