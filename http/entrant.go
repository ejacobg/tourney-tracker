package http

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
)

func (s *Server) registerEntrantRoutes() {
	s.router.HandlerFunc(http.MethodGet, "/entrants/:id/player", s.getEntrantPlayer)
	s.router.HandlerFunc(http.MethodGet, "/entrants/:id/player/edit", s.getEntrantPlayerForm)
	s.router.HandlerFunc(http.MethodPut, "/entrants/:id/player", s.putEntrantPlayer)
}

// getEntrantPlayer returns a table row showing the Entrant's name, Player name, points earned, and their placement.
func (s *Server) getEntrantPlayer(w http.ResponseWriter, r *http.Request) {
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

	s.Render(w, 200, "entrants/view.go.html", "player", map[string]any{
		"Entrant": entrant,
		"Points":  points,
	})
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

// putEntrantPlayer accepts form data consisting of a "player" field containing the value of the new Player ID.
// The new Player will then be applied to the Entrant, and an updated table row element will be returned.
func (s *Server) putEntrantPlayer(w http.ResponseWriter, r *http.Request) {
	// Get Entrant ID.
	entrantID, err := readIDParam(r)
	if err != nil {
		NotFoundResponse(w, "Invalid entrant ID.")
		return
	}

	// Get Player ID.
	err = r.ParseForm()
	if err != nil {
		BadRequestResponse(w, "Failed to parse form.")
		return
	}

	var playerID sql.NullInt64
	playerID.Int64, err = strconv.ParseInt(r.PostForm.Get("player"), 10, 64)
	playerID.Valid = true

	// If the "player" field could not be parsed, then it will be treated as a null value.
	if err != nil {
		playerID.Valid = false
	}

	// Apply new Player to Entrant.
	err = s.EntrantService.SetPlayer(entrantID, playerID)
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to update player: %s", err))
		return
	}

	// Render updated row.
	entrant, points, err := s.EntrantService.GetEntrantWithPoints(entrantID)
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to get entrant and points: %s", err))
		return
	}

	s.Render(w, 200, "entrants/view.go.html", "player", map[string]any{
		"Entrant": entrant,
		"Points":  points,
	})
}
