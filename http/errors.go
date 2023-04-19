package http

import (
	"log"
	"net/http"
)

// ErrorResponse response with and logs the given error message to the console, alongside the given response code.
// Any error messages will be rendered inside the #error element of the page.
func ErrorResponse(w http.ResponseWriter, error string, code int) {
	log.Println(error)

	w.Header()["HX-Reswap"] = []string{"innerHTML"}
	w.Header()["HX-Retarget"] = []string{"#error"}
	http.Error(w, error, code)
}

func ServerErrorResponse(w http.ResponseWriter, error string) {
	ErrorResponse(w, error, http.StatusInternalServerError)
}

func BadRequestResponse(w http.ResponseWriter, error string) {
	ErrorResponse(w, error, http.StatusBadRequest)
}

func UnprocessableEntityResponse(w http.ResponseWriter, error string) {
	ErrorResponse(w, error, http.StatusUnprocessableEntity)
}

func NotFoundResponse(w http.ResponseWriter, error string) {
	ErrorResponse(w, error, http.StatusNotFound)
}

func MethodNotAllowedResponse(w http.ResponseWriter, error string) {
	ErrorResponse(w, error, http.StatusMethodNotAllowed)
}
