package cli

import (
	"reflect"
	"testing"
)

func TestParseDeviceSerials(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name: "single physical device",
			input: `List of devices attached
R5CR71L28YB	device`,
			want: []string{"R5CR71L28YB"},
		},
		{
			name: "emulator and physical device",
			input: `List of devices attached
emulator-5554	device
R5CR71L28YB	device`,
			want: []string{"emulator-5554", "R5CR71L28YB"},
		},
		{
			name: "includes unauthorized devices",
			input: `List of devices attached
R5CR71L28YB	unauthorized`,
			want: []string{"R5CR71L28YB"},
		},
		{
			name:  "no devices",
			input: "List of devices attached",
			want:  nil,
		},
		{
			name:  "empty output",
			input: "",
			want:  nil,
		},
		{
			name: "skips offline devices",
			input: `List of devices attached
R5CR71L28YB	offline`,
			want: nil,
		},
		{
			name: "tcp/ip device",
			input: `List of devices attached
192.168.1.100:5555	device`,
			want: []string{"192.168.1.100:5555"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDeviceSerials(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDeviceSerials() = %v, want %v", got, tt.want)
			}
		})
	}
}

