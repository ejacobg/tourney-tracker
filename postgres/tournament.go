package postgres

import (
	"database/sql"
	"errors"
	tournament "github.com/ejacobg/tourney-tracker"
	"github.com/lib/pq"
)

// TournamentService represents a service for managing tournaments.
type TournamentService struct {
	DB *sql.DB
}

func (ts TournamentService) GetPreviews() (previews []tournament.Preview, err error) {
	query := `
SELECT tournaments.id, tournaments.name, tiers.name
FROM tournaments
INNER JOIN tiers on tiers.id = tournaments.tier_id`

	rows, err := ts.DB.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var preview tournament.Preview

		err = rows.Scan(&preview.ID, &preview.Name, &preview.Tier)
		if err != nil {
			return
		}

		previews = append(previews, preview)
	}

	return previews, rows.Err()
}

func (ts TournamentService) GetNamesByTier(tierID int64) (names []tournament.Name, err error) {
	query := `
SELECT id, name
FROM tournaments
WHERE tier_id = $1`

	rows, err := ts.DB.Query(query, tierID)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name tournament.Name

		err = rows.Scan(&name.ID, &name.Name)
		if err != nil {
			return
		}

		names = append(names, name)
	}

	return names, rows.Err()
}

func (ts TournamentService) GetTournament(id int64) (tourney tournament.Tournament, err error) {
	if id < 1 {
		return tourney, ErrRecordNotFound
	}

	query := `
SELECT tournaments.id, tournaments.name, url, bracket_reset, placements, tier_id, tiers.name, tiers.multiplier
FROM tournaments INNER JOIN tiers ON tier_id = tiers.id
WHERE tournaments.id = $1;`

	err = ts.DB.QueryRow(query, id).Scan(
		&tourney.ID,
		&tourney.Name,
		&tourney.URL,
		&tourney.BracketReset,
		pq.Array(&tourney.Placements),
		&tourney.Tier.ID,
		&tourney.Tier.Name,
		&tourney.Tier.Multiplier,
	)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		err = ErrRecordNotFound
	}

	return
}

func (ts TournamentService) CreateTournament(tourney *tournament.Tournament, entrants []tournament.Entrant) error {
	tx, err := ts.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = createTournament(tx, tourney)
	if err != nil {
		return err
	}

	err = createEntrants(tx, entrants, tourney.ID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (ts TournamentService) SetTier(tournamentID, tierID int64) error {
	query := `
UPDATE tournaments
SET tier_id = $2
WHERE id = $1`

	_, err := ts.DB.Exec(query, tournamentID, tierID)

	return err
}

func (ts TournamentService) DeleteTournament(id int64) error {
	query := `
DELETE FROM tournaments
WHERE id = $1`

	_, err := ts.DB.Exec(query, id)

	return err
}

func (ts TournamentService) Insert(tourney *tournament.Tournament, entrants []tournament.Entrant) error {
	tx, err := ts.DB.Begin()
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

	err = ts.DB.QueryRow(query, tourney.Name, tourney.URL, tourney.BracketReset, pq.Array(tourney.Placements)).
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
		err = ts.DB.QueryRow(query, entrant.Name, entrant.Placement, tourney.ID).Scan(&entrants[i].ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (ts TournamentService) Get(id int64) (tourney tournament.Tournament, entrants []tournament.Entrant, err error) {
	if id < 1 {
		return tourney, entrants, ErrRecordNotFound
	}

	query := `
SELECT tournaments.id, tournaments.name, url, bracket_reset, placements, tier_id, tiers.name, tiers.multiplier
FROM tournaments INNER JOIN tiers ON tier_id = tiers.id
WHERE tournaments.id = $1;`

	err = ts.DB.QueryRow(query, id).Scan(
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

	rows, err := ts.DB.Query(query, tourney.ID)
	if err != nil {
		return
	}

	for rows.Next() {
		var entrant tournament.Entrant

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

func (ts TournamentService) GetTiers() (tiers []tournament.Tier, err error) {
	query := `
SELECT id, name, multiplier
FROM tiers;`

	rows, err := ts.DB.Query(query)
	if err != nil {
		return
	}

	for rows.Next() {
		var tier tournament.Tier

		err = rows.Scan(&tier.ID, &tier.Name, &tier.Multiplier)
		if err != nil {
			return
		}

		tiers = append(tiers, tier)
	}

	return tiers, rows.Err()
}

func (ts TournamentService) GetTier(tournamentID int64) (tier tournament.Tier, err error) {
	query := `
SELECT tiers.id, tiers.name, multiplier
FROM tournaments
INNER JOIN tiers on tournaments.tier_id = tiers.id
WHERE tournaments.id = $1;`

	return tier, ts.DB.QueryRow(query, tournamentID).Scan(&tier.ID, &tier.Name, &tier.Multiplier)
}

// Due to the way the database is set up, deleting a Tournament will also delete its entrants.

func createTournament(tx *sql.Tx, tourney *tournament.Tournament) error {
	query := `
WITH tourney AS (
    INSERT INTO tournaments (name, url, bracket_reset, placements, tier_id)
        VALUES ($1, $2, $3, $4, 1)
        RETURNING id, tier_id)
SELECT tourney.id, tourney.tier_id, tiers.name, tiers.multiplier
FROM tourney
         INNER JOIN tiers on tourney.tier_id = 1`

	return tx.QueryRow(query, tourney.Name, tourney.URL, tourney.BracketReset, pq.Array(tourney.Placements)).
		Scan(&tourney.ID, &tourney.Tier.ID, &tourney.Tier.Name, &tourney.Tier.Multiplier)
}
