package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"novel-drafting-api/internal/models"
)

// BookHandler holds dependencies for book-related HTTP handlers.
type BookHandler struct {
	pool          *pgxpool.Pool
	defaultUserID string
}

// NewBookHandler creates a new BookHandler.
func NewBookHandler(pool *pgxpool.Pool, defaultUserID string) *BookHandler {
	return &BookHandler{pool: pool, defaultUserID: defaultUserID}
}

// ListBooks returns all books belonging to the default user.
func (h *BookHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
	rows, err := h.pool.Query(r.Context(),
		`SELECT id, title, author, synopsis, genre, language, isbn, publisher, status, cover_image_url, metadata, created_at, updated_at
		 FROM books WHERE user_id = $1
		 ORDER BY updated_at DESC, created_at DESC`,
		h.defaultUserID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query books")
		return
	}
	defer rows.Close()

	books := make([]models.Book, 0)
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Synopsis, &b.Genre, &b.Language, &b.ISBN, &b.Publisher, &b.Status, &b.CoverImageURL, &b.Metadata, &b.CreatedAt, &b.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan book")
			return
		}
		books = append(books, b)
	}

	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "error iterating books")
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

	var b models.Book
	err := h.pool.QueryRow(r.Context(),
		`INSERT INTO books (user_id, title, author, synopsis, genre, language, isbn, publisher, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, user_id, title, author, synopsis, genre, language, isbn, publisher, status, cover_image_url, metadata, created_at, updated_at`,
		h.defaultUserID, req.Title, req.Author, req.Synopsis, req.Genre, req.Language, req.ISBN, req.Publisher, req.Status,
	).Scan(&b.ID, &b.UserID, &b.Title, &b.Author, &b.Synopsis, &b.Genre, &b.Language, &b.ISBN, &b.Publisher, &b.Status, &b.CoverImageURL, &b.Metadata, &b.CreatedAt, &b.UpdatedAt)

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

	var b models.Book
	err := h.pool.QueryRow(r.Context(),
		`UPDATE books 
		 SET title = $1, author = $2, synopsis = $3, genre = $4, language = $5, isbn = $6, publisher = $7, status = $8, updated_at = NOW()
		 WHERE id = $9 AND user_id = $10
		 RETURNING id, user_id, title, author, synopsis, genre, language, isbn, publisher, status, cover_image_url, metadata, created_at, updated_at`,
		req.Title, req.Author, req.Synopsis, req.Genre, req.Language, req.ISBN, req.Publisher, req.Status, bookID, h.defaultUserID,
	).Scan(&b.ID, &b.UserID, &b.Title, &b.Author, &b.Synopsis, &b.Genre, &b.Language, &b.ISBN, &b.Publisher, &b.Status, &b.CoverImageURL, &b.Metadata, &b.CreatedAt, &b.UpdatedAt)

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

	commandTag, err := h.pool.Exec(r.Context(),
		"DELETE FROM books WHERE id = $1 AND user_id = $2",
		bookID, h.defaultUserID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete book")
		return
	}
	if commandTag.RowsAffected() == 0 {
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

	// Get book title.
	var title string
	err := h.pool.QueryRow(r.Context(),
		"SELECT title FROM books WHERE id = $1 AND user_id = $2",
		bookID, h.defaultUserID,
	).Scan(&title)
	if err != nil {
		writeError(w, http.StatusNotFound, "book not found")
		return
	}

	// Get all chapters ordered by position.
	rows, err := h.pool.Query(r.Context(),
		`SELECT title, content FROM chapters
		 WHERE book_id = $1
		 ORDER BY position_index ASC`,
		bookID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query chapters")
		return
	}
	defer rows.Close()

	chapters := make([]models.ChapterPreview, 0)
	for rows.Next() {
		var ch models.ChapterPreview
		if err := rows.Scan(&ch.Title, &ch.Content); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan chapter")
			return
		}
		chapters = append(chapters, ch)
	}

	json.NewEncoder(w).Encode(models.BookPreview{
		Title:    title,
		Chapters: chapters,
	})
}
