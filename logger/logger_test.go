package logger

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestInitialize(t *testing.T) {
	var singleton *logrus.Logger
	type args struct {
		o *Options
	}
	tests := []struct {
		name   string
		args   args
		before func(t *testing.T)
		after  func(t *testing.T)
	}{
		{
			name: "should initialize with defaults",
			args: args{
				o: &Options{},
			},
			before: func(t *testing.T) {
				Reset()
			},
			after: func(t *testing.T) {
				singleton = Log()
				if got := singleton.GetLevel(); got != logrus.InfoLevel {
					t.Errorf("unexpected default level. got %s, expected %s", got, logrus.InfoLevel)
				}
				if got := len(singleton.Hooks); got != 0 {
					t.Errorf("unexpected hooks length. got %s, expected %s", strconv.Itoa(got), strconv.Itoa(0))
				}
			},
		},
		{
			name: "should return identical singleton",
			args: args{
				o: &Options{},
			},
			after: func(t *testing.T) {
				if !reflect.DeepEqual(Log(), singleton) {
					t.Errorf("unexpected logger returned.")
				}
			},
		},
		{
			name: "should reflect custom level",
			args: args{
				o: &Options{
					Level: logrus.ErrorLevel,
				},
			},
			before: func(t *testing.T) {
				Reset()
			},
			after: func(t *testing.T) {
				newSingleton := Log()
				if reflect.DeepEqual(newSingleton, singleton) {
					t.Errorf("unexpected logger equality.")
				}
				singleton = newSingleton
				if got := singleton.Level; got != logrus.ErrorLevel {
					t.Errorf("unexpected singleton error level. got %s, expecting %s", got, logrus.ErrorLevel)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.after != nil {
				defer tt.after(t)
			}
			if tt.before != nil {
				tt.before(t)
			}
			if tt.before == nil && tt.after == nil {
				t.Errorf("%s: no before or after hooks, nothing tested!", tt.name)
			}
			Initialize(tt.args.o)
		})
	}
}

func TestLog(t *testing.T) {
	var singleton *logrus.Logger

	tests := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)
		want   *logrus.Logger
	}{
		0: {
			name: "ensure logger crashes if uninitialized",
			before: func(t *testing.T) {
				Reset()
			},
			after: func(t *testing.T) {
				if r := recover(); r == nil {
					t.Errorf("The code did not panic")
				}
			},
			want: nil,
		},
		1: {
			name: "ensure logger returns the expected logging instance",
			before: func(t *testing.T) {
				singleton = Initialize(&Options{})
			},
			want: singleton,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.after != nil {
				defer tt.after(t)
			}
			if tt.before != nil {
				tt.before(t)
			}
			if got := Log(); tt.want != nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Log() = %v, want %v", got, tt.want)
			}
		})
	}
}
