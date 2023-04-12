package convert

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var ErrUnrecognizedURL = errors.New("unrecognized url")

var Client = &http.Client{
	Timeout: 10 * time.Second,
}

// Get will use the Client to send the given request. It will then attempt to fill the Response type using the data in the response body.
// Alternatively, the Response can be an interface with the Tournament() and Entrants() methods.
func Get[Response any](req *http.Request) (*Response, error) {
	res, err := Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response code: %d", res.StatusCode)
	}

	var data Response
	err = json.NewDecoder(res.Body).Decode(&data)

	return &data, err
}
