package postgres

import (
	"database/sql"
	"errors"
	tournament "github.com/ejacobg/tourney-tracker"
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
