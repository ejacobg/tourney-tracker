// Package tournament contains code that deals with Tournament objects, including handlers.
package tournament

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var ErrUnrecognizedURL = errors.New("unrecognized url")

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

var Client = &http.Client{
	Timeout: 10 * time.Second,
}

type Tournament struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
	// BracketReset is true if any bracket reset points should be applied to the second-place entrant.
	BracketReset bool `json:"bracketReset"`
	// Placements contains the unique placements of a tournament, in reverse-sorted order.
	// For example, if the final standings for an 8-man tournament are [7, 7, 5, 5, 4, 3, 2, 1], then the unique placements are [7, 5, 4, 3, 2, 1].
	Placements []int64 `json:"placements"` // Can't scan into the normal int type, use int64 or sql.NullInt64. (https://stackoverflow.com/questions/47962615/query-for-an-integer-array-from-postresql-always-returns-uint8)
	Tier
}

type Tier struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Multiplier int    `json:"multiplier"`
}

type Entrant struct {
	ID           int64          `json:"id"`
	Name         string         `json:"name"`
	Placement    int64          `json:"placement"`
	TournamentID int64          `json:"tournamentID"`
	PlayerName   sql.NullString `json:"playerName"` // Storing the name rather than the ID.
	// 	Provide an extra field for the player object (if it exists?)
}

type Player struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

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

// Get will use the Client to send the given request. It will then attempt to fill the Response type using the data in the response body.
// Alternatively, the Response can be an interface with the Tournament() and Entrants() methods.
func Get[Response any](req *http.Request) (*Response, error) {
	res, err := Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response code: %d", res.StatusCode)
	}

	var data Response
	err = json.NewDecoder(res.Body).Decode(&data)

	return &data, err
}
