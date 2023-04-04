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
	Placements []int `json:"placements"`
	Tier
}

type Tier struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Multiplier int    `json:"multiplier"`
}

type Entrant struct {
	ID           int64         `json:"id"`
	Name         string        `json:"name"`
	Placement    int           `json:"placement"`
	TournamentID int64         `json:"tournamentID"`
	PlayerID     sql.NullInt64 `json:"playerID"`
	// 	Provide an extra field for the player object (if it exists?)
}

type Player struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
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
