SELECT * FROM players;

SELECT entrants.placement, idx(tournaments.placements, entrants.placement)
FROM entrants LEFT OUTER JOIN tournaments on tournaments.id = entrants.tournament_id
WHERE entrants.tournament_id = 14;

-- This query is used by postgres.PlayerService.GetRanks().
SELECT players.id,
       players.name,
       entrants.placement,
       idx(tournaments.placements, entrants.placement),
       tournaments.bracket_reset,
       tiers.multiplier
FROM players
         LEFT OUTER JOIN entrants on entrants.player_id = players.id
         LEFT OUTER JOIN tournaments on tournaments.id = entrants.tournament_id
         LEFT OUTER JOIN tiers on tiers.id = tournaments.tier_id;

