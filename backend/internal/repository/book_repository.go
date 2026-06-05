package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"novel-drafting-api/internal/models"
)

type BookRepository interface {
	GetAll(ctx context.Context, userID string) ([]models.Book, error)
	GetByID(ctx context.Context, userID string, bookID string) (models.Book, error)
	Create(ctx context.Context, userID string, book *models.Book) (models.Book, error)
	Update(ctx context.Context, userID string, bookID string, book *models.Book) (models.Book, error)
	Delete(ctx context.Context, userID string, bookID string) (int64, error)
	GetPreview(ctx context.Context, userID string, bookID string) (models.BookPreview, error)
}

type bookRepository struct {
	pool *pgxpool.Pool
}

func NewBookRepository(pool *pgxpool.Pool) BookRepository {
	return &bookRepository{pool: pool}
}

func (r *bookRepository) GetAll(ctx context.Context, userID string) ([]models.Book, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, title, author, synopsis, genre, language, isbn, publisher, status, cover_image_url, metadata, created_at, updated_at
		 FROM books WHERE user_id = $1
		 ORDER BY updated_at DESC, created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := make([]models.Book, 0)
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Synopsis, &b.Genre, &b.Language, &b.ISBN, &b.Publisher, &b.Status, &b.CoverImageURL, &b.Metadata, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		books = append(books, b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}

func (r *bookRepository) GetByID(ctx context.Context, userID string, bookID string) (models.Book, error) {
	var b models.Book
	err := r.pool.QueryRow(ctx,
		`SELECT id, title, author, synopsis, genre, language, isbn, publisher, status, cover_image_url, metadata, created_at, updated_at
		 FROM books WHERE user_id = $1 AND id = $2`,
		userID, bookID,
	).Scan(&b.ID, &b.Title, &b.Author, &b.Synopsis, &b.Genre, &b.Language, &b.ISBN, &b.Publisher, &b.Status, &b.CoverImageURL, &b.Metadata, &b.CreatedAt, &b.UpdatedAt)
	return b, err
}

func (r *bookRepository) Create(ctx context.Context, userID string, book *models.Book) (models.Book, error) {
	var b models.Book
	err := r.pool.QueryRow(ctx,
		`INSERT INTO books (user_id, title, author, synopsis, genre, language, isbn, publisher, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, user_id, title, author, synopsis, genre, language, isbn, publisher, status, cover_image_url, metadata, created_at, updated_at`,
		userID, book.Title, book.Author, book.Synopsis, book.Genre, book.Language, book.ISBN, book.Publisher, book.Status,
	).Scan(&b.ID, &b.UserID, &b.Title, &b.Author, &b.Synopsis, &b.Genre, &b.Language, &b.ISBN, &b.Publisher, &b.Status, &b.CoverImageURL, &b.Metadata, &b.CreatedAt, &b.UpdatedAt)
	return b, err
}

func (r *bookRepository) Update(ctx context.Context, userID string, bookID string, book *models.Book) (models.Book, error) {
	var b models.Book
	err := r.pool.QueryRow(ctx,
		`UPDATE books 
		 SET title = $1, author = $2, synopsis = $3, genre = $4, language = $5, isbn = $6, publisher = $7, status = $8, updated_at = NOW()
		 WHERE id = $9 AND user_id = $10
		 RETURNING id, user_id, title, author, synopsis, genre, language, isbn, publisher, status, cover_image_url, metadata, created_at, updated_at`,
		book.Title, book.Author, book.Synopsis, book.Genre, book.Language, book.ISBN, book.Publisher, book.Status, bookID, userID,
	).Scan(&b.ID, &b.UserID, &b.Title, &b.Author, &b.Synopsis, &b.Genre, &b.Language, &b.ISBN, &b.Publisher, &b.Status, &b.CoverImageURL, &b.Metadata, &b.CreatedAt, &b.UpdatedAt)
	return b, err
}

func (r *bookRepository) Delete(ctx context.Context, userID string, bookID string) (int64, error) {
	tag, err := r.pool.Exec(ctx,
		"DELETE FROM books WHERE id = $1 AND user_id = $2",
		bookID, userID,
	)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *bookRepository) GetPreview(ctx context.Context, userID string, bookID string) (models.BookPreview, error) {
	var title string
	err := r.pool.QueryRow(ctx,
		"SELECT title FROM books WHERE id = $1 AND user_id = $2",
		bookID, userID,
	).Scan(&title)
	if err != nil {
		return models.BookPreview{}, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT title, content FROM chapters
		 WHERE book_id = $1
		 ORDER BY position_index ASC`,
		bookID,
	)
	if err != nil {
		return models.BookPreview{}, err
	}
	defer rows.Close()

	chapters := make([]models.ChapterPreview, 0)
	for rows.Next() {
		var ch models.ChapterPreview
		if err := rows.Scan(&ch.Title, &ch.Content); err != nil {
			return models.BookPreview{}, err
		}
		chapters = append(chapters, ch)
	}

	return models.BookPreview{
		Title:    title,
		Chapters: chapters,
	}, nil
}
