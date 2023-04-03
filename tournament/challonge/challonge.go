// Package challonge contains code for obtaining tournament information using the Challonge API (https://api.challonge.com/v1).
// This package assumes that the tournaments are already complete.
package challonge

import (
	"errors"
	"github.com/ejacobg/tourney-tracker/tournament"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"net/http"
	"net/url"
	"strings"
)

// Note that this uses the v1 API. I had some trouble getting the v2 API to work.
const baseURL = "https://api.challonge.com/v1/tournaments/"

type response struct {
	Tournament struct {
		Name         string        `json:"name"`
		URL          string        `json:"full_challonge_url"`
		Participants []participant `json:"participants"`
		Matches      []match       `json:"matches"`
	}
}

// tournament will create a tournament.Tournament object using the data from the response.
func (r *response) tournament() tournament.Tournament {
	return tournament.Tournament{
		Name:         r.Tournament.Name,
		URL:          r.Tournament.URL,
		BracketReset: applyResetPoints(r.Tournament.Matches),
		Placements:   uniquePlacements(r.Tournament.Participants),
	}
}

// entrants will return a []tournament.Entrant using the data from the response.
func (r *response) entrants() (entrants []tournament.Entrant) {
	for _, p := range r.Tournament.Participants {
		entrants = append(entrants, tournament.Entrant{Name: p.Participant.Name, Placement: p.Participant.FinalRank})
	}
	return
}

// applyResetPoints returns true if the second-place finisher made a bracket reset.
// Bracket reset points should be applied if:
//  1. The last two matches occurred between first and second place.
//     If this is true, then the last two matches must be the grand final and the grand final reset.
//     Otherwise, the last two matches would be the loser's final and the grand final.
//  2. The winners of the last two matches are different.
//     If the winner of the grand final and the winner of the grand final reset are different, then that means the second-place finisher made a reset, but did not win.
//     Note that this condition is true if the last two matches are the loser's final and the grand final. However, this case is handled by (1).
func applyResetPoints(matches []match) bool {
	// The last match is either the grand final or the grand final reset.
	// In either case, the first and second place finalists are both present in this match.
	last := slices.IndexFunc(matches, func(m match) bool {
		return m.Match.Order == len(matches)
	})
	first := matches[last].Match.WinnerID
	second := matches[last].Match.LoserID

	// The previous match is either the loser's final or the grand final.
	prev := slices.IndexFunc(matches, func(m match) bool {
		return m.Match.Order == len(matches)-1
	})

	// 1. The last two matches occurred between first and second place.
	if !in(first, matches[prev].Match.Player1ID, matches[prev].Match.Player2ID) {
		return false
	}
	if !in(second, matches[prev].Match.Player1ID, matches[prev].Match.Player2ID) {
		return false
	}

	// 2. The winners of the last two matches are different.
	return matches[last].Match.WinnerID != matches[prev].Match.WinnerID
}

func in(id int, ids ...int) bool {
	return slices.Contains(ids, id)
}

// uniquePlacements returns the unique placements across all the given entrants, in reverse-sorted order.
func uniquePlacements(participants []participant) []int {
	// Keep track of all the placements we've seen before.
	placements := make(map[int]bool)

	// If we come across a placement we haven't seen before, add it to the map.
	for _, p := range participants {
		if !placements[p.Participant.FinalRank] {
			placements[p.Participant.FinalRank] = true
		}
	}

	// The keys of the map will be all the unique placements.
	keys := maps.Keys(placements)

	// Sort our keys in descending order.
	slices.SortFunc(keys, func(a, b int) bool {
		return a > b
	})

	return keys
}

type match struct {
	Match struct {
		Player1ID int `json:"player1_id"`
		Player2ID int `json:"player2_id"`
		WinnerID  int `json:"winner_id"`
		LoserID   int `json:"loser_id"`
		Order     int `json:"suggested_play_order"`
	}
}

type participant struct {
	Participant struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		FinalRank int    `json:"final_rank"`
	}
}

// FromURL takes a URL to a Challonge tournament, calls the API with the provided credentials, and returns the parsed tournament and its entrants.
// A tournament URL takes the form: https://challonge.com/<tournament-id> (eg. https://challonge.com/8ozc6ffz)
func FromURL(URL *url.URL, username, password string) (tourney tournament.Tournament, entrants []tournament.Entrant, err error) {
	// Only accept challonge.com URLs.
	if URL.Host != "challonge.com" {
		return tourney, entrants, tournament.ErrUnrecognizedURL
	}

	tournamentID, err := parseID(URL)
	if err != nil {
		return
	}

	req, err := newRequest(tournamentID, username, password)
	if err != nil {
		return
	}

	res, err := tournament.Get[response](req)
	if err != nil {
		return
	}

	tourney = res.tournament()
	entrants = res.entrants()
	return
}

// parseID will extract the <tournament-id> value from the given Challonge URL.
func parseID(URL *url.URL) (tournamentID string, err error) {
	path := strings.Split(URL.Path, "/")
	if len(path) < 2 {
		return "", errors.New("not enough path parameters")
	}

	// An ideal path would look like this: ["", "<tournamentID>"]
	tournamentID = path[1]
	return
}

// newRequest returns a *http.Request for the given tournament using the given basic authentication credentials.
func newRequest(tournamentID, username, password string) (*http.Request, error) {
	// Request URLs take the form: https://api.challonge.com/v1/tournaments/<tournamentID>.json
	req, err := http.NewRequest(http.MethodGet, baseURL+tournamentID+".json", nil)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Add("include_participants", "1")
	query.Add("include_matches", "1")
	req.URL.RawQuery = query.Encode()

	req.SetBasicAuth(username, password)

	return req, nil
}
