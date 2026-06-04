package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"novel-drafting-api/internal/models"
)

// NotesHandler holds dependencies for book-note HTTP handlers.
type NotesHandler struct {
	pool *pgxpool.Pool
}

// NewNotesHandler creates a new NotesHandler.
func NewNotesHandler(pool *pgxpool.Pool) *NotesHandler {
	return &NotesHandler{pool: pool}
}

// ListNotes returns all notes belonging to a specific book, ordered by updated_at DESC.
func (h *NotesHandler) ListNotes(w http.ResponseWriter, r *http.Request) {
	bookID := r.PathValue("book_id")
	if bookID == "" {
		writeError(w, http.StatusBadRequest, "invalid book_id")
		return
	}

	rows, err := h.pool.Query(r.Context(),
		`SELECT id, book_id, title, type, content, updated_at
		 FROM book_notes
		 WHERE book_id = $1
		 ORDER BY updated_at DESC`,
		bookID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query notes")
		return
	}
	defer rows.Close()

	notes := make([]models.BookNote, 0)
	for rows.Next() {
		var n models.BookNote
		if err := rows.Scan(&n.ID, &n.BookID, &n.Title, &n.Type, &n.Content, &n.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan note")
			return
		}
		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "error iterating notes")
		return
	}

	json.NewEncoder(w).Encode(notes)
}

// CreateNote creates a new note for a specific book.
func (h *NotesHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	bookID := r.PathValue("book_id")
	if bookID == "" {
		writeError(w, http.StatusBadRequest, "invalid book_id")
		return
	}

	var req models.CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if req.Type != "character" && req.Type != "worldbuilding" {
		writeError(w, http.StatusBadRequest, "type must be 'character' or 'worldbuilding'")
		return
	}

	// Default content to empty JSON object if not provided.
	if req.Content == nil {
		req.Content = json.RawMessage(`{}`)
	}

	var n models.BookNote
	err := h.pool.QueryRow(r.Context(),
		`INSERT INTO book_notes (book_id, title, type, content)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, book_id, title, type, content, updated_at`,
		bookID, req.Title, req.Type, req.Content,
	).Scan(&n.ID, &n.BookID, &n.Title, &n.Type, &n.Content, &n.UpdatedAt)

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create note")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(n)
}

// UpdateNote updates the title and content of an existing note.
func (h *NotesHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	noteID := r.PathValue("note_id")
	if noteID == "" {
		writeError(w, http.StatusBadRequest, "invalid note_id")
		return
	}

	var req models.UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var n models.BookNote
	err := h.pool.QueryRow(r.Context(),
		`UPDATE book_notes
		 SET title = COALESCE(NULLIF($1, ''), title),
		     content = COALESCE($2, content),
		     updated_at = NOW()
		 WHERE id = $3
		 RETURNING id, book_id, title, type, content, updated_at`,
		req.Title, req.Content, noteID,
	).Scan(&n.ID, &n.BookID, &n.Title, &n.Type, &n.Content, &n.UpdatedAt)

	if err != nil {
		writeError(w, http.StatusNotFound, "note not found")
		return
	}

	json.NewEncoder(w).Encode(n)
}

// DeleteNote deletes a specific note.
func (h *NotesHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	noteID := r.PathValue("note_id")
	if noteID == "" {
		writeError(w, http.StatusBadRequest, "invalid note_id")
		return
	}

	tag, err := h.pool.Exec(r.Context(),
		`DELETE FROM book_notes WHERE id = $1`,
		noteID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete note")
		return
	}

	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "note not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
