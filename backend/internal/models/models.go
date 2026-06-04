package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// User represents a row in the users table.
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

// Book represents a row in the books table.
type Book struct {
	ID            uuid.UUID        `json:"id"`
	UserID        uuid.UUID        `json:"userId"`
	Title         string           `json:"title"`
	CoverImageURL *string          `json:"coverImageUrl"`
	Metadata      json.RawMessage  `json:"metadata"`
	CreatedAt     time.Time        `json:"createdAt"`
}

// Chapter represents a full row in the chapters table (includes content).
type Chapter struct {
	ID            uuid.UUID       `json:"id"`
	BookID        uuid.UUID       `json:"bookId"`
	Title         string          `json:"title"`
	Content       json.RawMessage `json:"content"`
	PositionIndex int             `json:"positionIndex"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

// ChapterListItem is a lightweight chapter representation for list endpoints.
type ChapterListItem struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	PositionIndex int       `json:"positionIndex"`
}

// ReorderRequest is the payload for the chapter reorder endpoint.
type ReorderRequest struct {
	BookID     string   `json:"bookId"`
	OrderedIDs []string `json:"orderedIds"`
}

// BookPreview is the response for the book preview endpoint.
type BookPreview struct {
	Title    string           `json:"title"`
	Chapters []ChapterPreview `json:"chapters"`
}

// ChapterPreview contains title + content for the preview endpoint.
type ChapterPreview struct {
	Title   string          `json:"title"`
	Content json.RawMessage `json:"content"`
}

// UserDictionary represents a custom word added by a user.
type UserDictionary struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"userId"`
	Word   string    `json:"word"`
}

// DictionaryWordRequest is the payload for adding a new dictionary word.
type DictionaryWordRequest struct {
	Word string `json:"word"`
}

// ChapterVersion is a lightweight version record (no content) for list endpoints.
type ChapterVersion struct {
	ID           uuid.UUID `json:"id"`
	ChapterID    uuid.UUID `json:"chapterId"`
	SnapshotType string    `json:"snapshotType"`
	CreatedAt    time.Time `json:"createdAt"`
}

// ChapterVersionFull includes the full JSONB content for a single version fetch.
type ChapterVersionFull struct {
	ID           uuid.UUID       `json:"id"`
	ChapterID    uuid.UUID       `json:"chapterId"`
	Content      json.RawMessage `json:"content"`
	SnapshotType string          `json:"snapshotType"`
	CreatedAt    time.Time       `json:"createdAt"`
}

// BookNote represents a row in the book_notes table (character sheet or worldbuilding entry).
type BookNote struct {
	ID        uuid.UUID       `json:"id"`
	BookID    uuid.UUID       `json:"bookId"`
	Title     string          `json:"title"`
	Type      string          `json:"type"`
	Content   json.RawMessage `json:"content"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

// CreateNoteRequest is the payload for creating a new book note.
type CreateNoteRequest struct {
	Title   string          `json:"title"`
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
}

// UpdateNoteRequest is the payload for updating an existing book note.
type UpdateNoteRequest struct {
	Title   string          `json:"title"`
	Content json.RawMessage `json:"content"`
}
