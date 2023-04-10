-- These are the default tiers the program will start with.
INSERT INTO tiers
VALUES (1, 'C', 75),
       (2, 'B', 150),
       (3, 'A', 200),
       (4, 'S', 300);

-- This query is used by Model.GetTiers().
SELECT id, name, multiplier
FROM tiers;

-- This query is used by Model.GetTier().
SELECT tiers.id, tiers.name, multiplier
FROM tournaments
INNER JOIN tiers on tournaments.tier_id = tiers.id
WHERE tournaments.id = 1;