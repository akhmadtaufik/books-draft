package handlers

import (
	"encoding/json"
	"net/http"

	"novel-drafting-api/internal/models"
	"novel-drafting-api/internal/repository"
)

// ChapterHandler holds dependencies for chapter-related HTTP handlers.
type ChapterHandler struct {
	repo repository.ChapterRepository
}

// NewChapterHandler creates a new ChapterHandler.
func NewChapterHandler(repo repository.ChapterRepository) *ChapterHandler {
	return &ChapterHandler{repo: repo}
}

// ListChapters returns all chapters for a book (without content).
func (h *ChapterHandler) ListChapters(w http.ResponseWriter, r *http.Request) {
	bookID := r.PathValue("bookId")
	if bookID == "" {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	chapters, err := h.repo.List(r.Context(), bookID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query chapters")
		return
	}

	json.NewEncoder(w).Encode(chapters)
}

// CreateChapter creates a new chapter at the end of the book.
func (h *ChapterHandler) CreateChapter(w http.ResponseWriter, r *http.Request) {
	bookID := r.PathValue("bookId")
	if bookID == "" {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	var req struct {
		Title   string          `json:"title"`
		Content json.RawMessage `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if req.Content == nil {
		req.Content = json.RawMessage(`{}`)
	}

	ch, err := h.repo.Create(r.Context(), bookID, req.Title, req.Content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create chapter")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ch)
}

// GetChapter returns a single chapter with content.
func (h *ChapterHandler) GetChapter(w http.ResponseWriter, r *http.Request) {
	chapterID := r.PathValue("id")
	if chapterID == "" {
		writeError(w, http.StatusBadRequest, "invalid chapter id")
		return
	}

	ch, err := h.repo.GetByID(r.Context(), chapterID)
	if err != nil {
		writeError(w, http.StatusNotFound, "chapter not found")
		return
	}

	json.NewEncoder(w).Encode(ch)
}

// UpdateChapter updates a chapter's title and/or content.
func (h *ChapterHandler) UpdateChapter(w http.ResponseWriter, r *http.Request) {
	chapterID := r.PathValue("id")
	if chapterID == "" {
		writeError(w, http.StatusBadRequest, "invalid chapter id")
		return
	}

	var req struct {
		Title   *string          `json:"title"`
		Content *json.RawMessage `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == nil && req.Content == nil {
		writeError(w, http.StatusBadRequest, "nothing to update")
		return
	}

	// Build dynamic update.
	ch, err := h.repo.Update(r.Context(), chapterID, req.Title, req.Content)
	if err != nil {
		writeError(w, http.StatusNotFound, "chapter not found")
		return
	}

	json.NewEncoder(w).Encode(ch)
}

// DeleteChapter deletes a chapter and reindexes the remaining chapters.
func (h *ChapterHandler) DeleteChapter(w http.ResponseWriter, r *http.Request) {
	chapterID := r.PathValue("id")
	if chapterID == "" {
		writeError(w, http.StatusBadRequest, "invalid chapter id")
		return
	}

	err := h.repo.Delete(r.Context(), chapterID)
	if err != nil {
		// Could differentiate errors but simplistic generic error here
		writeError(w, http.StatusInternalServerError, "failed to delete chapter")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ReorderChapters updates the position_index for a list of chapter IDs.
func (h *ChapterHandler) ReorderChapters(w http.ResponseWriter, r *http.Request) {
	var req models.ReorderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.BookID == "" || len(req.OrderedIDs) == 0 {
		writeError(w, http.StatusBadRequest, "bookId and orderedIds are required")
		return
	}

	bookID := req.BookID
	if bookID == "" {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	err := h.repo.Reorder(r.Context(), bookID, req.OrderedIDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reorder chapters")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
