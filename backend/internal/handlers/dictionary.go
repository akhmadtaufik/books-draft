package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"novel-drafting-api/internal/models"
)

type DictionaryHandler struct {
	pool          *pgxpool.Pool
	defaultUserID string
}

func NewDictionaryHandler(pool *pgxpool.Pool, defaultUserID string) *DictionaryHandler {
	return &DictionaryHandler{
		pool:          pool,
		defaultUserID: defaultUserID,
	}
}

func (h *DictionaryHandler) GetDictionary(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	rows, err := h.pool.Query(ctx,
		"SELECT word FROM user_dictionaries WHERE user_id = $1",
		h.defaultUserID,
	)
	if err != nil {
		http.Error(w, `{"error": "failed to query dictionary"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	words := []string{}
	for rows.Next() {
		var word string
		if err := rows.Scan(&word); err != nil {
			continue
		}
		words = append(words, word)
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

	// Use ON CONFLICT DO NOTHING so adding an existing word doesn't throw an error.
	_, err := h.pool.Exec(ctx,
		`INSERT INTO user_dictionaries (user_id, word) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		h.defaultUserID, req.Word,
	)
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
