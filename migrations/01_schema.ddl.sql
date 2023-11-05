-- Add migration script here
SET TIMEZONE TO 'Europe/Prague';

CREATE TABLE IF NOT EXISTS quote (
    id UUID PRIMARY KEY,
    book VARCHAR(63) NOT NULL,
    quote VARCHAR(255) NOT NULL,
    inserted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(book, quote)
)