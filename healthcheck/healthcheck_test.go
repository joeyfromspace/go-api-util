package healthcheck

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/joeyfromspace/go-api-util/v2/util/testwriter"
)

func TestHealthyStatus_String(t *testing.T) {
	tests := []struct {
		name string
		h    HealthyStatus
		want string
	}{
		{
			name: "should convert to the correct string for Healthy",
			h:    Healthy,
			want: "Healthy",
		},
		{
			name: "should convert to the correct string for Unhealthy",
			h:    Unhealthy,
			want: "Unhealthy",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.h.String(); got != tt.want {
				t.Errorf("HealthyStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		o *NewOptions
	}
	tests := []struct {
		name     string
		args     args
		testFunc func(t *testing.T, h http.HandlerFunc)
	}{
		{
			name: "Should write a 500 status for unhealthy process",
			args: args{
				o: &NewOptions{
					Function: func() bool {
						return false
					},
				},
			},
			testFunc: func(t *testing.T, h http.HandlerFunc) {
				tw := testwriter.New(&testwriter.NewOptions{
					ExpectedHeader: map[string][]string{
						"Content-Type": {"application/json"},
					},
					ExpectedStatusCode: 500,
				})
				tr := new(http.Request)
				h(tw, tr)
				if got := tw.StatusCode(); got != tw.ExpectedStatusCode() {
					t.Errorf("unexpected status code in response: got %s, expected %s", strconv.Itoa(got), strconv.Itoa(tw.ExpectedStatusCode()))
				}
				if got := tw.Header()["Content-Type"][0]; got != tw.ExpectedHeader()["Content-Type"][0] {
					t.Errorf("unexpected header in response: got %s, expected %s", got, tw.ExpectedHeader()["Content-Type"][0])
				}
			},
		},
		{
			name: "should response with 200 status for healthy process",
			args: args{
				o: &NewOptions{
					Function: func() bool {
						return true
					},
				},
			},
			testFunc: func(t *testing.T, h http.HandlerFunc) {
				tw := testwriter.New(&testwriter.NewOptions{
					ExpectedHeader: map[string][]string{
						"Content-Type": {"application/json"},
					},
					ExpectedStatusCode: 200,
				})
				tr := new(http.Request)
				h(tw, tr)
				if got := tw.StatusCode(); got != tw.ExpectedStatusCode() {
					t.Errorf("unexpected status code in response: got %s, expected %s", strconv.Itoa(got), strconv.Itoa(tw.ExpectedStatusCode()))
				}
			},
		},
		{
			name: "should respond with the expected body",
			args: args{
				o: &NewOptions{
					Function: func() bool {
						return true
					},
				},
			},
			testFunc: func(t *testing.T, h http.HandlerFunc) {
				tw := testwriter.New(&testwriter.NewOptions{
					ExpectedHeader: map[string][]string{
						"Content-Type": {"application/json"},
					},
					ExpectedStatusCode: 200,
				})
				tr := new(http.Request)
				time.Sleep(1500 * time.Millisecond)
				h(tw, tr)
				b := tw.Body()
				var hr Response
				err := json.Unmarshal(b, &hr)
				if err != nil {
					t.Errorf("unexpected result unmarshalling json response: %s", err)
				}
				if hr.Uptime.Seconds() < 1 {
					t.Errorf("unexpected uptime value: %s", strconv.Itoa(int(hr.Uptime.Seconds())))
				}
				if hr.Health != Healthy {
					t.Errorf("unexpected health value: got %s, expected %s", hr.Health, Healthy)
				}
				if !hr.CurrentTime.Before(time.Now()) {
					t.Errorf("unexpected time value: got %s, expected < %s", hr.CurrentTime, time.Now())
				}
				if hr.TimeInCurrentState.String() != hr.Uptime.String() {
					t.Errorf("expected time in current state to match uptime. Got %s, expected %s", hr.TimeInCurrentState, hr.Uptime)
				}
			},
		},
		{
			name: "should be healthy by default",
			args: args{
				o: &NewOptions{},
			},
			testFunc: func(t *testing.T, h http.HandlerFunc) {
				tw := testwriter.New(&testwriter.NewOptions{
					ExpectedHeader: map[string][]string{
						"Content-Type": {"application/json"},
					},
					ExpectedStatusCode: 200,
				})
				tr := new(http.Request)
				h(tw, tr)
				if got := tw.StatusCode(); got != tw.ExpectedStatusCode() {
					t.Errorf("unexpected status code in response: got %s, expected %s", strconv.Itoa(got), strconv.Itoa(tw.ExpectedStatusCode()))
				}
			},
		},
	}
	for _, tt := range tests {
		h := New(tt.args.o)
		tt.testFunc(t, h)
	}
}
