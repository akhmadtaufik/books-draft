package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"

	"novel-drafting-api/internal/models"
)

type NoteRepository interface {
	List(ctx context.Context, bookID string) ([]models.BookNote, error)
	Create(ctx context.Context, bookID string, title string, noteType string, content json.RawMessage) (models.BookNote, error)
	Update(ctx context.Context, id string, title *string, content *json.RawMessage) (models.BookNote, error)
	Delete(ctx context.Context, id string) (int64, error)
}

type noteRepository struct {
	pool *pgxpool.Pool
}

func NewNoteRepository(pool *pgxpool.Pool) NoteRepository {
	return &noteRepository{pool: pool}
}

func (r *noteRepository) List(ctx context.Context, bookID string) ([]models.BookNote, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, book_id, title, type, content, updated_at
		 FROM book_notes
		 WHERE book_id = $1
		 ORDER BY updated_at DESC`,
		bookID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := make([]models.BookNote, 0)
	for rows.Next() {
		var n models.BookNote
		if err := rows.Scan(&n.ID, &n.BookID, &n.Title, &n.Type, &n.Content, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

func (r *noteRepository) Create(ctx context.Context, bookID string, title string, noteType string, content json.RawMessage) (models.BookNote, error) {
	var n models.BookNote
	err := r.pool.QueryRow(ctx,
		`INSERT INTO book_notes (book_id, title, type, content)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, book_id, title, type, content, updated_at`,
		bookID, title, noteType, content,
	).Scan(&n.ID, &n.BookID, &n.Title, &n.Type, &n.Content, &n.UpdatedAt)
	return n, err
}

func (r *noteRepository) Update(ctx context.Context, id string, title *string, content *json.RawMessage) (models.BookNote, error) {
	var n models.BookNote
	err := r.pool.QueryRow(ctx,
		`UPDATE book_notes
		 SET title = COALESCE(NULLIF($1, ''), title),
		     content = COALESCE($2, content),
		     updated_at = NOW()
		 WHERE id = $3
		 RETURNING id, book_id, title, type, content, updated_at`,
		title, content, id,
	).Scan(&n.ID, &n.BookID, &n.Title, &n.Type, &n.Content, &n.UpdatedAt)
	return n, err
}

func (r *noteRepository) Delete(ctx context.Context, id string) (int64, error) {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM book_notes WHERE id = $1`,
		id,
	)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
