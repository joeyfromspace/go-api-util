package bodyparser

import (
	"bytes"
	"net/http"
	httptest "net/http/httptest"
	"reflect"
	"testing"
)

func TestParseJSON(t *testing.T) {
	type args struct {
		r *http.Request
		i interface{}
	}
	type someObject struct {
		Name   string `json:"name"`
		IsCool bool   `json:"isCool"`
	}
	r := bytes.NewReader([]byte("{\"name\":\"foo\",\"isCool\":true}"))
	badR := bytes.NewReader([]byte("{\"name\":\"bar\""))
	testRequest := httptest.NewRequest(http.MethodPost, "https://www.example.com/cool/test", r)
	testRequestWithBadBody := httptest.NewRequest(http.MethodPost, "https://www.example.com/cool/test", badR)
	var testObject someObject
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		expected someObject
	}{
		{
			name:     "should decode to the expected interface",
			args:     args{r: testRequest, i: &testObject},
			wantErr:  false,
			expected: someObject{Name: "foo", IsCool: true},
		},
		{
			name:    "should return parsing errors",
			args:    args{r: testRequestWithBadBody, i: &testObject},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ParseJSON(tt.args.r, &tt.args.i); (err != nil) != tt.wantErr {
				t.Errorf("ParseJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(testObject, tt.expected) {
				t.Errorf("ParseJSON() unexpected output got %v, expected %v", testObject, tt.expected)
			}
		})
	}
}
