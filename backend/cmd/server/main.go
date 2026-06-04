package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"novel-drafting-api/internal/database"
	"novel-drafting-api/internal/handlers"
	"novel-drafting-api/internal/middleware"
)

func main() {
	// Attempt to load .env file; it's okay if it doesn't exist in production
	_ = godotenv.Load()

	ctx := context.Background()

	// Connect to PostgreSQL.
	pool, err := database.Connect(ctx)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	// Seed a default user for Phase 1 (no auth).
	defaultUserID, err := database.SeedDefaultUser(ctx, pool)
	if err != nil {
		log.Fatalf("seed default user failed: %v", err)
	}

	// Create handlers.
	bookHandler := handlers.NewBookHandler(pool, defaultUserID.String())
	chapterHandler := handlers.NewChapterHandler(pool)
	dictionaryHandler := handlers.NewDictionaryHandler(pool, defaultUserID.String())
	versionHandler := handlers.NewVersionHandler(pool)
	notesHandler := handlers.NewNotesHandler(pool)

	// Register routes using Go 1.22+ method patterns.
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"novel-drafting-api"}`))
	})

	// Dictionary routes
	mux.HandleFunc("GET /api/dictionary", dictionaryHandler.GetDictionary)
	mux.HandleFunc("POST /api/dictionary", dictionaryHandler.AddWord)

	// Book routes
	mux.HandleFunc("GET /api/books", bookHandler.ListBooks)
	mux.HandleFunc("POST /api/books", bookHandler.CreateBook)
	mux.HandleFunc("GET /api/books/{id}/preview", bookHandler.GetBookPreview)

	// Chapter routes (book-scoped)
	mux.HandleFunc("GET /api/books/{bookId}/chapters", chapterHandler.ListChapters)
	mux.HandleFunc("POST /api/books/{bookId}/chapters", chapterHandler.CreateChapter)

	// Chapter routes (chapter-scoped)
	mux.HandleFunc("GET /api/chapters/{id}", chapterHandler.GetChapter)
	mux.HandleFunc("PUT /api/chapters/{id}", chapterHandler.UpdateChapter)
	mux.HandleFunc("DELETE /api/chapters/{id}", chapterHandler.DeleteChapter)
	mux.HandleFunc("PUT /api/chapters/reorder", chapterHandler.ReorderChapters)

	// Version / revision history routes
	mux.HandleFunc("POST /api/chapters/{id}/versions", versionHandler.CreateVersion)
	mux.HandleFunc("GET /api/chapters/{id}/versions", versionHandler.ListVersions)
	mux.HandleFunc("GET /api/versions/{version_id}", versionHandler.GetVersion)
	mux.HandleFunc("DELETE /api/versions/{version_id}", versionHandler.DeleteVersion)

	// Book notes (Story Bible – Character Sheets & Worldbuilding Wiki)
	mux.HandleFunc("GET /api/books/{book_id}/notes", notesHandler.ListNotes)
	mux.HandleFunc("POST /api/books/{book_id}/notes", notesHandler.CreateNote)
	mux.HandleFunc("PUT /api/notes/{note_id}", notesHandler.UpdateNote)
	mux.HandleFunc("DELETE /api/notes/{note_id}", notesHandler.DeleteNote)

	// Apply middleware chain: CORS → Logger → JSONContent → mux
	handler := middleware.CORS(middleware.Logger(middleware.JSONContent(mux)))

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown.
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		log.Println("🛑 Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("server shutdown error: %v", err)
		}
	}()

	log.Printf("🚀 Novel Drafting API listening on :%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
	log.Println("✅ Server stopped gracefully")
}
