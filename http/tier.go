package http

import (
	"fmt"
	"net/http"
)

func (s *Server) registerTierRoutes() {
	s.router.HandlerFunc(http.MethodGet, "/tiers", s.getTiers)
	s.router.HandlerFunc(http.MethodGet, "/tiers/:id", s.getTier)
}

// getTiers renders all the current tiers. Right now, the current tiers are considered immutable.
func (s *Server) getTiers(w http.ResponseWriter, _ *http.Request) {
	tiers, err := s.TierService.GetTiers()
	if err != nil {
		ServerErrorResponse(w, "Failed to get tiers.")
		return
	}

	s.Render(w, 200, "tiers/index.go.html", "base", tiers)
}

// getTier renders the names of all the tournaments with the given Tier.
func (s *Server) getTier(w http.ResponseWriter, r *http.Request) {
	id, err := readIDParam(r)
	if err != nil {
		NotFoundResponse(w, "Invalid tier ID.")
		return
	}

	tier, err := s.TierService.GetTier(id)
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to get tier: %s", err))
		return
	}

	names, err := s.TournamentService.GetNamesByTier(id)
	if err != nil {
		ServerErrorResponse(w, fmt.Sprintf("Failed to get names: %s", err))
		return
	}

	s.Render(w, 200, "tiers/view.go.html", "base", map[string]any{
		"Tier":  tier,
		"Names": names,
	})
}
