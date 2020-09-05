package bodyparser

import (
	"encoding/json"
	"net/http"
)

// ParseJSON will take the byte stream of a request body and unmarshal it into the inputted interface
func ParseJSON(r *http.Request, i interface{}) error {
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&i)
	if err != nil {
		return err
	}
	return nil
}
