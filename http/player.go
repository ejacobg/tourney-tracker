package http

import (
	"fmt"
	tournament "github.com/ejacobg/tourney-tracker"
	"net/http"
)

func (s *Server) registerPlayerRoutes() {
	s.router.HandlerFunc(http.MethodGet, "/", s.getRankings)
	s.router.HandlerFunc(http.MethodGet, "/players", s.getPlayers)
	s.router.HandlerFunc(http.MethodPost, "/players/new", s.postPlayer)
	s.router.HandlerFunc(http.MethodGet, "/players/:id", s.getPlayer)
	s.router.HandlerFunc(http.MethodGet, "/players/:id/name", s.getPlayerName)
	s.router.HandlerFunc(http.MethodGet, "/players/:id/name/edit", s.getPlayerNameForm)
	s.router.HandlerFunc(http.MethodPut, "/players/:id/name", s.putPlayerName)
	s.router.HandlerFunc(http.MethodDelete, "/players/:id", s.deletePlayer)
}

func (s *Server) getRankings(w http.ResponseWriter, r *http.Request) {
	ranks, err := s.PlayerService.GetRanks()
	if err != nil {
		ServerErrorResponse(w, "Failed to get ranks.")
		return
	}

	s.Render(w, 200, "index.go.html", "base", ranks)
}

// getPlayers renders a table of all saved players, as well as a form for adding a new Player.
func (s *Server) getPlayers(w http.ResponseWriter, _ *http.Request) {
	players, err := s.PlayerService.GetPlayers()
	if err != nil {
		ServerErrorResponse(w, "Failed to get players.")
		return
	}

	s.Render(w, 200, "players/index.go.html", "base", players)
}

// postPlayer accepts form data consisting of a "name" field containing the name of the Player to create.
// If successful, a redirect to the new Player object will be returned.
func (s *Server) postPlayer(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		BadRequestResponse(w, "Failed to parse form.")
		return
	}

	player := tournament.Player{Name: r.PostForm.Get("name")}

	err = s.PlayerService.CreatePlayer(&player)
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to create player: %s", err))
		return
	}

	redirect := fmt.Sprintf("/players/%d", player.ID)
	w.Header()["HX-Redirect"] = []string{redirect}
	http.Redirect(w, r, redirect, http.StatusCreated)
}

// getPlayer returns the details for the given Player, as well as all the tournaments they have attended.
func (s *Server) getPlayer(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		NotFoundResponse(w, "Invalid player ID.")
		return
	}

	player, err := s.PlayerService.GetPlayer(id)
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to get player: %s", err))
		return
	}

	attendance, err := s.EntrantService.GetAttendance(id)
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to get player attendance: %s", err))
		return
	}

	s.Render(w, 200, "players/view.go.html", "base", map[string]any{
		"Player":     player,
		"Attendance": attendance,
	})
}

// getPlayerName will respond with an element displaying the Player's name, alongside a button to go to the name editing form.
func (s *Server) getPlayerName(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		NotFoundResponse(w, "Invalid player ID.")
		return
	}

	player, err := s.PlayerService.GetPlayer(id)
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to get player: %s", err))
		return
	}

	s.Render(w, 200, "players/view.go.html", "name", player)
}

// getPlayerNameForm will respond with a form element that allows for changing of a Player's name.
func (s *Server) getPlayerNameForm(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		NotFoundResponse(w, "Invalid player ID.")
		return
	}

	player, err := s.PlayerService.GetPlayer(id)
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to get player: %s", err))
		return
	}

	s.Render(w, 200, "players/edit.go.html", "name", player)
}

// putPlayerName accepts form data consisting of a "name" field containing the value of the new name.
// The new name will then be applied to the Player, and a refresh of the Player page will be returned.
func (s *Server) putPlayerName(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		NotFoundResponse(w, "Invalid player ID.")
		return
	}

	err = r.ParseForm()
	if err != nil {
		BadRequestResponse(w, "Failed to parse form.")
		return
	}

	player := tournament.Player{
		ID:   id,
		Name: r.PostForm.Get("name"),
	}

	err = s.PlayerService.UpdatePlayer(&player)
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to update player: %s", err))
		return
	}

	// Refresh is needed to update the title and header of the page.
	w.Header()["HX-Refresh"] = []string{"true"}
	w.WriteHeader(http.StatusOK)
}

// deletePlayer deletes the given Player and updates all of its Entrant records.
func (s *Server) deletePlayer(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		NotFoundResponse(w, "Invalid player ID.")
		return
	}

	err = s.PlayerService.DeletePlayer(id)
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to delete player: %s", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
