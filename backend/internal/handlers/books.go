package handlers

import (
	"encoding/json"
	"net/http"

	"novel-drafting-api/internal/models"
	"novel-drafting-api/internal/repository"
)

// BookHandler holds dependencies for book-related HTTP handlers.
type BookHandler struct {
	repo          repository.BookRepository
	defaultUserID string
}

// NewBookHandler creates a new BookHandler.
func NewBookHandler(repo repository.BookRepository, defaultUserID string) *BookHandler {
	return &BookHandler{repo: repo, defaultUserID: defaultUserID}
}

// ListBooks returns all books belonging to the default user.
func (h *BookHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.repo.GetAll(r.Context(), h.defaultUserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query books")
		return
	}

	json.NewEncoder(w).Encode(books)
}

// CreateBook creates a new book for the default user.
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title     string `json:"title"`
		Author    string `json:"author"`
		Synopsis  string `json:"synopsis"`
		Genre     string `json:"genre"`
		Language  string `json:"language"`
		ISBN      string `json:"isbn"`
		Publisher string `json:"publisher"`
		Status    string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if req.Language == "" {
		req.Language = "Indonesian"
	}
	if req.Status == "" {
		req.Status = "Draft"
	}

	bookToCreate := models.Book{
		Title:     req.Title,
		Author:    req.Author,
		Synopsis:  req.Synopsis,
		Genre:     req.Genre,
		Language:  req.Language,
		ISBN:      req.ISBN,
		Publisher: req.Publisher,
		Status:    req.Status,
	}

	b, err := h.repo.Create(r.Context(), h.defaultUserID, &bookToCreate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create book")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(b)
}

// UpdateBook updates an existing book.
func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	bookID := r.PathValue("id")
	if bookID == "" {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	var req struct {
		Title     string `json:"title"`
		Author    string `json:"author"`
		Synopsis  string `json:"synopsis"`
		Genre     string `json:"genre"`
		Language  string `json:"language"`
		ISBN      string `json:"isbn"`
		Publisher string `json:"publisher"`
		Status    string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if req.Language == "" {
		req.Language = "Indonesian"
	}
	if req.Status == "" {
		req.Status = "Draft"
	}

	bookToUpdate := models.Book{
		Title:     req.Title,
		Author:    req.Author,
		Synopsis:  req.Synopsis,
		Genre:     req.Genre,
		Language:  req.Language,
		ISBN:      req.ISBN,
		Publisher: req.Publisher,
		Status:    req.Status,
	}

	b, err := h.repo.Update(r.Context(), h.defaultUserID, bookID, &bookToUpdate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update book")
		return
	}

	json.NewEncoder(w).Encode(b)
}

// DeleteBook deletes an existing book.
func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	bookID := r.PathValue("id")
	if bookID == "" {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	rowsAffected, err := h.repo.Delete(r.Context(), h.defaultUserID, bookID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete book")
		return
	}
	if rowsAffected == 0 {
		writeError(w, http.StatusNotFound, "book not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetBookPreview returns a book title and all its chapters (title + content) for preview/export.
func (h *BookHandler) GetBookPreview(w http.ResponseWriter, r *http.Request) {
	bookID := r.PathValue("id")
	if bookID == "" {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	preview, err := h.repo.GetPreview(r.Context(), h.defaultUserID, bookID)
	if err != nil {
		// Could differentiate between 404 and 500 based on error, but keeping it simple for now
		writeError(w, http.StatusInternalServerError, "failed to get book preview")
		return
	}

	json.NewEncoder(w).Encode(preview)
}
