package tournament

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

// Controller provides several HTTP handlers for servicing tournament-related requests.
type Controller struct {
	Model Model
	Views struct {
		Index, View, Edit *template.Template
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
