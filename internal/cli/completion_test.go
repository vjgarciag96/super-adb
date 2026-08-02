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

// cmdWithSerial returns a minimal cobra command with the -s/--serial flag set to serial.
func cmdWithSerial(serial string) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("serial", "s", "", "")
	if serial != "" {
		_ = cmd.Flags().Set("serial", serial)
	}
	return cmd
}

func TestCompletePackages_UsesFlagSerial(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("package:com.example.foo\npackage:com.example.bar\n", nil)

	fn := makeCompletePackages(f)
	got, directive := fn(cmdWithSerial("emulator-5554"), nil, "")

	if len(f.Calls) != 1 {
		t.Fatalf("expected 1 ADB call, got %d", len(f.Calls))
	}
	if f.Calls[0].Serial != "emulator-5554" {
		t.Errorf("expected serial %q, got %q", "emulator-5554", f.Calls[0].Serial)
	}
	want := []string{"com.example.foo", "com.example.bar"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("packages = %v, want %v", got, want)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("unexpected directive %v", directive)
	}
}

func TestCompletePackages_FallsBackToEnv(t *testing.T) {
	t.Setenv("SADB_DEVICE", "R5CR71L28YB")

	f := &adbtest.FakeRunner{}
	f.QueueResponse("package:com.example.baz\n", nil)

	fn := makeCompletePackages(f)
	fn(cmdWithSerial(""), nil, "")

	if f.Calls[0].Serial != "R5CR71L28YB" {
		t.Errorf("expected env serial %q, got %q", "R5CR71L28YB", f.Calls[0].Serial)
	}
}

func TestCompletePackages_RunnerError_ReturnsEmpty(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", errFake("adb error"))

	fn := makeCompletePackages(f)
	got, directive := fn(cmdWithSerial("emulator-5554"), nil, "")

	if len(got) != 0 {
		t.Errorf("expected empty packages on error, got %v", got)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("unexpected directive %v", directive)
	}
}
