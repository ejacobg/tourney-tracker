// Package startgg contains code for obtaining tournament information using the start.gg API (https://developer.start.gg/).
// This package assumes that the tournaments are already complete.
package startgg

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/ejacobg/tourney-tracker/tournament"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"net/http"
	"net/url"
	"strings"
)

// Obtain an API key: https://developer.start.gg/docs/authentication
// Test your queries: https://developer.start.gg/explorer/

// This query only supports up to 500 entrants. I currently do not have plans to support more than 500 entrants.
const query = `
query TournamentEventQuery($tournament: String, $event: String) {
    tournament(slug: $tournament) {
        name
    }
    event(slug: $event) {
        name
        slug
        entrants(query: { page: 1, perPage: 500 }) {
            nodes {
                name
                standing {
                    placement
                }
            }
        }
        sets(page: 1, perPage: 3, sortType: RECENT) {
            nodes {
                fullRoundText
                winnerId
            }
        }
    }
}`

// request holds the data to be sent with the API request.
type request struct {
	Query     string            `json:"query"`
	Variables map[string]string `json:"variables"`
}

// response will hold the data returned from the above query.
type response struct {
	Data struct {
		Tournament struct {
			Name string
		}
		Event struct {
			Name     string
			Slug     string
			Entrants struct {
				Nodes []entrant
			}
			Sets struct {
				Nodes []set
			}
		}
	}
}

type entrant struct {
	Name     string
	Standing struct {
		Placement int
	}
}

type set struct {
	FullRoundText string
	WinnerID      int
}

// tournament will create a tournament.Tournament object using the data from the response.
func (r *response) tournament() tournament.Tournament {
	return tournament.Tournament{
		Name:         r.Data.Tournament.Name + " - " + r.Data.Event.Name,
		URL:          "https://start.gg/" + r.Data.Event.Slug,
		BracketReset: applyResetPoints(r.Data.Event.Sets.Nodes),
		Placements:   uniquePlacements(r.Data.Event.Entrants.Nodes),
	}
}

// entrants will return a []tournament.Entrant using the data from the response.
func (r *response) entrants() (entrants []tournament.Entrant) {
	for _, e := range r.Data.Event.Entrants.Nodes {
		entrants = append(entrants, tournament.Entrant{Name: e.Name, Placement: e.Standing.Placement})
	}
	return
}

// applyResetPoints returns true if the second-place finisher made a bracket reset.
// Bracket reset points should be applied if:
//  1. There exists a "Grand Final" and "Grand Final Reset" round.
//  2. The winners of the grand final and grand final reset are different.
func applyResetPoints(sets []set) bool {
	reset := slices.IndexFunc(sets, func(s set) bool {
		return s.FullRoundText == "Grand Final Reset"
	})
	if reset == -1 {
		return false
	}

	grands := slices.IndexFunc(sets, func(s set) bool {
		return s.FullRoundText == "Grand Final"
	})
	// If the Grand Final Reset exists, then the Grand Final should also exist (i.e. grands should always be a valid index).

	return sets[reset].WinnerID != sets[grands].WinnerID
}

// uniquePlacements returns the unique placements across all the given entrants, in reverse-sorted order.
func uniquePlacements(entrants []entrant) []int {
	// Keep track of all the placements we've seen before.
	placements := make(map[int]bool)

	// If we come across a placement we haven't seen before, add it to the map.
	for _, e := range entrants {
		if !placements[e.Standing.Placement] {
			placements[e.Standing.Placement] = true
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

// FromURL returns takes a URL to a start.gg event, calls the API with the provided API key, and returns the parsed tournament and its entrants.
// An event URL takes this form: https://start.gg/tournament/<tournament-slug>/event/<event-slug> (eg. https://start.gg/tournament/shinto-series-smash-1/event/singles-1v1)
func FromURL(URL *url.URL, key string) (tourney tournament.Tournament, entrants []tournament.Entrant, err error) {
	// Only accept start.gg (formerly smash.gg) URLs.
	if !(URL.Host == "start.gg" || URL.Host == "smash.gg") {
		return tourney, entrants, tournament.ErrUnrecognizedURL
	}

	tournamentSlug, eventSlug, err := parseSlugs(URL)
	if err != nil {
		return
	}

	req, err := newRequest(tournamentSlug, eventSlug, key)
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

// parseSlugs will extract the <tournament-slug> and <event-slug> values from the given start.gg event URL.
func parseSlugs(URL *url.URL) (tournamentSlug, eventSlug string, err error) {
	path := strings.Split(URL.Path, "/")
	if len(path) < 5 {
		return "", "", errors.New("not enough path parameters")
	}

	// An ideal path would look like this: ["", "tournament", <tournament-slug>, "event", <event-slug>]
	tournamentSlug, eventSlug = path[2], path[4]
	return
}

// newRequest returns a *http.Request populated with the data needed by the start.gg API.
// See https://developer.start.gg/docs/sending-requests for more.
func newRequest(tournamentSlug, eventSlug, key string) (*http.Request, error) {
	data := request{
		Query: query,
		Variables: map[string]string{
			"tournament": tournamentSlug,
			"event":      eventSlug,
		},
	}

	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.start.gg/gql/alpha", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	return req, nil
}
