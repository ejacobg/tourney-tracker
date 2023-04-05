CREATE TABLE IF NOT EXISTS players
(
    id   bigserial PRIMARY KEY,
    name text UNIQUE NOT NULL
);