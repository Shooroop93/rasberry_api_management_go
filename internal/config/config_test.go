package config

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLoadSuccess(t *testing.T) {
	setValidEnv(t)

	got, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Config{
		TelegramToken:  "test-token",
		TelegramChatID: "12345",

		RcloneEnabled:      true,
		StartBackupAtStart: false,
		RcloneProfiles: []string{
			"yandex_disk",
			"ssd_toshiba",
		},
		BackupRoot:       "/home/admin/backup",
		RcloneConfigPath: "/home/admin/.config/rclone/rclone.conf",
		RcloneCron:       "0 0 3 ? * MON,WED,SAT",

		StatsInterval: 5 * time.Second,
		Checkers:      8,
		Transfers:     4,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf(
			"unexpected config:\ngot:  %+v\nwant: %+v",
			got,
			want,
		)
	}
}

func TestLoadReturnsErrorForMissingRequiredEnv(t *testing.T) {
	tests := []struct {
		name string
		env  string
	}{
		{
			name: "telegram token",
			env:  "TELEGRAM_TOKEN",
		},
		{
			name: "telegram chat id",
			env:  "TELEGRAM_CHAT_ID",
		},
		{
			name: "rclone enabled",
			env:  "RCLONE_ENABLED",
		},
		{
			name: "start backup at start",
			env:  "START_BACKUP_AT_START",
		},
		{
			name: "rclone profiles",
			env:  "RCLONE_PROFILES",
		},
		{
			name: "backup root",
			env:  "BACKUP_ROOT",
		},
		{
			name: "rclone config path",
			env:  "RCLONE_CONFIG_PATH",
		},
		{
			name: "rclone cron",
			env:  "RCLONE_CRON",
		},
		{
			name: "stats interval",
			env:  "STATS_INTERVAL",
		},
		{
			name: "checkers",
			env:  "CHECKERS",
		},
		{
			name: "transfers",
			env:  "TRANSFERS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setValidEnv(t)

			t.Setenv(tt.env, "")

			_, err := Load()
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			want := tt.env + " is required"

			if !strings.Contains(err.Error(), want) {
				t.Fatalf(
					"unexpected error: got %q, want error containing %q",
					err.Error(),
					want,
				)
			}
		})
	}
}

func TestLoadReturnsErrorForInvalidBool(t *testing.T) {
	tests := []struct {
		name string
		env  string
	}{
		{
			name: "invalid rclone enabled",
			env:  "RCLONE_ENABLED",
		},
		{
			name: "invalid start backup at start",
			env:  "START_BACKUP_AT_START",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setValidEnv(t)

			t.Setenv(tt.env, "not-bool")

			_, err := Load()
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			want := "invalid " + tt.env

			if !strings.Contains(err.Error(), want) {
				t.Fatalf(
					"unexpected error: got %q, want error containing %q",
					err.Error(),
					want,
				)
			}
		})
	}
}

func TestLoadTrimsRcloneProfiles(t *testing.T) {
	setValidEnv(t)

	t.Setenv(
		"RCLONE_PROFILES",
		" yandex_disk , ssd_toshiba ",
	)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{
		"yandex_disk",
		"ssd_toshiba",
	}

	if !reflect.DeepEqual(cfg.RcloneProfiles, want) {
		t.Fatalf(
			"unexpected profiles: got %#v, want %#v",
			cfg.RcloneProfiles,
			want,
		)
	}
}

func TestLoadReturnsErrorForEmptyRcloneProfile(t *testing.T) {
	setValidEnv(t)

	t.Setenv(
		"RCLONE_PROFILES",
		"yandex_disk, ,ssd_toshiba",
	)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "RCLONE_PROFILES contains empty profile") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadReturnsErrorForInvalidStatsInterval(t *testing.T) {
	setValidEnv(t)

	t.Setenv("STATS_INTERVAL", "five seconds")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "invalid STATS_INTERVAL") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadReturnsErrorForNonPositiveStatsInterval(t *testing.T) {
	setValidEnv(t)

	t.Setenv("STATS_INTERVAL", "0s")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(
		err.Error(),
		"STATS_INTERVAL must be greater than zero",
	) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadReturnsErrorForInvalidInteger(t *testing.T) {
	tests := []struct {
		name string
		env  string
	}{
		{
			name: "invalid checkers",
			env:  "CHECKERS",
		},
		{
			name: "invalid transfers",
			env:  "TRANSFERS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setValidEnv(t)

			t.Setenv(tt.env, "not-number")

			_, err := Load()
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			want := "invalid " + tt.env

			if !strings.Contains(err.Error(), want) {
				t.Fatalf(
					"unexpected error: got %q, want error containing %q",
					err.Error(),
					want,
				)
			}
		})
	}
}

func TestLoadReturnsErrorForNonPositiveInteger(t *testing.T) {
	tests := []struct {
		name string
		env  string
	}{
		{
			name: "zero checkers",
			env:  "CHECKERS",
		},
		{
			name: "zero transfers",
			env:  "TRANSFERS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setValidEnv(t)

			t.Setenv(tt.env, "0")

			_, err := Load()
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			want := tt.env + " must be greater than zero"

			if !strings.Contains(err.Error(), want) {
				t.Fatalf(
					"unexpected error: got %q, want error containing %q",
					err.Error(),
					want,
				)
			}
		})
	}
}

func setValidEnv(t *testing.T) {
	t.Helper()

	t.Setenv("TELEGRAM_TOKEN", "test-token")
	t.Setenv("TELEGRAM_CHAT_ID", "12345")

	t.Setenv("RCLONE_ENABLED", "true")
	t.Setenv("START_BACKUP_AT_START", "false")
	t.Setenv(
		"RCLONE_PROFILES",
		"yandex_disk,ssd_toshiba",
	)
	t.Setenv(
		"BACKUP_ROOT",
		"/home/admin/backup",
	)
	t.Setenv(
		"RCLONE_CONFIG_PATH",
		"/home/admin/.config/rclone/rclone.conf",
	)
	t.Setenv(
		"RCLONE_CRON",
		"0 0 3 ? * MON,WED,SAT",
	)

	t.Setenv("STATS_INTERVAL", "5s")
	t.Setenv("CHECKERS", "8")
	t.Setenv("TRANSFERS", "4")
}
