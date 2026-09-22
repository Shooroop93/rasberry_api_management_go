package rclone

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestParseLogLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		want    string
		wantErr bool
	}{
		{
			name: "valid json",
			line: `{"msg":"Transferred: 10 MiB / 20 MiB, 50%"}`,
			want: "Transferred: 10 MiB / 20 MiB, 50%",
		},
		{
			name: "valid json with other fields",
			line: `{"level":"info","msg":"hello","source":"stats"}`,
			want: "hello",
		},
		{
			name: "msg is missing",
			line: `{"level":"info"}`,
			want: "",
		},
		{
			name:    "invalid json",
			line:    `{"msg":`,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLogLine(tt.line)

			if tt.wantErr {
				if err == nil {
					t.Fatal("ParseLogLine() error = nil, want error")
				}

				return
			}

			if err != nil {
				t.Fatalf("ParseLogLine() unexpected error = %v", err)
			}

			if got != tt.want {
				t.Errorf(
					"ParseLogLine() = %q, want %q",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestParseLogLineWrapsJSONError(t *testing.T) {
	_, err := ParseLogLine(`{"msg":`)

	if err == nil {
		t.Fatal("ParseLogLine() error = nil, want error")
	}

	var syntaxErr *json.SyntaxError

	if !errors.As(err, &syntaxErr) {
		t.Errorf(
			"ParseLogLine() error does not wrap *json.SyntaxError: %v",
			err,
		)
	}
}

func TestFindTransferred(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want string
	}{
		{
			name: "transferred first line",
			msg:  "Transferred: 10 MiB / 20 MiB, 50%\nErrors: 0",
			want: "Transferred: 10 MiB / 20 MiB, 50%",
		},
		{
			name: "transferred middle line",
			msg:  "Checks: 10 / 10\nTransferred: 15 MiB / 20 MiB, 75%\nErrors: 0",
			want: "Transferred: 15 MiB / 20 MiB, 75%",
		},
		{
			name: "trims spaces",
			msg:  "Errors: 0\n    Transferred: 20 MiB / 20 MiB, 100%    \nElapsed: 5s",
			want: "Transferred: 20 MiB / 20 MiB, 100%",
		},
		{
			name: "windows line endings",
			msg:  "Errors: 0\r\nTransferred: 1 MiB / 2 MiB, 50%\r\nElapsed: 1s",
			want: "Transferred: 1 MiB / 2 MiB, 50%",
		},
		{
			name: "not found",
			msg:  "Errors: 0\nChecks: 10 / 10\nElapsed: 5s",
			want: "",
		},
		{
			name: "empty message",
			msg:  "",
			want: "",
		},
		{
			name: "does not match transferred in middle",
			msg:  "Something Transferred: 10 MiB",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findTransferred(tt.msg)

			if got != tt.want {
				t.Errorf(
					"findTransferred() = %q, want %q",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestFormatProgressMessage(t *testing.T) {
	folder := "alice"
	profile := "yandex_disk"
	progress := "Transferred: 10 MiB / 20 MiB, 50%"

	got := FormatProgressMessage(
		folder,
		profile,
		progress,
	)

	want := "Происходит backup для пользователя: alice. " +
		"По профилю rclone: yandex_disk.\n" +
		"Transferred: 10 MiB / 20 MiB, 50%"

	if got != want {
		t.Errorf(
			"FormatProgressMessage() = %q, want %q",
			got,
			want,
		)
	}
}
