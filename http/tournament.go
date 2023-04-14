package http

import (
	"errors"
	"fmt"
	tournament "github.com/ejacobg/tourney-tracker"
	"github.com/ejacobg/tourney-tracker/convert"
	"github.com/ejacobg/tourney-tracker/convert/challonge"
	"github.com/ejacobg/tourney-tracker/convert/startgg"
	"net/http"
	"net/url"
	"strconv"
)

func (s *Server) registerTournamentRoutes() {
	s.router.HandlerFunc(http.MethodGet, "/tournaments", s.getTournaments)
	s.router.HandlerFunc(http.MethodPost, "/tournaments/new", s.postTournamentURL)
	s.router.HandlerFunc(http.MethodGet, "/tournaments/:id", s.getTournament)
	s.router.HandlerFunc(http.MethodGet, "/tournaments/:id/tier", s.getTournamentTier)
	s.router.HandlerFunc(http.MethodGet, "/tournaments/:id/tier/edit", s.getTournamentTierForm)
	s.router.HandlerFunc(http.MethodPut, "/tournaments/:id/tier", s.putTournamentTier)
}

// getTournaments renders a table of all saved tournaments, as well as a form for adding a new Tournament.
func (s *Server) getTournaments(w http.ResponseWriter, _ *http.Request) {
	previews, err := s.TournamentService.GetPreviews()
	if err != nil {
		ServerErrorResponse(w, "Failed to get previews.")
		return
	}

	s.Render(w, 200, "tournaments/index.go.html", "base", previews)
}

// postTournamentURL accepts form data consisting of a "url" field containing a URL to a tournament.
// If an error occurs while processing the URL, an error message will be returned.
// Otherwise, a redirect to the new Tournament object will be returned.
func (s *Server) postTournamentURL(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		BadRequestResponse(w, "Failed to parse form.")
		return
	}

	URL, err := url.Parse(r.PostForm.Get("url"))
	if err != nil {
		UnprocessableEntityResponse(w, "Failed to parse URL.")
		return
	}

	var (
		tourney  tournament.Tournament
		entrants []tournament.Entrant
	)

	switch URL.Host {
	case "challonge.com":
		tourney, entrants, err = challonge.FromURL(URL, s.challongeUsername, s.challongePassword)
	case "start.gg", "smash.gg":
		tourney, entrants, err = startgg.FromURL(URL, s.startggKey)
	default:
		err = convert.ErrUnrecognizedURL
	}

	if err != nil {
		switch {
		case errors.Is(err, convert.ErrUnrecognizedURL):
			UnprocessableEntityResponse(w, fmt.Sprintf("Unrecognized host: %q", URL.Host))
		default:
			UnprocessableEntityResponse(w, fmt.Sprintf("Parsing error: %s", err))
		}
		return
	}

	err = s.TournamentService.CreateTournament(&tourney, entrants)
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to create tournament: %s", err))
		return
	}

	redirect := fmt.Sprintf("/tournaments/%d", tourney.ID)
	w.Header()["HX-Redirect"] = []string{redirect}
	http.Redirect(w, r, redirect, http.StatusCreated)
}

// getTournament will read the "id" route parameter and display the details for the given tournament.
func (s *Server) getTournament(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	tourney, err := s.TournamentService.GetTournament(id)
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to get tournament: %s", err))
		return
	}

	entrants, err := s.EntrantService.GetEntrants(id)
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to get entrants: %s", err))
		return
	}

	points := tournament.NewPointMap(tourney.BracketReset, tourney.Placements, tourney.Tier.Multiplier)

	s.Render(w, 200, "tournaments/view.go.html", "base", map[string]any{
		"Tourney":  tourney,
		"Entrants": entrants,
		"Points":   points,
	})
}

// getTournamentTier will respond with an element displaying the Tournament's Tier, alongside a button to go to the Tier editing form.
func (s *Server) getTournamentTier(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	tier, err := s.TierService.GetTournamentTier(id)
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to get tier: %s", err))
		return
	}

	s.Render(w, 200, "tournaments/view.go.html", "tier", map[string]any{
		"TournamentID": id,
		"Tier":         tier,
	})
}

// getTournamentTierForm will respond with a form element that allows for changing of a Tournament's Tier.
func (s *Server) getTournamentTierForm(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	tiers, err := s.TierService.GetTiers()
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to get tiers: %s", err))
		return
	}

	s.Render(w, 200, "tournaments/edit.go.html", "tier", map[string]any{
		"TournamentID": id,
		"Tiers":        tiers,
	})
}

// putTournamentTier accepts form data consisting of a "tier" field containing the value of the new Tier.
// The new Tier will then be applied to the Tournament, and a refresh of the Tournament page will be returned.
func (s *Server) putTournamentTier(w http.ResponseWriter, r *http.Request) {
	// Get Tournament ID.
	tournamentID, err := readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Get Tier ID.
	err = r.ParseForm()
	if err != nil {
		BadRequestResponse(w, "Failed to parse form.")
		return
	}

	tierID, err := strconv.ParseInt(r.PostForm.Get("tier"), 10, 64)
	if err != nil {
		UnprocessableEntityResponse(w, fmt.Sprintf("Invalid tier ID."))
		return
	}

	// Apply new Tier to Tournament.
	err = s.TournamentService.SetTier(tournamentID, tierID)
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to update tier: %s", err))
		return
	}

	// Refresh the page.
	w.Header()["HX-Refresh"] = []string{"true"}
	w.WriteHeader(http.StatusOK)
}
