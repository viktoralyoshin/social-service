-- +goose Up

CREATE SCHEMA IF NOT EXISTS social;

CREATE TABLE IF NOT EXISTS social.reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,      -- ID из Auth Service
    game_id UUID NOT NULL,      -- ID из Game Service
    rating INTEGER NOT NULL CHECK (rating >= 0 AND rating <= 100), -- Валидация на уровне БД
    text TEXT,                    -- Текст отзыва (может быть NULL, если просто оценка)
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- УНИКАЛЬНЫЙ ИНДЕКС: Юзер не может поставить 2 оценки одной игре
    CONSTRAINT unique_user_game UNIQUE (user_id, game_id)
);

-- +goose Down

DROP TABLE IF EXISTS social.reviews;
DROP SCHEMA IF EXISTS social;