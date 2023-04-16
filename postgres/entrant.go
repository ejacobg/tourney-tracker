package postgres

import (
	"database/sql"
	"errors"
	tournament "github.com/ejacobg/tourney-tracker"
	"golang.org/x/exp/slices"
)

// EntrantService represents a service for managing entrants.
type EntrantService struct {
	DB *sql.DB
}

func (es EntrantService) GetEntrants(tournamentID int64) (entrants []tournament.Entrant, err error) {
	query := `
SELECT entrants.id, entrants.name, placement, tournament_id, players.name
FROM entrants LEFT OUTER JOIN players ON entrants.player_id = players.id
WHERE tournament_id = $1`

	rows, err := es.DB.Query(query, tournamentID)
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

	return entrants, rows.Err()
}

func (es EntrantService) GetEntrantWithPoints(id int64) (entrant tournament.Entrant, points int, err error) {
	// Get Entrant and Tournament.
	tx, err := es.DB.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	entrant, err = getEntrant(tx, id)
	if err != nil {
		return
	}

	tourney, err := getTournament(tx, entrant.TournamentID)
	if err != nil {
		return
	}

	err = tx.Commit()
	if err != nil {
		return
	}

	// Calculate points.
	PV := slices.Index(tourney.Placements, entrant.Placement)
	if PV == -1 {
		panic("could not find entrant placement")
	}

	points = tournament.UP*PV + tournament.ATT
	if entrant.Placement == 1 {
		points += tournament.FIRST
	} else if entrant.Placement == 2 && tourney.BracketReset {
		points += tournament.BR
	}
	points *= tourney.Tier.Multiplier

	return
}

func (es EntrantService) GetAttendance(playerID int64) (attendance []tournament.Attendee, err error) {
	query := `
SELECT tournaments.id, tournaments.name, tiers.name, entrants.name, entrants.placement
FROM entrants
         LEFT OUTER JOIN tournaments on entrants.tournament_id = tournaments.id
         LEFT OUTER JOIN tiers on tournaments.tier_id = tiers.id
WHERE entrants.player_id = $1`

	rows, err := es.DB.Query(query, playerID)
	if err != nil {
		return
	}

	for rows.Next() {
		var attendee tournament.Attendee

		err = rows.Scan(
			&attendee.Tournament.ID,
			&attendee.Tournament.Name,
			&attendee.Tournament.Tier,
			&attendee.Entrant.Name,
			&attendee.Entrant.Placement,
		)

		if err != nil {
			return
		}

		attendance = append(attendance, attendee)
	}

	return attendance, rows.Err()
}

func (es EntrantService) CreateEntrants(entrants []tournament.Entrant, tournamentID int64) error {
	tx, err := es.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = createEntrants(tx, entrants, tournamentID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (es EntrantService) SetPlayer(entrantID int64, playerID sql.NullInt64) error {
	query := `
UPDATE entrants
SET player_id = $2
WHERE id = $1`

	_, err := es.DB.Exec(query, entrantID, playerID)

	return err
}

// Due to the way the database is set up, deleting a Tournament will also delete its entrants, but this will still be implemented.
func (es EntrantService) DeleteEntrants(tournamentID int64) error {
	query := `
DELETE FROM entrants
WHERE tournament_id = $1`

	_, err := es.DB.Exec(query, tournamentID)

	return err
}

func createEntrants(tx *sql.Tx, entrants []tournament.Entrant, tournamentID int64) error {
	query := `
INSERT INTO entrants (name, placement, tournament_id)
VALUES ($1, $2, $3)
RETURNING id;`

	for i, entrant := range entrants {
		// This will update the entrant IDs as it goes along. If any errors occur, any written IDs will be invalidated.
		err := tx.QueryRow(query, entrant.Name, entrant.Placement, tournamentID).Scan(&entrants[i].ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func getEntrant(tx *sql.Tx, id int64) (entrant tournament.Entrant, err error) {
	if id < 1 {
		return entrant, ErrRecordNotFound
	}

	query := `
SELECT entrants.id, entrants.name, placement, tournament_id, players.name
FROM entrants
         LEFT OUTER JOIN players ON entrants.player_id = players.id
WHERE entrants.id = $1`

	err = tx.QueryRow(query, id).Scan(
		&entrant.ID,
		&entrant.Name,
		&entrant.Placement,
		&entrant.TournamentID,
		&entrant.PlayerName,
	)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		err = ErrRecordNotFound
	}

	return
}
