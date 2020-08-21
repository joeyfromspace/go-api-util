package viewer

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"testing"

	"github.com/joeyfromspace/go-api-util/v2/util/testwriter"
)

type testStruct struct {
	Foo testStructInner `json:"foo"`
}

type testStructInner struct {
	Bar string `json:"bar"`
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
				w: testwriter.New(&testwriter.NewOptions{
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
			tw := tt.args.w.(testwriter.ResponseWriter)
			for key, value := range tw.Header() {
				if len(tw.ExpectedHeader()[key]) > 0 && !reflect.DeepEqual(value, tw.ExpectedHeader()[key]) {
					t.Errorf("ExpectedHeader() mismatch. got %s, expecting %s", value, tw.ExpectedHeader()[key])
				}
			}
			if got := tw.StatusCode(); !reflect.DeepEqual(got, tw.ExpectedStatusCode()) {
				t.Errorf("ExpectedStatusCode() mismatch. got %s, expecting %s", strconv.Itoa(got), strconv.Itoa(tw.ExpectedStatusCode()))
			}
			var v testStruct
			err := json.Unmarshal(tw.Body(), &v)
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
