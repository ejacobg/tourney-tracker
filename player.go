package tourney_tracker

// Player represents a person whose tournament record we wish to track.
type Player struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// PlayerService represents a service for managing players.
type PlayerService interface {
	// GetPlayers returns all players.
	GetPlayers() ([]Player, error)

	// GetPlayer returns a single Player by ID.
	GetPlayer(id int64) (Player, error)

	// CreatePlayer adds the given Player to the database.
	CreatePlayer(player *Player) error

	// UpdatePlayer updates the given Player.
	UpdatePlayer(player *Player) error

	// DeletePlayer deletes the given Player.
	// Deleting a Player should nullify any entrants pointing to it.
	DeletePlayer(id int64) error
}
