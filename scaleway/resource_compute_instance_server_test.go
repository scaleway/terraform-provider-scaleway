package scaleway

import (
	"reflect"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

func Test_stateToAction(t *testing.T) {
	type args struct {
		state string
	}
	tests := []struct {
		name  string
		state string
		want  instance.ServerAction
	}{
		{
			name:  "Started",
			state: "started",
			want:  instance.ServerActionPoweron,
		},
		{
			name:  "Stopped",
			state: "stopped",
			want:  instance.ServerActionPoweroff,
		},
		{
			name:  "Standby",
			state: "standby",
			want:  instance.ServerActionStopInPlace,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stateToAction(tt.state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateToAction() = %v, want %v", got, tt.want)
			}
		})
	}
}
