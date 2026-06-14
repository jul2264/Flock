package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/jul2264/Flock/backend/internal/db"
)

func main() {
	// Load .env file
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Set Clerk secret key
	clerk.SetKey(os.Getenv("CLERK_SECRET_KEY"))

	// Connect to database
	database := db.Connect()
	defer database.Close()

	// Run migrations
	db.RunMigrations(database)

	// Create router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Flock API is running!")
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
