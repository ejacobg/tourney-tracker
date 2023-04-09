-- This is the query used by Model.Insert().
INSERT INTO entrants (name, placement, tournament_id)
VALUES ('test', 1, 1)
RETURNING id;

-- This is the query used by Model.Get().
SELECT entrants.id, entrants.name, placement, tournament_id, players.name
FROM entrants LEFT OUTER JOIN players ON entrants.player_id = players.id
WHERE tournament_id = 1;