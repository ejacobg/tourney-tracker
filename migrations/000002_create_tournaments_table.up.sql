CREATE TABLE IF NOT EXISTS tournaments
(
    id            bigserial PRIMARY KEY,
    name          text                         NOT NULL,
    url           text                         NOT NULL,
    bracket_reset boolean                      NOT NULL,
    placements    integer[]                    NOT NULL,
    tier_id       bigint REFERENCES tiers (id) NOT NULL
);