package rclone

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunReadsStdoutAndStderr(t *testing.T) {
	cmd := helperCommand(t, "mixed-output")

	var runErr error

	output := captureStdout(t, func() {
		runErr = Run(cmd)
	})

	if runErr != nil {
		t.Fatalf("Run() error = %v", runErr)
	}

	if !strings.Contains(
		output,
		"Transferred: message from stdout",
	) {
		t.Errorf(
			"Run() output does not contain stdout message: %q",
			output,
		)
	}

	if !strings.Contains(
		output,
		"Transferred: message from stderr",
	) {
		t.Errorf(
			"Run() output does not contain stderr message: %q",
			output,
		)
	}
}

func TestRunReturnsWaitError(t *testing.T) {
	cmd := helperCommand(t, "exit-error")

	err := Run(cmd)

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

	err := Run(cmd)

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

	err := Run(cmd)

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

	err := Run(cmd)

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

	err := Run(cmd)

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

		// Run должен получить parseErr и убить этот процесс.
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

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = writer

	fn()

	err = writer.Close()
	if err != nil {
		os.Stdout = originalStdout
		t.Fatal(err)
	}

	os.Stdout = originalStdout

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}

	err = reader.Close()
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}
