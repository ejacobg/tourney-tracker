CREATE TABLE IF NOT EXISTS entrants
(
    id            bigserial PRIMARY KEY,
    name          text    NOT NULL,
    placement     integer NOT NULL,
    tournament_id bigint  NOT NULL REFERENCES tournaments (id) ON DELETE CASCADE,
    player_id     bigint  REFERENCES players (id) ON DELETE SET NULL,
    UNIQUE (tournament_id, player_id)
);