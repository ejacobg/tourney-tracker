package tournament

import (
	"database/sql"
	"github.com/lib/pq"
)

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
