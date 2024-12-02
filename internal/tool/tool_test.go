package tool

import (
	"testing"
	"time"
)

func Test_parseTimeString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Duration
		wantErr bool
	}{
		{"test1", args{"6912000s"}, 6912000 * time.Second, false},
		{"test2", args{"12371s"}, 12371 * time.Second, false},
		{"test3", args{"12371m"}, 12371 * time.Minute, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTimeString(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTimeString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseTimeString() got = %v, want %v", got, tt.want)
			}
		})
	}
}
