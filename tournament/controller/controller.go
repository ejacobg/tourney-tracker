package controller

import (
	"bytes"
	"fmt"
	"github.com/ejacobg/tourney-tracker/convert/challonge"
	"github.com/ejacobg/tourney-tracker/convert/startgg"
	"github.com/ejacobg/tourney-tracker/tournament"
	"html/template"
	"net/http"
	"net/url"
)

// Controller provides several HTTP handlers for servicing tournament-related requests.
type Controller struct {
	Model tournament.Model
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
func (c *Controller) Index(w http.ResponseWriter, r *http.Request) {
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

// New accepts form data consisting of a "url" field containing a URL to a tournament.
// If an error occurs while processing the URL, an error message will be returned.
// Otherwise, a redirect to the new tournament.Tournament object will be returned.
func (c *Controller) New(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form.", http.StatusBadRequest)
		return
	}

	URL, err := url.Parse(r.PostForm.Get("url"))
	if err != nil {
		http.Error(w, "Failed to parse form.", http.StatusUnprocessableEntity)
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
		err = tournament.ErrUnrecognizedURL
	}

	if err != nil {
		http.Error(w, "Unrecognized host: "+URL.Host, http.StatusUnprocessableEntity)
		return
	}

	err = c.Model.Insert(&tourney, entrants)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create tournament: %s", err), http.StatusInternalServerError)
		return
	}

	redirect := fmt.Sprintf("/tournaments/%d", tourney.ID)
	w.Header()["HX-Redirect"] = []string{redirect}
	http.Redirect(w, r, redirect, http.StatusCreated)
}
