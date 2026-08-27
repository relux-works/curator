package godriver

import (
	"context"
	"errors"
	"io"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestOSExecutorClosesStdinAndCapturesBoundedOutput(t *testing.T) {
	environment := indispensableEnvironment()
	if environment == nil {
		environment = map[string]string{}
	}
	environment["GO_WANT_GODRIVER_HELPER"] = "stdin"
	output, err := (OSExecutor{}).Run(context.Background(), Process{
		Executable: os.Args[0], Arguments: []string{"-test.run=TestGodriverHelperProcess"},
		Directory: t.TempDir(), Environment: environmentSlice(environment), Timeout: 5 * time.Second, OutputLimit: 64,
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(output.Stdout) != "stdin=0" {
		t.Fatalf("output = stdout %q stderr %q", output.Stdout, output.Stderr)
	}
}

func TestOSExecutorEnforcesOutputAndTimeBounds(t *testing.T) {
	for _, test := range []struct {
		name, mode string
		timeout    time.Duration
		limit      int64
		want       error
	}{
		{name: "output", mode: "output", timeout: 5 * time.Second, limit: 32, want: errOutputLimit},
		{name: "combined output", mode: "combined-output", timeout: 5 * time.Second, limit: 32, want: errOutputLimit},
		{name: "timeout", mode: "timeout", timeout: 20 * time.Millisecond, limit: 32, want: errProcessTimeout},
	} {
		t.Run(test.name, func(t *testing.T) {
			environment := indispensableEnvironment()
			if environment == nil {
				environment = map[string]string{}
			}
			environment["GO_WANT_GODRIVER_HELPER"] = test.mode
			_, err := (OSExecutor{}).Run(context.Background(), Process{
				Executable: os.Args[0], Arguments: []string{"-test.run=TestGodriverHelperProcess"},
				Directory: t.TempDir(), Environment: environmentSlice(environment), Timeout: test.timeout, OutputLimit: test.limit,
			})
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestGodriverHelperProcess(_ *testing.T) {
	switch os.Getenv("GO_WANT_GODRIVER_HELPER") {
	case "":
		return
	case "stdin":
		payload, err := io.ReadAll(os.Stdin)
		if err != nil {
			os.Exit(2)
		}
		_, _ = os.Stdout.WriteString("stdin=" + strconv.Itoa(len(payload)))
		os.Exit(0)
	case "output":
		_, _ = os.Stdout.Write(make([]byte, 4096))
		os.Exit(0)
	case "combined-output":
		_, _ = os.Stdout.Write(make([]byte, 24))
		_, _ = os.Stderr.Write(make([]byte, 24))
		os.Exit(0)
	case "timeout":
		time.Sleep(5 * time.Second)
		os.Exit(0)
	default:
		os.Exit(2)
	}
}
