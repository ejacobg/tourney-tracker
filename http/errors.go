package http

import (
	"log"
	"net/http"
)

// ServerError logs the given error to the console, then responds with the given error message.
func ServerError(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
