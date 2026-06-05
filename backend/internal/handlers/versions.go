package handlers

import (
	"encoding/json"
	"net/http"

	"novel-drafting-api/internal/repository"
)

// VersionHandler holds dependencies for version-related HTTP handlers.
type VersionHandler struct {
	repo repository.VersionRepository
}

// NewVersionHandler creates a new VersionHandler.
func NewVersionHandler(repo repository.VersionRepository) *VersionHandler {
	return &VersionHandler{repo: repo}
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

	version, err := h.repo.Create(r.Context(), chapterID, req.SnapshotType)
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

	versions, err := h.repo.List(r.Context(), chapterID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query versions")
		return
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

	version, err := h.repo.GetByID(r.Context(), versionID)
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

	rowsAffected, err := h.repo.Delete(r.Context(), versionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete version")
		return
	}
	if rowsAffected == 0 {
		writeError(w, http.StatusNotFound, "version not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
