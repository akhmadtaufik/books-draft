-- =============================================================
-- Novel Drafting App – Database Initialization Script
-- Matches PRD Section 3 (PostgreSQL)
-- =============================================================

-- Enable the pgcrypto extension for gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- -----------------------------------------------------------
-- 1. users
-- -----------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMP   NOT NULL DEFAULT NOW()
);

-- -----------------------------------------------------------
-- 2. books
-- -----------------------------------------------------------
CREATE TABLE IF NOT EXISTS books (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title           VARCHAR(500) NOT NULL,
    cover_image_url VARCHAR(2048),
    metadata        JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_books_user_id ON books(user_id);

-- -----------------------------------------------------------
-- 3. chapters
-- -----------------------------------------------------------
CREATE TABLE IF NOT EXISTS chapters (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    book_id         UUID        NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    title           VARCHAR(500) NOT NULL,
    content         JSONB       NOT NULL DEFAULT '{}'::jsonb,
    position_index  INTEGER     NOT NULL DEFAULT 0,
    updated_at      TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chapters_book_id ON chapters(book_id);
CREATE INDEX IF NOT EXISTS idx_chapters_position ON chapters(book_id, position_index);

-- GIN index on JSONB content for future full-text / structural queries
CREATE INDEX IF NOT EXISTS idx_chapters_content ON chapters USING GIN (content);

-- -----------------------------------------------------------
-- 4. user_dictionaries (Phase 2 – custom spell-check words)
-- -----------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_dictionaries (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    word       VARCHAR(255) NOT NULL,
    created_at TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_dict_user_id ON user_dictionaries(user_id);
-- Prevent duplicate words per user
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_dict_unique_word ON user_dictionaries(user_id, word);

-- -----------------------------------------------------------
-- 5. chapter_versions (Revision History snapshots)
-- -----------------------------------------------------------
CREATE TABLE IF NOT EXISTS chapter_versions (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_id    UUID        NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    content       JSONB       NOT NULL DEFAULT '{}'::jsonb,
    snapshot_type VARCHAR(50) NOT NULL CHECK (snapshot_type IN ('session_end', 'manual_milestone')),
    created_at    TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chapter_versions_chapter_id ON chapter_versions(chapter_id);
CREATE INDEX IF NOT EXISTS idx_chapter_versions_created_at ON chapter_versions(chapter_id, created_at DESC);

-- -----------------------------------------------------------
-- 6. book_notes (Plot Management – Character Sheets & Wiki)
-- -----------------------------------------------------------
CREATE TABLE IF NOT EXISTS book_notes (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    book_id    UUID        NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    title      VARCHAR(255) NOT NULL,
    type       VARCHAR(50) NOT NULL CHECK (type IN ('character', 'worldbuilding')),
    content    JSONB       NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_book_notes_book_id ON book_notes(book_id);
CREATE INDEX IF NOT EXISTS idx_book_notes_type ON book_notes(book_id, type);
