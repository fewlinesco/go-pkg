package httphandler

import (
	"encoding/json"
	"fmt"
	"github.com/fewlinesco/go-pkg/erroring"
	"net/http"
)

func ReadJSON(r *http.Request, operation erroring.Operation, input interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(input); err != nil {
		return &erroring.Error{
			Operation: operation,
			Kind:      erroring.KindUnparsable,
			Source:    erroring.SourceClient,
			Err:       fmt.Errorf("can't parse body: %v", err),
		}
	}
	return nil
}
