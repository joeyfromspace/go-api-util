package viewer

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"testing"
)

type testStruct struct {
	Foo testStructInner `json:"foo"`
}

type testStructInner struct {
	Bar string `json:"bar"`
}

type TestResponseWriter struct {
	t                  *testing.T
	state              map[string]interface{}
	expectedStatusCode int
	header             map[string][]string
	expectedHeader     map[string][]string
}

func (w TestResponseWriter) Header() http.Header {
	return w.header
}

func (w TestResponseWriter) State() map[string]interface{} {
	return w.state
}

func (w TestResponseWriter) WriteHeader(n int) {
	w.SetState("StatusCode", n)
}

func (w TestResponseWriter) Write(d []byte) (int, error) {
	w.SetState("Body", d)
	return 0, nil
}

func (w TestResponseWriter) ExpectedHeader() map[string][]string {
	return w.expectedHeader
}

func (w TestResponseWriter) ExpectedStatusCode() int {
	return w.expectedStatusCode
}

func (w TestResponseWriter) StatusCode() int {
	s := w.State()
	v, ok := s["StatusCode"].(int)
	if !ok {
		return 0
	}
	return v
}

func (w TestResponseWriter) WrittenBody() []byte {
	s := w.State()
	v, ok := s["Body"].([]byte)
	if !ok {
		return []byte("")
	}
	return v
}

func (w TestResponseWriter) SetState(k string, v interface{}) {
	st := w.State()
	st[k] = v
}

type testWriterOptions struct {
	ExpectedStatusCode int
	ExpectedHeader     map[string][]string
}

func NewTestResponseWriter(o *testWriterOptions) TestResponseWriter {
	eh := o.ExpectedHeader
	if eh == nil {
		eh = map[string][]string{}
	}

	esc := o.ExpectedStatusCode

	trw := TestResponseWriter{
		header:             map[string][]string{},
		expectedHeader:     eh,
		expectedStatusCode: esc,
		state:              map[string]interface{}{},
	}

	return trw
}

func TestSendJSON(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		o interface{}
		s int
	}
	tests := []struct {
		name string
		args args
	}{
		0: {
			name: "should not error",
			args: args{
				w: NewTestResponseWriter(&testWriterOptions{
					ExpectedHeader: map[string][]string{
						"Content-Type": {"application/json"},
					},
					ExpectedStatusCode: 201,
				}),
				o: testStruct{
					Foo: testStructInner{
						Bar: "baz",
					},
				},
				s: 201,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SendJSON(tt.args.w, tt.args.o, tt.args.s)
			tw := tt.args.w.(TestResponseWriter)
			for key, value := range tw.Header() {
				if len(tw.ExpectedHeader()[key]) > 0 && !reflect.DeepEqual(value, tw.ExpectedHeader()[key]) {
					t.Errorf("ExpectedHeader() mismatch. got %s, expecting %s", value, tw.ExpectedHeader()[key])
				}
			}
			if got := tw.StatusCode(); !reflect.DeepEqual(got, tw.ExpectedStatusCode()) {
				t.Errorf("ExpectedStatusCode() mismatch. got %s, expecting %s", strconv.Itoa(got), strconv.Itoa(tw.ExpectedStatusCode()))
			}
			var v testStruct
			err := json.Unmarshal(tw.WrittenBody(), &v)
			if err != nil {
				t.Errorf("Unmarshal error: %s", err)
				return
			}
			if got := v; !reflect.DeepEqual(got, tt.args.o) {
				t.Errorf("Body mismatch. Got %s, expected %s", got, v)
			}
		})
	}
}
