package handlers

import (
	"encoding/json"
	"net/http"

	"novel-drafting-api/internal/models"
	"novel-drafting-api/internal/repository"
)

// NotesHandler holds dependencies for book-note HTTP handlers.
type NotesHandler struct {
	repo repository.NoteRepository
}

// NewNotesHandler creates a new NotesHandler.
func NewNotesHandler(repo repository.NoteRepository) *NotesHandler {
	return &NotesHandler{repo: repo}
}

// ListNotes returns all notes belonging to a specific book, ordered by updated_at DESC.
func (h *NotesHandler) ListNotes(w http.ResponseWriter, r *http.Request) {
	bookID := r.PathValue("book_id")
	if bookID == "" {
		writeError(w, http.StatusBadRequest, "invalid book_id")
		return
	}

	notes, err := h.repo.List(r.Context(), bookID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query notes")
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

	n, err := h.repo.Create(r.Context(), bookID, req.Title, req.Type, req.Content)
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

	var titlePtr *string
	if req.Title != "" {
		titlePtr = &req.Title
	}
	var contentPtr *json.RawMessage
	if req.Content != nil {
		contentPtr = &req.Content
	}

	n, err := h.repo.Update(r.Context(), noteID, titlePtr, contentPtr)
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

	rowsAffected, err := h.repo.Delete(r.Context(), noteID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete note")
		return
	}

	if rowsAffected == 0 {
		writeError(w, http.StatusNotFound, "note not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
