package http

import (
	"fmt"
	"net/http"
)

func (s *Server) registerEntrantRoutes() {
	s.router.HandlerFunc(http.MethodGet, "/entrants/:id/player/edit", s.getEntrantPlayerForm)
}

// getEntrantPlayerForm will respond with a table row that allows for changing of an Entrant's Player.
func (s *Server) getEntrantPlayerForm(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		NotFoundResponse(w, "Invalid entrant ID.")
		return
	}

	entrant, points, err := s.EntrantService.GetEntrantWithPoints(id)
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to get entrant and points: %s", err))
		return
	}

	players, err := s.PlayerService.GetPlayers()
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to get players: %s", err))
		return
	}

	s.Render(w, 200, "entrants/edit.go.html", "player", map[string]any{
		"Entrant": entrant,
		"Points":  points,
		"Players": players,
	})
}
