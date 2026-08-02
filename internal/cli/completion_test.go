package cli

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
	"github.com/vicgarci/sadb/adb/adbtest"
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

// cmdWithSerial returns a minimal cobra command with the -s/--serial flag set.
func cmdWithSerial(serial string) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("serial", "s", "", "")
	if serial != "" {
		_ = cmd.Flags().Set("serial", serial)
	}
	return cmd
}

// Package completion must fetch packages directly without opening the Package Search TUI —
// the uninstall command's interactive picker path is a separate flow tested in uninstall_test.go.

func TestCompletePackages_ReturnsFetchedPackages(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("package:com.example.foo\npackage:com.example.bar\n", nil)

	got, directive := makeCompletePackages(f)(cmdWithSerial("emulator-5554"), nil, "")

	if len(f.Calls) != 1 || f.Calls[0].Args[0] != "shell" {
		t.Fatalf("expected one pm list packages call, got %+v", f.Calls)
	}
	want := []string{"com.example.foo", "com.example.bar"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("packages = %v, want %v", got, want)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("unexpected directive %v", directive)
	}
}

func TestCompletePackages_RunnerError_ReturnsEmpty(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", errFake("adb error"))

	got, _ := makeCompletePackages(f)(cmdWithSerial("emulator-5554"), nil, "")

	if len(got) != 0 {
		t.Errorf("expected empty on error, got %v", got)
	}
}

