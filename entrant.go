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
}

// Attendee represents a player's participation in a Tournament.
type Attendee struct {
	Tournament Preview
	Entrant    struct {
		Name      string
		Placement int64
	}
}
