package tourney_tracker

// Variables used in the point formula. These are not stored in the database.
const (
	// UP is the points given for each unique placement.
	UP = 5

	// ATT is the points given for showing up to a tournament.
	ATT = 10

	// FIRST is the points given for winning a tournament.
	FIRST = 10

	// BR is the points given to the second-place finisher if they made a bracket reset.
	BR = 5
)

// NewPointMap returns a mapping from each placement to the number of points it is worth.
// The point formula is as follows: (UP * PV + ATT + FIRST? + BR?) * TIER
func NewPointMap(bracketReset bool, placements []int64, multiplier int) map[int64]int {
	pm := make(map[int64]int)
	for i, placement := range placements {
		points := UP*i + ATT
		if placement == 1 {
			points += FIRST
		} else if placement == 2 && bracketReset {
			points += BR
		}

		pm[placement] = points * multiplier
	}
	return pm
}
