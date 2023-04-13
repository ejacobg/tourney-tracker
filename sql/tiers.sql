-- These are the default tiers the program will start with.
INSERT INTO tiers
VALUES (1, 'C', 75),
       (2, 'B', 150),
       (3, 'A', 200),
       (4, 'S', 300);

-- This query is used by postgres.TierService.GetTiers().
SELECT id, name, multiplier
FROM tiers;

-- This query is used by Model.GetTier().
SELECT tiers.id, tiers.name, multiplier
FROM tournaments
         INNER JOIN tiers on tournaments.tier_id = tiers.id
WHERE tournaments.id = 1;

-- This query is used by postgres.TierService.GetTier().
SELECT id, name, multiplier
FROM tiers
WHERE id = 1;

-- This query is used by postgres.TierService.GetTournamentTier().
SELECT tiers.id, tiers.name, tiers.multiplier
FROM tournaments
INNER JOIN tiers t on tournaments.tier_id = t.id
WHERE
-- This query is used by postgres.TierService.CreateTier().
INSERT INTO tiers (name, multiplier)
VALUES ('Z', 1)
RETURNING id;

-- This query is used by postgres.TierService.UpdateTier().
UPDATE tiers
SET name       = '1',
    multiplier = 2
WHERE id = 3;

-- This query is used by postgres.TierService.DeleteTier().
DELETE
FROM tiers
WHERE id = 1;