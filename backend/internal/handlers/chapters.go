package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"novel-drafting-api/internal/models"
)

// ChapterHandler holds dependencies for chapter-related HTTP handlers.
type ChapterHandler struct {
	pool *pgxpool.Pool
}

// NewChapterHandler creates a new ChapterHandler.
func NewChapterHandler(pool *pgxpool.Pool) *ChapterHandler {
	return &ChapterHandler{pool: pool}
}

// ListChapters returns all chapters for a book (without content).
func (h *ChapterHandler) ListChapters(w http.ResponseWriter, r *http.Request) {
	bookID := r.PathValue("bookId")
	if bookID == "" {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	rows, err := h.pool.Query(r.Context(),
		`SELECT id, title, position_index
		 FROM chapters WHERE book_id = $1
		 ORDER BY position_index ASC`,
		bookID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query chapters")
		return
	}
	defer rows.Close()

	chapters := make([]models.ChapterListItem, 0)
	for rows.Next() {
		var ch models.ChapterListItem
		if err := rows.Scan(&ch.ID, &ch.Title, &ch.PositionIndex); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan chapter")
			return
		}
		chapters = append(chapters, ch)
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

	// Calculate next position_index.
	var maxPos *int
	err := h.pool.QueryRow(r.Context(),
		"SELECT MAX(position_index) FROM chapters WHERE book_id = $1",
		bookID,
	).Scan(&maxPos)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to calculate position")
		return
	}

	nextPos := 0
	if maxPos != nil {
		nextPos = *maxPos + 1
	}

	var ch models.Chapter
	err = h.pool.QueryRow(r.Context(),
		`INSERT INTO chapters (book_id, title, content, position_index)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, book_id, title, content, position_index, updated_at`,
		bookID, req.Title, req.Content, nextPos,
	).Scan(&ch.ID, &ch.BookID, &ch.Title, &ch.Content, &ch.PositionIndex, &ch.UpdatedAt)

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

	var ch models.Chapter
	err := h.pool.QueryRow(r.Context(),
		`SELECT id, book_id, title, content, position_index, updated_at
		 FROM chapters WHERE id = $1`,
		chapterID,
	).Scan(&ch.ID, &ch.BookID, &ch.Title, &ch.Content, &ch.PositionIndex, &ch.UpdatedAt)

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
	var ch models.Chapter
	err := h.pool.QueryRow(r.Context(),
		`UPDATE chapters SET
			title = COALESCE($1, title),
			content = COALESCE($2, content),
			updated_at = NOW()
		 WHERE id = $3
		 RETURNING id, book_id, title, content, position_index, updated_at`,
		req.Title, req.Content, chapterID,
	).Scan(&ch.ID, &ch.BookID, &ch.Title, &ch.Content, &ch.PositionIndex, &ch.UpdatedAt)

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

	ctx := r.Context()

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to begin transaction")
		return
	}
	defer tx.Rollback(ctx)

	// Get the book_id before deleting.
	var bookID string
	err = tx.QueryRow(ctx,
		"SELECT book_id FROM chapters WHERE id = $1", chapterID,
	).Scan(&bookID)
	if err != nil {
		writeError(w, http.StatusNotFound, "chapter not found")
		return
	}

	// Delete the chapter.
	_, err = tx.Exec(ctx, "DELETE FROM chapters WHERE id = $1", chapterID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete chapter")
		return
	}

	// Reindex remaining chapters.
	rows, err := tx.Query(ctx,
		`SELECT id FROM chapters
		 WHERE book_id = $1
		 ORDER BY position_index ASC`,
		bookID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reindex chapters")
		return
	}

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			writeError(w, http.StatusInternalServerError, "failed to scan chapter id")
			return
		}
		ids = append(ids, id)
	}
	rows.Close()

	for i, id := range ids {
		_, err = tx.Exec(ctx,
			"UPDATE chapters SET position_index = $1 WHERE id = $2",
			i, id,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to reindex chapter")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit transaction")
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

	ctx := r.Context()

	tx, err := h.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to begin transaction")
		return
	}
	defer tx.Rollback(ctx)

	for i, chapterID := range req.OrderedIDs {
		if chapterID == "" {
			writeError(w, http.StatusBadRequest, "invalid chapter id in orderedIds")
			return
		}

		tag, err := tx.Exec(ctx,
			`UPDATE chapters SET position_index = $1
			 WHERE id = $2 AND book_id = $3`,
			i, chapterID, bookID,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to reorder chapter")
			return
		}
		if tag.RowsAffected() == 0 {
			writeError(w, http.StatusNotFound, "chapter not found in this book")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit reorder")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
