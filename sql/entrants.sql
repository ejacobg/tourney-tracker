-- This is the query used by Model.Insert().
INSERT INTO entrants (name, placement, tournament_id)
VALUES ('test', 1, 1)
RETURNING id;

-- This is the query used by postgres.EntrantService.GetEntrants().
SELECT entrants.id, entrants.name, placement, tournament_id, players.name
FROM entrants
         LEFT OUTER JOIN players ON entrants.player_id = players.id
WHERE tournament_id = 1;

-- This is the query used by postgres.EntrantService.GetAttendance().
SELECT tournaments.id, tournaments.name, tiers.name, entrants.name, entrants.placement
FROM entrants
         LEFT OUTER JOIN tournaments on entrants.tournament_id = tournaments.id
         LEFT OUTER JOIN tiers on tournaments.tier_id = tiers.id
WHERE entrants.player_id = 1;

-- This is the query used by postgres.EntrantService.SetPlayer().
UPDATE entrants
SET player_id = 1
WHERE id = 1;

-- This is the query used by postgres.EntrantService.DeleteEntrants().
DELETE FROM entrants
WHERE tournament_id = 1;