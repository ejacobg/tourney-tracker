-- This is the query used by Model.Insert().
INSERT INTO entrants (name, placement, tournament_id)
VALUES ('test', 1, 1)
RETURNING id;