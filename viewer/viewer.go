package viewer

import (
	"encoding/json"
	"net/http"

	apierrors "github.com/joeyfromspace/go-api-errors/v2"
)

// JSONEnvelope describes the standard envelope format for json data
type JSONEnvelope struct {
	Data interface{} `json:"data"`
}

// SendJSON sends a json payload with the supplied status code
func SendJSON(w http.ResponseWriter, o interface{}, s int) {
	w.Header().Add("Content-Type", "application/json")
	j, err := json.Marshal(o)

	if err != nil {
		// TODO Add logging here
		apierrors.SendInternalError(w, nil)
		return
	}

	if s == 0 {
		s = 200
	}

	w.WriteHeader(s)
	w.Write(j)
}
