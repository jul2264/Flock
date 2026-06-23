package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	httprateredis "github.com/go-chi/httprate-redis"
	"github.com/joho/godotenv"

	"github.com/jul2264/Flock/backend/internal/db"
	"github.com/jul2264/Flock/backend/internal/handlers"
	"github.com/jul2264/Flock/backend/internal/middleware"
	"github.com/jul2264/Flock/backend/internal/services"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load .env file (fall back to system env vars if not found)
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialise Clerk with the secret key so JWT verification works
	clerk.SetKey(os.Getenv("CLERK_SECRET_KEY"))

	// Connect to Postgres and run pending migrations
	database := db.Connect()
	defer database.Close()
	db.RunMigrations(database)

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
		log.Printf("Warning: failed to connect to Redis: %v. Realtime updates may not function.", err)
	}

	// ── Services ─────────────────────────────────────────────────────────────
	searchService := services.NewSearchService()
	userService := services.NewUserService(database)
	eventService := services.NewEventService(database, searchService)
	communityService := services.NewCommunityService(database, searchService)
	rsvpService := services.NewRSVPService(database, redisClient)
	interestService := services.NewInterestService(database)
	storageService, err := services.NewStorageService()
	if err != nil {
		log.Printf("Warning: failed to initialize storage service: %v. Media uploads will be unavailable.", err)
	}

	// ── Handlers ─────────────────────────────────────────────────────────────
	userHandler := handlers.NewUserHandler(userService)
	eventHandler := handlers.NewEventHandler(eventService, userService)
	communityHandler := handlers.NewCommunityHandler(communityService, userService)
	rsvpHandler := handlers.NewRSVPHandler(rsvpService)
	interestHandler := handlers.NewInterestHandler(interestService)
	searchHandler := handlers.NewSearchHandler(searchService)
	uploadHandler := handlers.NewUploadHandler(storageService)

	// ── Rate Limiting Setup ──────────────────────────────────────────────────
	redisURL = os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}
	u, err := url.Parse(redisURL)
	var host string = "localhost"
	var redisPort int = 6379
	if err == nil {
		host = u.Hostname()
		if host == "" {
			host = "localhost"
		}
		if portStr := u.Port(); portStr != "" {
			if p, err := strconv.Atoi(portStr); err == nil {
				redisPort = p
			}
		}
	}

	rateLimitRedisConfig := &httprateredis.Config{
		Host: host,
		Port: uint16(redisPort),
	}

	// Read routes: 100 requests per minute
	readLimiter := httprate.Limit(
		100,
		time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP),
		httprateredis.WithRedisLimitCounter(rateLimitRedisConfig),
	)

	// Write routes: 10 requests per minute
	writeLimiter := httprate.Limit(
		10,
		time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP),
		httprateredis.WithRedisLimitCounter(rateLimitRedisConfig),
	)

	// ── Router ───────────────────────────────────────────────────────────────
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	// ── Public routes ────────────────────────────────────────────────────────
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok","service":"Flock API"}`)
	})

	// Base Clerk auth middleware — applied to all protected groups below
	clerkAuth := middleware.ClerkMiddleware(os.Getenv("CLERK_SECRET_KEY"))

	// ── Tier 1: User-level routes (any authenticated user) ───────────────────
	r.Group(func(r chi.Router) {
		r.Use(clerkAuth)
		r.Use(middleware.RequireAuth)

		// Tier 1 READS (100 req/min limit)
		r.Group(func(r chi.Router) {
			r.Use(readLimiter)

			r.Get("/users/me", userHandler.GetMe)
			r.Get("/users/me/rsvps", rsvpHandler.ListUserRSVPs)
			r.Get("/users/me/interests", interestHandler.GetUserInterests)

			r.Get("/events", eventHandler.ListEvents)
			r.Get("/events/{id}", eventHandler.GetEvent)
			r.Get("/events/{id}/rsvps", rsvpHandler.ListEventRSVPs)
			r.Get("/events/{id}/interests", interestHandler.GetEventInterests)

			r.Get("/communities", communityHandler.ListCommunities)
			r.Get("/communities/{id}", communityHandler.GetCommunity)
			r.Get("/communities/{id}/members", communityHandler.ListCommunityMembers)

			r.Get("/interests", interestHandler.ListInterests)
			r.Get("/search", searchHandler.Search)
		})

		// Tier 1 WRITES (10 req/min limit)
		r.Group(func(r chi.Router) {
			r.Use(writeLimiter)

			r.Post("/users/sync", userHandler.SyncUser)
			r.Patch("/users/me", userHandler.UpdateMe)
			r.Post("/users/me/interests", interestHandler.SetUserInterests)

			r.Post("/events/{id}/rsvp", rsvpHandler.CreateRSVP)
			r.Delete("/events/{id}/rsvp", rsvpHandler.CancelRSVP)

			r.Post("/communities/{id}/join", communityHandler.JoinCommunity)
			r.Delete("/communities/{id}/leave", communityHandler.LeaveCommunity)

			r.Post("/upload/avatar", uploadHandler.GenerateAvatarUploadURL)
		})
	})

	// ── Tier 2: Organizer-level routes (organizer or admin) ──────────────────
	r.Group(func(r chi.Router) {
		r.Use(clerkAuth)
		r.Use(middleware.RequireAuth)
		r.Use(middleware.RequireRole(database, "organizer", "admin"))
		r.Use(writeLimiter) // Organizer mutations are writes

		// Create and manage own events
		r.Post("/events", eventHandler.CreateEvent)
		r.Patch("/events/{id}", eventHandler.UpdateEvent)
		r.Delete("/events/{id}", eventHandler.DeleteEvent)
		r.Post("/events/{id}/interests", interestHandler.SetEventInterests)

		// Create and manage own communities
		r.Post("/communities", communityHandler.CreateCommunity)
		r.Patch("/communities/{id}", communityHandler.UpdateCommunity)
		r.Delete("/communities/{id}", communityHandler.DeleteCommunity)

		// Media Uploads
		r.Post("/upload/event-banner", uploadHandler.GenerateEventBannerUploadURL)
		r.Post("/upload/community-image", uploadHandler.GenerateCommunityImageUploadURL)
	})

	// ── Tier 3: Admin-only routes ─────────────────────────────────────────────
	r.Group(func(r chi.Router) {
		r.Use(clerkAuth)
		r.Use(middleware.RequireAuth)
		r.Use(middleware.RequireRole(database, "admin"))
		r.Use(writeLimiter) // Admin mutations are writes

		// Seed and manage global interest taxonomy
		r.Post("/interests", interestHandler.CreateInterest)
	})

	// ── Start server ─────────────────────────────────────────────────────────
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Flock API listening on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
