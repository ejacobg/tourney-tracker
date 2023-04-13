package http

import (
	"bytes"
	"fmt"
	tournament "github.com/ejacobg/tourney-tracker"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"net/http"
	"time"
)

// Server provides several HTTP handlers for servicing http-related requests.
type Server struct {
	router *httprouter.Router

	// Addr is the address for the Server to listen on.
	Addr string

	// Templates holds all the templates used by the application.
	Templates map[string]*template.Template

	// Credentials needed for API calls.
	challongeUsername, challongePassword string
	startggKey                           string

	// Services used by the various HTTP routes.
	EntrantService    tournament.EntrantService
	PlayerService     tournament.PlayerService
	TierService       tournament.TierService
	TournamentService tournament.TournamentService
}

// NewServer creates a Server with the given credentials. The other fields should be applied manually.
func NewServer(challongeUsername, challongePassword, startggKey string) *Server {
	srv := Server{
		router:            httprouter.New(),
		challongeUsername: challongeUsername,
		challongePassword: challongePassword,
		startggKey:        startggKey,
	}

	srv.router.Handler("GET", "/static/*filepath", http.FileServer(http.Dir("ui")))

	srv.registerTournamentRoutes()

	return &srv
}

// Render will execute the "name" template of "tmpl", then write it to the response with the given status code.
func (s *Server) Render(w http.ResponseWriter, status int, tmpl, name string, data any) {
	t, ok := s.Templates[tmpl]
	if !ok {
		ServerErrorResponse(w, fmt.Sprintf("The template %q does not exist.", tmpl))
		return
	}

	buf := new(bytes.Buffer)

	err := t.ExecuteTemplate(buf, name, data)
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to render template: %s", err))
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (s *Server) ListenAndServe() error {
	srv := http.Server{
		Addr:         s.Addr,
		Handler:      s.router,
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return srv.ListenAndServe()
}
