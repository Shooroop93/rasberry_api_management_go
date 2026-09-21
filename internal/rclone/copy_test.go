package rclone

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func validCopyOptions() CopyOptions {
	return CopyOptions{
		BackupRoot:    "/srv/backups",
		StatsInterval: 5 * time.Second,
		Checkers:      8,
		Transfers:     4,
	}
}

func TestBuildCopyArgs_ValidInput(t *testing.T) {
	t.Parallel()
	opts := validCopyOptions()

	got, err := BuildCopyArgs(opts, "alice", "gdrive")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{
		"copy",
		filepath.Join("/srv/backups", "alice"),
		"gdrive:alice",
		"--use-json-log",
		"--log-level",
		"INFO",
		"--stats=5s",
		"--create-empty-src-dirs",
		"--checkers=8",
		"--transfers=4",
	}

	if !slices.Equal(got, want) {
		t.Fatalf("unexpected args:\nwant: %#v\ngot:  %#v", want, got)
	}
}

func TestBuildCopyArgs_BackupRootWithAndWithoutTrailingSeparator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		root string
	}{
		{
			name: "without trailing separator",
			root: "/srv/backups",
		},
		{
			name: "with trailing separator",
			root: "/srv/backups/",
		},
	}

	wantSource := filepath.Join("/srv/backups", "alice")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := validCopyOptions()
			opts.BackupRoot = tt.root

			got, err := BuildCopyArgs(opts, "alice", "gdrive")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got[1] != wantSource {
				t.Fatalf("unexpected source: want %q, got %q", wantSource, got[1])
			}
		})
	}
}

func TestBuildCopyArgs_PathWithSpaces(t *testing.T) {
	opts := validCopyOptions()
	opts.BackupRoot = "/srv/my backups"

	got, err := BuildCopyArgs(opts, "alice data", "gdrive")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantSource := filepath.Join("/srv/my backups", "alice data")

	if got[1] != wantSource {
		t.Fatalf("unexpected source: want %q, got %q", wantSource, got[1])
	}

	if got[2] != "gdrive:alice data" {
		t.Fatalf("unexpected destination: %q", got[2])
	}

	if len(got) != 10 {
		t.Fatalf("expected 10 arguments, got %d", len(got))
	}

	if strings.ContainsAny(got[1], "\"'") {
		t.Fatalf("source must not contain shell quotes: %q", got[1])
	}
}

func TestBuildCopyArgs_InvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		call    func() ([]string, error)
		wantErr string
	}{
		{
			name: "empty backup root",
			call: func() ([]string, error) {
				opts := validCopyOptions()
				opts.BackupRoot = ""
				return BuildCopyArgs(opts, "alice", "gdrive")
			},
			wantErr: "backupRoot must not be empty",
		},
		{
			name: "whitespace backup root",
			call: func() ([]string, error) {
				opts := validCopyOptions()
				opts.BackupRoot = "   "
				return BuildCopyArgs(opts, "alice", "gdrive")
			},
			wantErr: "backupRoot must not be empty",
		},
		{
			name: "empty folder name",
			call: func() ([]string, error) {
				return BuildCopyArgs(validCopyOptions(), "", "gdrive")
			},
			wantErr: "folderName must not be empty",
		},
		{
			name: "whitespace folder name",
			call: func() ([]string, error) {
				return BuildCopyArgs(validCopyOptions(), "   ", "gdrive")
			},
			wantErr: "folderName must not be empty",
		},
		{
			name: "empty profile",
			call: func() ([]string, error) {
				return BuildCopyArgs(validCopyOptions(), "alice", "")
			},
			wantErr: "profile must not be empty",
		},
		{
			name: "whitespace profile",
			call: func() ([]string, error) {
				return BuildCopyArgs(validCopyOptions(), "alice", "   ")
			},
			wantErr: "profile must not be empty",
		},
		{
			name: "zero checkers",
			call: func() ([]string, error) {
				opts := validCopyOptions()
				opts.Checkers = 0
				return BuildCopyArgs(opts, "alice", "gdrive")
			},
			wantErr: "checkers must be greater than zero",
		},
		{
			name: "negative checkers",
			call: func() ([]string, error) {
				opts := validCopyOptions()
				opts.Checkers = -1
				return BuildCopyArgs(opts, "alice", "gdrive")
			},
			wantErr: "checkers must be greater than zero",
		},
		{
			name: "zero transfers",
			call: func() ([]string, error) {
				opts := validCopyOptions()
				opts.Transfers = 0
				return BuildCopyArgs(opts, "alice", "gdrive")
			},
			wantErr: "transfers must be greater than zero",
		},
		{
			name: "negative transfers",
			call: func() ([]string, error) {
				opts := validCopyOptions()
				opts.Transfers = -1
				return BuildCopyArgs(opts, "alice", "gdrive")
			},
			wantErr: "transfers must be greater than zero",
		},
		{
			name: "zero stats interval",
			call: func() ([]string, error) {
				opts := validCopyOptions()
				opts.StatsInterval = 0
				return BuildCopyArgs(opts, "alice", "gdrive")
			},
			wantErr: "stats interval must be greater than zero",
		},
		{
			name: "negative stats interval",
			call: func() ([]string, error) {
				opts := validCopyOptions()
				opts.StatsInterval = -time.Second
				return BuildCopyArgs(opts, "alice", "gdrive")
			},
			wantErr: "stats interval must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := tt.call()

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.wantErr {
				t.Fatalf("unexpected error: want %q, got %q", tt.wantErr, err.Error())
			}

			if args != nil {
				t.Fatalf("expected nil args on error, got %#v", args)
			}
		})
	}
}
