CREATE TABLE IF NOT EXISTS tiers
(
    id         bigserial PRIMARY KEY,
    name       text    NOT NULL,
    multiplier integer NOT NULL
);