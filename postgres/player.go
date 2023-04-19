package postgres

import (
	"database/sql"
	"errors"
	tournament "github.com/ejacobg/tourney-tracker"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// PlayerService represents a service for managing players.
type PlayerService struct {
	DB *sql.DB
}

func (ps PlayerService) GetPlayers() (players []tournament.Player, err error) {
	query := `
SELECT id, name
FROM players`

	rows, err := ps.DB.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var player tournament.Player

		err = rows.Scan(&player.ID, &player.Name)
		if err != nil {
			return
		}

		players = append(players, player)
	}

	return players, rows.Err()
}

func (ps PlayerService) GetPlayer(id int64) (player tournament.Player, err error) {
	query := `
SELECT id, name
FROM players
WHERE id = $1`

	err = ps.DB.QueryRow(query, id).Scan(&player.ID, &player.Name)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		err = ErrRecordNotFound
	}

	return
}

func (ps PlayerService) GetRanks() ([]tournament.Rank, error) {
	query := `
SELECT players.id,
       players.name,
       entrants.placement,
       idx(tournaments.placements, entrants.placement),
       tournaments.bracket_reset,
       tiers.multiplier
FROM players
         LEFT OUTER JOIN entrants on entrants.player_id = players.id
         LEFT OUTER JOIN tournaments on tournaments.id = entrants.tournament_id
         LEFT OUTER JOIN tiers on tiers.id = tournaments.tier_id`

	var (
		placement    sql.NullInt64
		PV           sql.NullInt64
		bracketReset sql.NullBool
		multiplier   sql.NullInt64
	)

	// Map player IDs to their rank.
	ranks := make(map[int64]tournament.Rank)

	rows, err := ps.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rank tournament.Rank

		err = rows.Scan(&rank.Player.ID, &rank.Player.Name, &placement, &PV, &bracketReset, &multiplier)
		if err != nil {
			return nil, err
		}

		// If the placement value isn't valid, then we can't calculate any points.
		if !placement.Valid {
			// If we haven't seen this player before, give them 0 points.
			if _, ok := ranks[rank.Player.ID]; !ok {
				ranks[rank.Player.ID] = rank
			}
			continue
		}

		// Calculate the points earned.
		// Note: PostgreSQL uses 1-based indices, so we have to subtract the PV by 1.
		rank.Points = tournament.UP*int(PV.Int64-1) + tournament.ATT
		if placement.Int64 == 1 {
			rank.Points += tournament.FIRST
		} else if placement.Int64 == 2 && bracketReset.Bool {
			rank.Points += tournament.BR
		}
		rank.Points *= int(multiplier.Int64)

		// Add the calculated points to the appropriate player.
		rank.Points += ranks[rank.Player.ID].Points
		ranks[rank.Player.ID] = rank
	}

	// Sort our ranks in descending order.
	unsorted := maps.Values(ranks)
	slices.SortFunc(unsorted, func(a, b tournament.Rank) bool {
		return a.Points > b.Points
	})

	return unsorted, nil
}

func (ps PlayerService) CreatePlayer(player *tournament.Player) error {
	query := `
INSERT INTO players (name)
VALUES ($1)
RETURNING id`

	err := ps.DB.QueryRow(query, player.Name).Scan(&player.ID)

	return err
}

func (ps PlayerService) UpdatePlayer(player *tournament.Player) error {
	query := `UPDATE players
SET name = $2
WHERE id = $1`

	_, err := ps.DB.Exec(query, player.ID, player.Name)

	return err
}

func (ps PlayerService) DeletePlayer(id int64) error {
	query := `
DELETE FROM players
WHERE id = $1`

	_, err := ps.DB.Exec(query, id)

	// Due to the way the database is set up, deleting a Player will automatically nullify any entrants pointing to it.
	return err
}
