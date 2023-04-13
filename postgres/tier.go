package postgres

import (
	"database/sql"
	"errors"
	tournament "github.com/ejacobg/tourney-tracker"
)

// TierService represents a service for managing tiers.
type TierService struct {
	DB *sql.DB
}

func (ts TierService) GetTiers() (tiers []tournament.Tier, err error) {
	query := `
SELECT id, name, multiplier
FROM tiers`

	rows, err := ts.DB.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()

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

func (ts TierService) GetTier(id int64) (tier tournament.Tier, err error) {
	query := `
SELECT id, name, multiplier
FROM tiers
WHERE id = $1`

	err = ts.DB.QueryRow(query, id).Scan(&tier.ID, &tier.Name, &tier.Multiplier)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		err = ErrRecordNotFound
	}

	return
}

func (ts TierService) GetTournamentTier(tournamentID int64) (tier tournament.Tier, _ error) {
	query := `
SELECT tiers.id, tiers.name, multiplier
FROM tournaments
INNER JOIN tiers on tournaments.tier_id = tiers.id
WHERE tournaments.id = $1;`

	return tier, ts.DB.QueryRow(query, tournamentID).Scan(&tier.ID, &tier.Name, &tier.Multiplier)
}

func (ts TierService) CreateTier(tier *tournament.Tier) error {
	query := `
INSERT INTO tiers (name, multiplier)
VALUES ($1, $2)
RETURNING id`

	err := ts.DB.QueryRow(query, tier.Name, tier.Multiplier).Scan(&tier.ID)

	return err
}

func (ts TierService) UpdateTier(tier *tournament.Tier) error {
	query := `UPDATE tiers
SET name = $2, multiplier = $3
WHERE id = $1`

	_, err := ts.DB.Exec(query, tier.ID, tier.Name, tier.Multiplier)

	return err
}

func (ts TierService) DeleteTier(id int64) error {
	query := `
DELETE FROM tiers
WHERE id = $1`

	_, err := ts.DB.Exec(query, id)

	return err
}
