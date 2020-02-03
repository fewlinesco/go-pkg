package httphandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var ErrCantEncodeJSON = errors.New("can't encode json response")

type HTTPResponse interface {
	HTTPCode() int
}

func WriteJSON(w http.ResponseWriter, data HTTPResponse) error {
	json, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCantEncodeJSON, err)
	}

	w.WriteHeader(data.HTTPCode())
	w.Write(json)

	return nil
}
