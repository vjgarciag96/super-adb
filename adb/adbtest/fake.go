// Package adbtest provides a fake adb.Runner for use in tests.
// Import this package from any test file that needs to record and assert on ADB calls.
package adbtest

// Call records a single invocation of the fake runner.
type Call struct {
	Serial string
	Args   []string
}

// Response is the configured output for a single queued call.
type Response struct {
	Output string
	Err    error
}

// FakeRunner records every ADB call made through it.
// Responses are consumed in FIFO order; if the queue is exhausted the call returns ("", nil).
type FakeRunner struct {
	Calls     []Call
	responses []Response
}

// QueueResponse enqueues a response that will be returned for the next Run call.
func (f *FakeRunner) QueueResponse(output string, err error) {
	f.responses = append(f.responses, Response{Output: output, Err: err})
}

// Run records the call and returns the next queued response.
func (f *FakeRunner) Run(serial string, args ...string) (string, error) {
	// Capture a copy of args so callers can't mutate the recorded slice.
	argsCopy := make([]string, len(args))
	copy(argsCopy, args)
	f.Calls = append(f.Calls, Call{Serial: serial, Args: argsCopy})

	if len(f.responses) == 0 {
		return "", nil
	}
	resp := f.responses[0]
	f.responses = f.responses[1:]
	return resp.Output, resp.Err
}
