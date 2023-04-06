package tournament

import "database/sql"

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
