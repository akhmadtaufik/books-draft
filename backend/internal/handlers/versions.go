package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"novel-drafting-api/internal/models"
)

// VersionHandler holds dependencies for version-related HTTP handlers.
type VersionHandler struct {
	pool *pgxpool.Pool
}

// NewVersionHandler creates a new VersionHandler.
func NewVersionHandler(pool *pgxpool.Pool) *VersionHandler {
	return &VersionHandler{pool: pool}
}

// CreateVersion snapshots the current chapter content into chapter_versions.
// POST /api/chapters/{id}/versions
func (h *VersionHandler) CreateVersion(w http.ResponseWriter, r *http.Request) {
	chapterID := r.PathValue("id")
	if chapterID == "" {
		writeError(w, http.StatusBadRequest, "invalid chapter id")
		return
	}

	var req struct {
		SnapshotType string `json:"snapshot_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.SnapshotType != "session_end" && req.SnapshotType != "manual_milestone" {
		writeError(w, http.StatusBadRequest, "snapshot_type must be 'session_end' or 'manual_milestone'")
		return
	}

	ctx := r.Context()

	// Fetch the current chapter content.
	var currentContent json.RawMessage
	err := h.pool.QueryRow(ctx,
		"SELECT content FROM chapters WHERE id = $1",
		chapterID,
	).Scan(&currentContent)
	if err != nil {
		writeError(w, http.StatusNotFound, "chapter not found")
		return
	}

	// Insert a new version.
	var version models.ChapterVersion
	err = h.pool.QueryRow(ctx,
		`INSERT INTO chapter_versions (chapter_id, content, snapshot_type)
		 VALUES ($1, $2, $3)
		 RETURNING id, chapter_id, snapshot_type, created_at`,
		chapterID, currentContent, req.SnapshotType,
	).Scan(&version.ID, &version.ChapterID, &version.SnapshotType, &version.CreatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create version")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(version)
}

// ListVersions returns all versions for a chapter (without content), newest first.
// GET /api/chapters/{id}/versions
func (h *VersionHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	chapterID := r.PathValue("id")
	if chapterID == "" {
		writeError(w, http.StatusBadRequest, "invalid chapter id")
		return
	}

	rows, err := h.pool.Query(r.Context(),
		`SELECT id, chapter_id, snapshot_type, created_at
		 FROM chapter_versions
		 WHERE chapter_id = $1
		 ORDER BY created_at DESC`,
		chapterID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query versions")
		return
	}
	defer rows.Close()

	versions := make([]models.ChapterVersion, 0)
	for rows.Next() {
		var v models.ChapterVersion
		if err := rows.Scan(&v.ID, &v.ChapterID, &v.SnapshotType, &v.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan version")
			return
		}
		versions = append(versions, v)
	}

	json.NewEncoder(w).Encode(versions)
}

// GetVersion returns the full content of a specific version.
// GET /api/versions/{version_id}
func (h *VersionHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	versionID := r.PathValue("version_id")
	if versionID == "" {
		writeError(w, http.StatusBadRequest, "invalid version id")
		return
	}

	var version models.ChapterVersionFull
	err := h.pool.QueryRow(r.Context(),
		`SELECT id, chapter_id, content, snapshot_type, created_at
		 FROM chapter_versions WHERE id = $1`,
		versionID,
	).Scan(&version.ID, &version.ChapterID, &version.Content, &version.SnapshotType, &version.CreatedAt)
	if err != nil {
		writeError(w, http.StatusNotFound, "version not found")
		return
	}

	json.NewEncoder(w).Encode(version)
}

// DeleteVersion permanently removes a specific version snapshot.
// DELETE /api/versions/{version_id}
func (h *VersionHandler) DeleteVersion(w http.ResponseWriter, r *http.Request) {
	versionID := r.PathValue("version_id")
	if versionID == "" {
		writeError(w, http.StatusBadRequest, "invalid version id")
		return
	}

	tag, err := h.pool.Exec(r.Context(),
		"DELETE FROM chapter_versions WHERE id = $1",
		versionID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete version")
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "version not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
