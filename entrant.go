package tourney_tracker

import "database/sql"

// Entrant represents a participant in a Tournament. Entrants may or may not represent a Player.
type Entrant struct {
	ID           int64          `json:"id"`
	Name         string         `json:"name"`
	Placement    int64          `json:"placement"`
	TournamentID int64          `json:"tournamentID"`
	PlayerName   sql.NullString `json:"playerName"` // Storing the name rather than the player ID.
}

// EntrantService represents a service for managing entrants.
type EntrantService interface {
	// GetEntrants returns all entrants for a given Tournament.
	GetEntrants(tournamentID int64) ([]Entrant, error)

	// GetAttendance returns all attendance records for a given Player.
	GetAttendance(playerID int64) ([]Attendee, error)

	// CreateEntrants adds all the given entrants to the given tournament.
	// Entrants are typically parsed in bulk by the program, so it makes sense to just add them all at once.
	CreateEntrants(entrants []Entrant, tournamentID int64) error

	// SetPlayer updates the Player of the given Tier.
	SetPlayer(entrantID int64, playerID sql.NullInt64) error

	// DeleteEntrants deletes all entrants for the given Tournament.
	DeleteEntrants(tournamentID int64) error
}

// Attendee represents a player's participation in a Tournament.
type Attendee struct {
	Tournament Preview
	Entrant    struct {
		Name      string
		Placement int64
	}
}
