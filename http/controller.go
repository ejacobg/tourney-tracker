package tournament

import (
	"bytes"
	"errors"
	"fmt"
	tournament "github.com/ejacobg/tourney-tracker"
	"github.com/ejacobg/tourney-tracker/convert"
	"github.com/ejacobg/tourney-tracker/convert/challonge"
	"github.com/ejacobg/tourney-tracker/convert/startgg"
	"github.com/ejacobg/tourney-tracker/postgres"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
)

// Controller provides several HTTP handlers for servicing http-related requests.
type Controller struct {
	Model postgres.Model
	Views struct {
		Index, View, Edit *template.Template
	}
	// Credentials needed for API calls.
	challongeUsername, challongePassword string
	startggKey                           string
}

// New creates a Controller with the given credentials. The other fields should be applied manually.
func New(challongeUsername, challongePassword, startggKey string) *Controller {
	return &Controller{
		challongeUsername: challongeUsername,
		challongePassword: challongePassword,
		startggKey:        startggKey,
	}
}

// Index renders a table of all saved tournaments, as well as a form for adding a new one.
func (c *Controller) Index(w http.ResponseWriter, _ *http.Request) {
	previews, err := c.Model.GetPreviews()
	if err != nil {
		http.Error(w, "Failed to retrieve previews.", http.StatusInternalServerError)
		return
	}

	if c.Views.Index == nil {
		http.Error(w, "Index page does not exist.", http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)

	err = c.Views.Index.ExecuteTemplate(buf, "base", previews)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	buf.WriteTo(w)
}

// New accepts form data consisting of a "url" field containing a URL to a http.
// If an error occurs while processing the URL, an error message will be returned.
// Otherwise, a redirect to the new http.Tournament object will be returned.
func (c *Controller) New(w http.ResponseWriter, r *http.Request) {
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
		tourney, entrants, err = challonge.FromURL(URL, c.challongeUsername, c.challongePassword)
	case "start.gg", "smash.gg":
		tourney, entrants, err = startgg.FromURL(URL, c.startggKey)
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

	err = c.Model.Insert(&tourney, entrants)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create http: %s", err), http.StatusInternalServerError)
		return
	}

	redirect := fmt.Sprintf("/tournaments/%d", tourney.ID)
	w.Header()["HX-Redirect"] = []string{redirect}
	http.Redirect(w, r, redirect, http.StatusCreated)
}

// View will read the "id" route parameter and display the details for the given http.
func (c *Controller) View(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	tourney, entrants, err := c.Model.Get(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve http: %s", err), http.StatusInternalServerError)
		return
	}

	points := tournament.NewPointMap(tourney.BracketReset, tourney.Placements, tourney.Tier.Multiplier)

	if c.Views.View == nil {
		http.Error(w, "View page does not exist.", http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)

	err = c.Views.View.ExecuteTemplate(buf, "base", map[string]any{
		"Tourney":  tourney,
		"Entrants": entrants,
		"Points":   points,
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	buf.WriteTo(w)
}

// ViewTier will respond with an element displaying the http's tier, alongside a button to go to the tier editing form.
func (c *Controller) ViewTier(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	tier, err := c.Model.GetTier(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get tier: %s", err), http.StatusInternalServerError)
	}

	if c.Views.View == nil {
		http.Error(w, "View tier template does not exist.", http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)

	err = c.Views.View.ExecuteTemplate(buf, "tier", map[string]any{
		"TournamentID": id,
		"Tier":         tier,
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	buf.WriteTo(w)
}

// EditTier will respond with a form element that allows for changing of a http's tier.
func (c *Controller) EditTier(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	tiers, err := c.Model.GetTiers()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get tiers: %s", err), http.StatusInternalServerError)
		return
	}

	if c.Views.Edit == nil {
		http.Error(w, "Edit tier template does not exist.", http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)

	err = c.Views.Edit.ExecuteTemplate(buf, "tier", map[string]any{
		"TournamentID": id,
		"Tiers":        tiers,
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	buf.WriteTo(w)
}
