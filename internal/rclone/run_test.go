package rclone

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestRunReadsStdoutAndStderr(t *testing.T) {
	cmd := helperCommand(t, "mixed-output")

	var progress []string

	err := Run(cmd, func(value string) {
		progress = append(progress, value)
	})

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !slices.Contains(
		progress,
		"Transferred: message from stdout",
	) {
		t.Errorf(
			"Run() progress = %v, want stdout progress",
			progress,
		)
	}

	if !slices.Contains(
		progress,
		"Transferred: message from stderr",
	) {
		t.Errorf(
			"Run() progress = %v, want stderr progress",
			progress,
		)
	}

	if len(progress) != 2 {
		t.Errorf(
			"Run() progress count = %d, want 2: %v",
			len(progress),
			progress,
		)
	}
}

func TestRunReturnsWaitError(t *testing.T) {
	cmd := helperCommand(t, "exit-error")

	err := Run(cmd, func(string) {})

	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}

	if !strings.Contains(
		err.Error(),
		"wait rclone process",
	) {
		t.Errorf(
			"Run() error = %q, want wait rclone process error",
			err,
		)
	}

	var exitErr *exec.ExitError

	if !errors.As(err, &exitErr) {
		t.Fatalf(
			"Run() error does not wrap *exec.ExitError: %v",
			err,
		)
	}

	if exitErr.ExitCode() != 7 {
		t.Errorf(
			"exit code = %d, want 7",
			exitErr.ExitCode(),
		)
	}
}

func TestRunReturnsStartError(t *testing.T) {
	command := filepath.Join(
		t.TempDir(),
		"command-that-does-not-exist",
	)

	cmd := exec.Command(command)

	err := Run(cmd, func(string) {})

	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "start failed") {
		t.Errorf(
			"Run() error = %q, want start failed error",
			err,
		)
	}
}

func TestRunReturnsStdoutPipeError(t *testing.T) {
	cmd := helperCommand(t, "mixed-output")

	cmd.Stdout = io.Discard

	err := Run(cmd, func(string) {})

	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}

	if !strings.Contains(
		err.Error(),
		"stdout pipe failed",
	) {
		t.Errorf(
			"Run() error = %q, want stdout pipe failed error",
			err,
		)
	}
}

func TestRunReturnsScannerError(t *testing.T) {
	cmd := helperCommand(t, "long-line")

	err := Run(cmd, func(string) {})

	if err == nil {
		t.Fatal("Run() error = nil, want scanner error")
	}

	if !strings.Contains(
		err.Error(),
		"read rclone output",
	) {
		t.Errorf(
			"Run() error = %q, want scanner error",
			err,
		)
	}
}

func TestRunReturnsParseError(t *testing.T) {
	cmd := helperCommand(t, "invalid-json")

	err := Run(cmd, func(string) {})

	if err == nil {
		t.Fatal("Run() error = nil, want parse error")
	}

	if !strings.Contains(
		err.Error(),
		"parse rclone output",
	) {
		t.Errorf(
			"Run() error = %q, want parse rclone output error",
			err,
		)
	}

	var syntaxErr *json.SyntaxError

	if !errors.As(err, &syntaxErr) {
		t.Errorf(
			"Run() error does not wrap *json.SyntaxError: %v",
			err,
		)
	}
}

func helperCommand(t *testing.T, scenario string) *exec.Cmd {
	t.Helper()

	cmd := exec.Command(
		os.Args[0],
		"-test.run=TestRunHelperProcess",
		"--",
		scenario,
	)

	cmd.Env = append(
		os.Environ(),
		"GO_RCLONE_RUN_HELPER=1",
	)

	return cmd
}

func TestRunHelperProcess(t *testing.T) {
	if os.Getenv("GO_RCLONE_RUN_HELPER") != "1" {
		return
	}

	scenario := os.Args[len(os.Args)-1]

	switch scenario {
	case "mixed-output":
		fmt.Fprintln(
			os.Stdout,
			`{"msg":"Transferred: message from stdout"}`,
		)

		fmt.Fprintln(
			os.Stderr,
			`{"msg":"ordinary rclone log"}`,
		)

		fmt.Fprintln(
			os.Stderr,
			`{"msg":"Transferred: message from stderr"}`,
		)

		os.Exit(0)

	case "exit-error":
		fmt.Fprintln(
			os.Stderr,
			`{"msg":"process failed"}`,
		)

		os.Exit(7)

	case "invalid-json":
		fmt.Fprintln(
			os.Stdout,
			`not json`,
		)

		// Run должен получить parseErr и остановить процесс.
		time.Sleep(30 * time.Second)

		os.Exit(0)

	case "long-line":
		fmt.Fprintln(
			os.Stdout,
			strings.Repeat("x", 70*1024),
		)

		os.Exit(0)

	default:
		os.Exit(2)
	}
}
