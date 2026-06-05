package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"

	"novel-drafting-api/internal/models"
)

type VersionRepository interface {
	Create(ctx context.Context, chapterID string, snapshotType string) (models.ChapterVersion, error)
	List(ctx context.Context, chapterID string) ([]models.ChapterVersion, error)
	GetByID(ctx context.Context, id string) (models.ChapterVersionFull, error)
	Delete(ctx context.Context, id string) (int64, error)
}

type versionRepository struct {
	pool *pgxpool.Pool
}

func NewVersionRepository(pool *pgxpool.Pool) VersionRepository {
	return &versionRepository{pool: pool}
}

func (r *versionRepository) Create(ctx context.Context, chapterID string, snapshotType string) (models.ChapterVersion, error) {
	// Fetch the current chapter content.
	var currentContent json.RawMessage
	err := r.pool.QueryRow(ctx,
		"SELECT content FROM chapters WHERE id = $1",
		chapterID,
	).Scan(&currentContent)
	if err != nil {
		return models.ChapterVersion{}, err
	}

	// Insert a new version.
	var version models.ChapterVersion
	err = r.pool.QueryRow(ctx,
		`INSERT INTO chapter_versions (chapter_id, content, snapshot_type)
		 VALUES ($1, $2, $3)
		 RETURNING id, chapter_id, snapshot_type, created_at`,
		chapterID, currentContent, snapshotType,
	).Scan(&version.ID, &version.ChapterID, &version.SnapshotType, &version.CreatedAt)
	
	return version, err
}

func (r *versionRepository) List(ctx context.Context, chapterID string) ([]models.ChapterVersion, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, chapter_id, snapshot_type, created_at
		 FROM chapter_versions
		 WHERE chapter_id = $1
		 ORDER BY created_at DESC`,
		chapterID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make([]models.ChapterVersion, 0)
	for rows.Next() {
		var v models.ChapterVersion
		if err := rows.Scan(&v.ID, &v.ChapterID, &v.SnapshotType, &v.CreatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}

	return versions, nil
}

func (r *versionRepository) GetByID(ctx context.Context, id string) (models.ChapterVersionFull, error) {
	var version models.ChapterVersionFull
	err := r.pool.QueryRow(ctx,
		`SELECT id, chapter_id, content, snapshot_type, created_at
		 FROM chapter_versions WHERE id = $1`,
		id,
	).Scan(&version.ID, &version.ChapterID, &version.Content, &version.SnapshotType, &version.CreatedAt)
	return version, err
}

func (r *versionRepository) Delete(ctx context.Context, id string) (int64, error) {
	tag, err := r.pool.Exec(ctx,
		"DELETE FROM chapter_versions WHERE id = $1",
		id,
	)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
