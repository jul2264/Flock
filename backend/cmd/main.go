package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/jul2264/Flock/backend/internal/db"
	"github.com/jul2264/Flock/backend/internal/handlers"
	"github.com/jul2264/Flock/backend/internal/middleware"
	"github.com/jul2264/Flock/backend/internal/services"
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

	// ── Services ─────────────────────────────────────────────────────────────
	searchService := services.NewSearchService()
	userService := services.NewUserService(database)
	eventService := services.NewEventService(database, searchService)
	communityService := services.NewCommunityService(database, searchService)
	rsvpService := services.NewRSVPService(database)
	interestService := services.NewInterestService(database)

	// ── Handlers ─────────────────────────────────────────────────────────────
	userHandler := handlers.NewUserHandler(userService)
	eventHandler := handlers.NewEventHandler(eventService)
	communityHandler := handlers.NewCommunityHandler(communityService)
	rsvpHandler := handlers.NewRSVPHandler(rsvpService)
	interestHandler := handlers.NewInterestHandler(interestService)
	searchHandler := handlers.NewSearchHandler(searchService)

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

	// ── Protected routes (Clerk JWT required) ────────────────────────────────
	r.Group(func(r chi.Router) {
		r.Use(middleware.ClerkMiddleware(os.Getenv("CLERK_SECRET_KEY")))
		r.Use(middleware.RequireAuth)

		// Users
		r.Route("/users", func(r chi.Router) {
			r.Post("/sync", userHandler.SyncUser) // upsert user on login
			r.Get("/me", userHandler.GetMe)        // get own profile
			r.Patch("/me", userHandler.UpdateMe)   // update own profile
		})

		// RSVPs
		r.Get("/users/me/rsvps", rsvpHandler.ListUserRSVPs)

		// Events
		r.Route("/events", func(r chi.Router) {
			r.Post("/", eventHandler.CreateEvent)
			r.Get("/", eventHandler.ListEvents)
			r.Get("/{id}", eventHandler.GetEvent)

			// Nested RSVP endpoints
			r.Post("/{id}/rsvp", rsvpHandler.CreateRSVP)
			r.Delete("/{id}/rsvp", rsvpHandler.CancelRSVP)
			r.Get("/{id}/rsvps", rsvpHandler.ListEventRSVPs)

			// Nested Interest endpoints
			r.Get("/{id}/interests", interestHandler.GetEventInterests)
			r.Post("/{id}/interests", interestHandler.SetEventInterests)
		})

		// Communities
		r.Route("/communities", func(r chi.Router) {
			r.Post("/", communityHandler.CreateCommunity)
			r.Get("/", communityHandler.ListCommunities)
			r.Get("/{id}", communityHandler.GetCommunity)
		})

		// Interests
		r.Get("/interests", interestHandler.ListInterests)
		r.Post("/interests", interestHandler.CreateInterest)
		r.Get("/users/me/interests", interestHandler.GetUserInterests)
		r.Post("/users/me/interests", interestHandler.SetUserInterests)

		// Search
		r.Get("/search", searchHandler.Search)
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
