// Package startgg contains code for obtaining tournament information using the start.gg API (https://developer.start.gg/).
// This package assumes that the tournaments are already complete.
package startgg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ejacobg/tourney-tracker/tournament"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"net/http"
	"net/url"
	"strings"
)

// Obtain an API key: https://developer.start.gg/docs/authentication
// Test your queries: https://developer.start.gg/explorer/

const query = `
query TournamentEventQuery($tournament: String, $event: String) {
    tournament(slug: $tournament) {
        name
        url(relative: false)
    }
    event(slug: $event) {
        name
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
                lPlacement
            }
        }
    }
}`

// request holds the data to be sent with the API request.
type request struct {
	Query     string            `json:"query"`
	Variables map[string]string `json:"variables"`
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

// response will hold the data returned from the above query.
type response struct {
	Data struct {
		Tournament struct {
			Name string
			URL  string
		}
		Event struct {
			Name     string
			Entrants struct {
				Nodes []entrant
			}
		}
		Sets struct {
			Nodes []set
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
	LPlacement    int
}

// tournament will create a tournament.Tournament object using the data from the response.
func (r *response) tournament() tournament.Tournament {
	return tournament.Tournament{
		Name:         r.Data.Tournament.Name,
		URL:          r.Data.Tournament.URL,
		BracketReset: applyResetPoints(r.Data.Sets.Nodes),
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
func applyResetPoints(sets []set) bool {
	return slices.ContainsFunc(sets, func(s set) bool {
		// This feels wrong. Double-check the logic here.
		return s.FullRoundText == "Grand Final Reset" && s.LPlacement == 2
	})
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

	res, err := getTournament(tournamentSlug, eventSlug, key)
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

// getTournament will make a request to the start.gg API for the given tournament, returning the response data if successful, and an error otherwise.
// getTournament uses the tournament.Client value to make its request.
func getTournament(tournamentSlug, eventSlug, key string) (*response, error) {
	req, err := newRequest(tournamentSlug, eventSlug, key)
	if err != nil {
		return nil, err
	}

	res, err := tournament.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response code: %d", res.StatusCode)
	}

	var data response
	err = json.NewDecoder(res.Body).Decode(&data)

	return &data, err
}
