package http

import "net/http"

func (s *Server) registerFormulaRoutes() {
	s.router.HandlerFunc(http.MethodGet, "/about", s.getFormula)
}

func (s *Server) getFormula(w http.ResponseWriter, _ *http.Request) {
	s.Render(w, 200, "about.go.html", "base", nil)
}
