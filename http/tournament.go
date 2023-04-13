package http

import (
	"errors"
	"fmt"
	tournament "github.com/ejacobg/tourney-tracker"
	"github.com/ejacobg/tourney-tracker/convert"
	"github.com/ejacobg/tourney-tracker/convert/challonge"
	"github.com/ejacobg/tourney-tracker/convert/startgg"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/url"
	"strconv"
)

func (s *Server) registerTournamentRoutes() {
	s.router.HandlerFunc("GET", "/", s.Index)
	s.router.HandlerFunc("POST", "/tournaments/new", s.New)
	s.router.HandlerFunc("GET", "/tournaments/:id", s.View)
	s.router.HandlerFunc("GET", "/tournaments/:id/tier", s.ViewTier)
	s.router.HandlerFunc("GET", "/tournaments/:id/tier/edit", s.EditTier)
}

// Index renders a table of all saved tournaments, as well as a form for adding a new one.
func (s *Server) Index(w http.ResponseWriter, _ *http.Request) {
	previews, err := s.TournamentService.GetPreviews()
	if err != nil {
		ServerError(w, fmt.Errorf("failed to retrieve previews"))
		return
	}

	s.Render(w, 200, "tournaments/index.go.html", "base", previews)
}

// NewServer accepts form data consisting of a "url" field containing a URL to a http.
// If an error occurs while processing the URL, an error message will be returned.
// Otherwise, a redirect to the new http.Tournament object will be returned.
func (s *Server) New(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form.", http.StatusBadRequest)
		return
	}

	URL, err := url.Parse(r.PostForm.Get("url"))
	if err != nil {
		http.Error(w, "Failed to parse URL.", http.StatusUnprocessableEntity)
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
			http.Error(w, fmt.Sprintf("Unrecognized host: %q", URL.Host), http.StatusUnprocessableEntity)
		default:
			http.Error(w, fmt.Sprintf("Parsing error: %s", err), http.StatusUnprocessableEntity)
		}
		return
	}

	err = s.TournamentService.CreateTournament(&tourney, entrants)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create http: %s", err), http.StatusInternalServerError)
		return
	}

	redirect := fmt.Sprintf("/tournaments/%d", tourney.ID)
	w.Header()["HX-Redirect"] = []string{redirect}
	http.Redirect(w, r, redirect, http.StatusCreated)
}

// View will read the "id" route parameter and display the details for the given http.
func (s *Server) View(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	tourney, err := s.TournamentService.GetTournament(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve tournament: %s", err), http.StatusInternalServerError)
		return
	}

	entrants, err := s.EntrantService.GetEntrants(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve entrants: %s", err), http.StatusInternalServerError)
		return
	}

	points := tournament.NewPointMap(tourney.BracketReset, tourney.Placements, tourney.Tier.Multiplier)

	s.Render(w, 200, "tournaments/view.go.html", "base", map[string]any{
		"Tourney":  tourney,
		"Entrants": entrants,
		"Points":   points,
	})
}

// ViewTier will respond with an element displaying the http's tier, alongside a button to go to the tier editing form.
func (s *Server) ViewTier(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	tier, err := s.TierService.GetTournamentTier(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get tier: %s", err), http.StatusInternalServerError)
		return
	}

	s.Render(w, 200, "tournaments/view.go.html", "tier", map[string]any{
		"TournamentID": id,
		"Tier":         tier,
	})
}

// EditTier will respond with a form element that allows for changing of a http's tier.
func (s *Server) EditTier(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	tiers, err := s.TierService.GetTiers()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get tiers: %s", err), http.StatusInternalServerError)
		return
	}

	s.Render(w, 200, "tournaments/edit.go.html", "tier", map[string]any{
		"TournamentID": id,
		"Tiers":        tiers,
	})
}
