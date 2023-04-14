package http

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

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

// readIDParam returns the value of the :id route parameter, or an error if it could not be read.
func readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}
