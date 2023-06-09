package tourney_tracker

// Tier represents the relative importance of a Tournament.
type Tier struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Multiplier int    `json:"multiplier"`
}

// TierService represents a service for managing tiers.
type TierService interface {
	// GetTiers returns all tiers.
	GetTiers() ([]Tier, error)

	// GetTier returns a single Tier by ID.
	GetTier(id int64) (Tier, error)

	// GetTournamentTier returns the Tier for the given Tournament.
	GetTournamentTier(tournamentID int64) (Tier, error)

	// CreateTier adds the given Tier to the database.
	CreateTier(tier *Tier) error

	// UpdateTier updates the given Tier.
	UpdateTier(tier *Tier) error

	// DeleteTier deletes the given Tier.
	// Note that deleting a tier that still has tournaments attached to it should fail.
	// It is up to the user to ensure that all tournaments update their Tier before attempting to delete.
	DeleteTier(id int64) error
}
