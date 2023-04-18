// Package tourney_tracker contains core types that deal with handling Tournament objects.
package tourney_tracker

// Tournament holds fields relevant to the point calculation. A tournament is generally considered immutable after creation, except for its Tier.
// It is assumed that the original tournament has already been completed. In-progress tournaments may not be parsed correctly.
// It is assumed that the tournament is double-elimination, however the point formula may still work for other bracket types.
type Tournament struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`

	// BracketReset is true if any bracket reset points should be applied to the second-place entrant.
	BracketReset bool `json:"bracketReset"`

	// Placements contains the unique placements of a tournament, in reverse-sorted order.
	// For example, if the final standings for an 8-man tournament are [7, 7, 5, 5, 4, 3, 2, 1], then the unique placements are [7, 5, 4, 3, 2, 1].
	Placements []int64 `json:"placements"` // Can't scan into the normal int type, use int64 or sql.NullInt64. (https://stackoverflow.com/questions/47962615/query-for-an-integer-array-from-postresql-always-returns-uint8)

	Tier Tier
}

// TournamentService represents a service for managing tournaments.
type TournamentService interface {
	// GetPreviews returns previews for all tournaments.
	GetPreviews() ([]Preview, error)

	// GetNamesByTier returns the names of all tournaments with the given tier.
	GetNamesByTier(tierID int64) ([]Name, error)

	// GetTournament returns a single Tournament by ID.
	GetTournament(id int64) (Tournament, error)

	// CreateTournament adds the given Tournament and its entrants to the database.
	// The Tournament and entrants should be created in the same transaction.
	CreateTournament(tourney *Tournament, entrants []Entrant) error

	// SetTier updates the Tier of the given Tournament.
	SetTier(tournamentID, tierID int64) error

	// DeleteTournament deletes a Tournament.
	DeleteTournament(id int64) error
}

// Preview represents a subset of a Tournament object, namely its ID, name, and Tier.
type Preview struct {
	ID   int64
	Name string
	Tier string
}

// Name represents a unique Tournament name.
type Name struct {
	ID   int64
	Name string
}
