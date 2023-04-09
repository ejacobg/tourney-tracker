-- These are some example tournaments, mainly used for testing.
INSERT INTO tournaments (name, url, bracket_reset, placements, tier_id)
VALUES ('(SSC C TIER) Gator Grind #9', 'https://challonge.com/kpqlgghc', false, '{17, 13, 9, 7, 5, 4, 3, 2, 1}', 1),
       ('(SSC C Tier) Gator Grind #12', 'https://challonge.com/8ozc6ffz', false, '{17, 13, 9, 7, 5, 4, 3, 2, 1}', 1),
       ('(SSC C Tier) Gator Grind #7', 'https://challonge.com/t4kq4f5b', true, '{17, 13, 9, 7, 5, 4, 3, 2, 1}', 1),
       ('Silver State Smash x Pirate Hackers Black Lives Matter Charity Tournament - Singles 1v1',
        'https://start.gg/tournament/silver-state-smash-x-pirate-hackers-black-lives-matter-charity/event/singles-1v1',
        false, '{33, 17, 13, 9, 7, 5, 4, 3, 2, 1}', 3),
       ('Wrangler Rumble #1 - Ultimate Singles', 'https://start.gg/tournament/wrangler-rumble-1/event/ultimate-singles',
        false, '{13, 9, 7, 5, 4, 3, 2, 1}', 2),
       ('Shinto Series: Smash #1 - Singles 1v1', 'https://start.gg/tournament/shinto-series-smash-1/event/singles-1v1',
        true, '{97, 65, 49, 33, 25, 17, 13, 9, 7, 5, 4, 3, 2, 1}', 1);

-- This is the query used by Model.Insert().
WITH tourney AS (
    INSERT INTO tournaments (name, url, bracket_reset, placements, tier_id)
        VALUES ('', '', false, '{}', 1)
        RETURNING id, tier_id)
SELECT tourney.id, tourney.tier_id, tiers.name, tiers.multiplier
FROM tourney
         INNER JOIN tiers on tourney.tier_id = 1;

-- This is the query used by Model.Get().
SELECT tournaments.id, tournaments.name, url, bracket_reset, placements, tier_id, tiers.name, tiers.multiplier
FROM tournaments INNER JOIN tiers ON tier_id = tiers.id
WHERE tournaments.id = 1;