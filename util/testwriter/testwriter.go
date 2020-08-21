package testwriter

import (
	"net/http"
	"testing"
)

// ResponseWriter is a testing interface for http.ResponseWriter that allows for checking internal state, what is written in the response, and the returned status code and headers
type ResponseWriter struct {
	t                  *testing.T
	state              map[string]interface{}
	expectedStatusCode int
	header             map[string][]string
	expectedHeader     map[string][]string
}

// Header returns the internally stored header map of the ResponseWriter
func (w ResponseWriter) Header() http.Header {
	return w.header
}

// State returns the internally stored state map of the ResponseWriter
func (w ResponseWriter) State() map[string]interface{} {
	return w.state
}

// WriteHeader works the same as the base http.ResponseWriter function, except the status code is also commited to the internal state map under the "StatusCode" key
func (w ResponseWriter) WriteHeader(n int) {
	w.SetState("StatusCode", n)
}

// Write is a stub for http.ResponseWriter.Write(). The inputted []byte value is committed to the internal state map under the "Body" key
func (w ResponseWriter) Write(d []byte) (int, error) {
	w.SetState("Body", d)
	return 0, nil
}

// ExpectedHeader returns the header mapping expected by the invoking test
func (w ResponseWriter) ExpectedHeader() map[string][]string {
	return w.expectedHeader
}

// ExpectedStatusCode returns the expected status code expected by the invoking test
func (w ResponseWriter) ExpectedStatusCode() int {
	return w.expectedStatusCode
}

// StatusCode returns the status code written to the internal state (such as via the w.WriteHeader() method)
func (w ResponseWriter) StatusCode() int {
	s := w.State()
	v, ok := s["StatusCode"].(int)
	if !ok {
		return 0
	}
	return v
}

// Body returns the byte slice of the internally stored body (including the input of w.Write())
func (w ResponseWriter) Body() []byte {
	s := w.State()
	v, ok := s["Body"].([]byte)
	if !ok {
		return []byte("")
	}
	return v
}

// SetState commits a key/value pair to the internal state of the ResponseWriter
func (w ResponseWriter) SetState(k string, v interface{}) {
	st := w.State()
	st[k] = v
}

// NewOptions set expectations for the response writer instance to make writing assertions easier in unit tests
type NewOptions struct {
	ExpectedStatusCode int
	ExpectedHeader     map[string][]string
}

// New returns a test implementation of the http.ResponseWriter interface. The intent is to more easily assert its inner-workings
func New(o *NewOptions) ResponseWriter {
	eh := o.ExpectedHeader
	if eh == nil {
		eh = map[string][]string{}
	}

	esc := o.ExpectedStatusCode

	trw := ResponseWriter{
		header:             map[string][]string{},
		expectedHeader:     eh,
		expectedStatusCode: esc,
		state:              map[string]interface{}{},
	}

	return trw
}
