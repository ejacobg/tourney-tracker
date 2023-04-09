package tournament

import (
	"database/sql"
	"errors"
	"github.com/lib/pq"
)

var ErrRecordNotFound = errors.New("record not found")

// Model provides several methods for interacting with the database.
type Model struct {
	DB *sql.DB
}

// Preview represents a subset of a Tournament object, namely its ID, name, and tier.
type Preview struct {
	ID         int64
	Tournament string
	Tier       string
}

func (m Model) GetPreviews() (previews []Preview, err error) {
	query := `
SELECT tournaments.id, tournaments.name, tiers.name
FROM tournaments
INNER JOIN tiers on tiers.id = tournaments.tier_id`

	rows, err := m.DB.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var preview Preview

		err = rows.Scan(&preview.ID, &preview.Tournament, &preview.Tier)
		if err != nil {
			return
		}

		previews = append(previews, preview)
	}

	return previews, rows.Err()
}

func (m Model) Insert(tourney *Tournament, entrants []Entrant) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Hard-coding the tier ID. Right now, I'm assuming that the C-tier ID will always exist.
	// A better solution might be to write a sub-query for these.
	query := `
WITH tourney AS (
    INSERT INTO tournaments (name, url, bracket_reset, placements, tier_id)
        VALUES ($1, $2, $3, $4, 1)
        RETURNING id, tier_id)
SELECT tourney.id, tourney.tier_id, tiers.name, tiers.multiplier
FROM tourney
         INNER JOIN tiers on tourney.tier_id = 1;`

	err = m.DB.QueryRow(query, tourney.Name, tourney.URL, tourney.BracketReset, pq.Array(tourney.Placements)).
		Scan(&tourney.ID, &tourney.Tier.ID, &tourney.Tier.Name, &tourney.Tier.Multiplier)
	if err != nil {
		return err
	}

	query = `
INSERT INTO entrants (name, placement, tournament_id)
VALUES ($1, $2, $3)
RETURNING id;`

	for i, entrant := range entrants {
		// This will update the entrant IDs as it goes along. If any errors occur, any written IDs will be invalidated.
		err = m.DB.QueryRow(query, entrant.Name, entrant.Placement, tourney.ID).Scan(&entrants[i].ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (m Model) Get(id int64) (tourney Tournament, entrants []Entrant, err error) {
	if id < 1 {
		return tourney, entrants, ErrRecordNotFound
	}

	query := `
SELECT tournaments.id, tournaments.name, url, bracket_reset, placements, tier_id, tiers.name, tiers.multiplier
FROM tournaments INNER JOIN tiers ON tier_id = tiers.id
WHERE tournaments.id = $1;`

	err = m.DB.QueryRow(query, id).Scan(
		&tourney.ID,
		&tourney.Name,
		&tourney.URL,
		&tourney.BracketReset,
		pq.Array(&tourney.Placements),
		&tourney.Tier.ID,
		&tourney.Tier.Name,
		&tourney.Tier.Multiplier,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrRecordNotFound
		}
		return
	}

	query = `
SELECT entrants.id, entrants.name, placement, tournament_id, players.name
FROM entrants LEFT OUTER JOIN players ON entrants.player_id = players.id
WHERE tournament_id = $1;`

	rows, err := m.DB.Query(query, tourney.ID)
	if err != nil {
		return
	}

	for rows.Next() {
		var entrant Entrant

		err = rows.Scan(
			&entrant.ID,
			&entrant.Name,
			&entrant.Placement,
			&entrant.TournamentID,
			&entrant.PlayerName,
		)

		if err != nil {
			return
		}

		entrants = append(entrants, entrant)
	}

	return tourney, entrants, rows.Err()
}
