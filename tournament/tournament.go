// Package tournament contains code that deals with Tournament objects, including handlers.
package tournament

type Tournament struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
	// BracketReset is true if any bracket reset points should be applied to the second-place entrant.
	BracketReset bool `json:"bracketReset"`
	// Placements represents the number of unique placements in a tournament.
	// For example, if the final standings for an 8-man tournament are [7, 7, 5, 5, 4, 3, 2, 1] then there are 6 unique placements.
	Placements int `json:"placements"`
	Tier
}

type Tier struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Multiplier int    `json:"multiplier"`
}
