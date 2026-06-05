package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"novel-drafting-api/internal/models"
)

type ChapterRepository interface {
	List(ctx context.Context, bookID string) ([]models.ChapterListItem, error)
	Create(ctx context.Context, bookID string, title string, content json.RawMessage) (models.Chapter, error)
	GetByID(ctx context.Context, id string) (models.Chapter, error)
	Update(ctx context.Context, id string, title *string, content *json.RawMessage) (models.Chapter, error)
	Delete(ctx context.Context, id string) error
	Reorder(ctx context.Context, bookID string, orderedIDs []string) error
}

type chapterRepository struct {
	pool *pgxpool.Pool
}

func NewChapterRepository(pool *pgxpool.Pool) ChapterRepository {
	return &chapterRepository{pool: pool}
}

func (r *chapterRepository) List(ctx context.Context, bookID string) ([]models.ChapterListItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, title, position_index
		 FROM chapters WHERE book_id = $1
		 ORDER BY position_index ASC`,
		bookID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chapters := make([]models.ChapterListItem, 0)
	for rows.Next() {
		var ch models.ChapterListItem
		if err := rows.Scan(&ch.ID, &ch.Title, &ch.PositionIndex); err != nil {
			return nil, err
		}
		chapters = append(chapters, ch)
	}

	return chapters, nil
}

func (r *chapterRepository) Create(ctx context.Context, bookID string, title string, content json.RawMessage) (models.Chapter, error) {
	var maxPos *int
	err := r.pool.QueryRow(ctx,
		"SELECT MAX(position_index) FROM chapters WHERE book_id = $1",
		bookID,
	).Scan(&maxPos)
	if err != nil {
		return models.Chapter{}, err
	}

	nextPos := 0
	if maxPos != nil {
		nextPos = *maxPos + 1
	}

	var ch models.Chapter
	err = r.pool.QueryRow(ctx,
		`INSERT INTO chapters (book_id, title, content, position_index)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, book_id, title, content, position_index, updated_at`,
		bookID, title, content, nextPos,
	).Scan(&ch.ID, &ch.BookID, &ch.Title, &ch.Content, &ch.PositionIndex, &ch.UpdatedAt)

	return ch, err
}

func (r *chapterRepository) GetByID(ctx context.Context, id string) (models.Chapter, error) {
	var ch models.Chapter
	err := r.pool.QueryRow(ctx,
		`SELECT id, book_id, title, content, position_index, updated_at
		 FROM chapters WHERE id = $1`,
		id,
	).Scan(&ch.ID, &ch.BookID, &ch.Title, &ch.Content, &ch.PositionIndex, &ch.UpdatedAt)
	return ch, err
}

func (r *chapterRepository) Update(ctx context.Context, id string, title *string, content *json.RawMessage) (models.Chapter, error) {
	var ch models.Chapter
	err := r.pool.QueryRow(ctx,
		`UPDATE chapters SET
			title = COALESCE($1, title),
			content = COALESCE($2, content),
			updated_at = NOW()
		 WHERE id = $3
		 RETURNING id, book_id, title, content, position_index, updated_at`,
		title, content, id,
	).Scan(&ch.ID, &ch.BookID, &ch.Title, &ch.Content, &ch.PositionIndex, &ch.UpdatedAt)
	return ch, err
}

func (r *chapterRepository) Delete(ctx context.Context, id string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var bookID string
	err = tx.QueryRow(ctx, "SELECT book_id FROM chapters WHERE id = $1", id).Scan(&bookID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "DELETE FROM chapters WHERE id = $1", id)
	if err != nil {
		return err
	}

	rows, err := tx.Query(ctx,
		`SELECT id FROM chapters
		 WHERE book_id = $1
		 ORDER BY position_index ASC`,
		bookID,
	)
	if err != nil {
		return err
	}

	ids := make([]string, 0)
	for rows.Next() {
		var chapterID string
		if err := rows.Scan(&chapterID); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, chapterID)
	}
	rows.Close()

	for i, cid := range ids {
		_, err = tx.Exec(ctx,
			"UPDATE chapters SET position_index = $1 WHERE id = $2",
			i, cid,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *chapterRepository) Reorder(ctx context.Context, bookID string, orderedIDs []string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, chapterID := range orderedIDs {
		tag, err := tx.Exec(ctx,
			`UPDATE chapters SET position_index = $1
			 WHERE id = $2 AND book_id = $3`,
			i, chapterID, bookID,
		)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			// Could return custom err, but for now we'll rely on the handler
		}
	}

	return tx.Commit(ctx)
}
