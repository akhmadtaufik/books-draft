package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"novel-drafting-api/internal/models"
	"novel-drafting-api/internal/repository"
)

type DictionaryHandler struct {
	repo          repository.DictionaryRepository
	defaultUserID string
}

func NewDictionaryHandler(repo repository.DictionaryRepository, defaultUserID string) *DictionaryHandler {
	return &DictionaryHandler{
		repo:          repo,
		defaultUserID: defaultUserID,
	}
}

func (h *DictionaryHandler) GetDictionary(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	words, err := h.repo.GetDictionary(ctx, h.defaultUserID)
	if err != nil {
		http.Error(w, `{"error": "failed to query dictionary"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"words": words,
	})
}

func (h *DictionaryHandler) AddWord(w http.ResponseWriter, r *http.Request) {
	var req models.DictionaryWordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Word == "" {
		http.Error(w, `{"error": "invalid request body or missing word"}`, http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	err := h.repo.AddWord(ctx, h.defaultUserID, req.Word)
	if err != nil {
		http.Error(w, `{"error": "failed to insert word"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "word added to dictionary",
		"word":    req.Word,
	})
}
