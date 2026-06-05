package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DictionaryRepository interface {
	GetDictionary(ctx context.Context, userID string) ([]string, error)
	AddWord(ctx context.Context, userID string, word string) error
}

type dictionaryRepository struct {
	pool *pgxpool.Pool
}

func NewDictionaryRepository(pool *pgxpool.Pool) DictionaryRepository {
	return &dictionaryRepository{pool: pool}
}

func (r *dictionaryRepository) GetDictionary(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		"SELECT word FROM user_dictionaries WHERE user_id = $1",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var words []string
	for rows.Next() {
		var word string
		if err := rows.Scan(&word); err != nil {
			continue
		}
		words = append(words, word)
	}

	return words, nil
}

func (r *dictionaryRepository) AddWord(ctx context.Context, userID string, word string) error {
	// Use ON CONFLICT DO NOTHING so adding an existing word doesn't throw an error.
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_dictionaries (user_id, word) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, word,
	)
	return err
}
