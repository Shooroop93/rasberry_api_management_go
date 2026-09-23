package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	TelegramToken  string
	TelegramChatID string

	RcloneEnabled      bool
	StartBackupAtStart bool
	RcloneProfiles     []string
	BackupRoot         string
	RcloneConfigPath   string
	RcloneCron         string

	StatsInterval time.Duration
	Checkers      int
	Transfers     int
}

func Load() (Config, error) {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		return Config{}, errors.New("TELEGRAM_TOKEN is required")
	}

	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	if chatID == "" {
		return Config{}, errors.New("TELEGRAM_CHAT_ID is required")
	}

	rcloneEnabledRaw := os.Getenv("RCLONE_ENABLED")
	if rcloneEnabledRaw == "" {
		return Config{}, errors.New("RCLONE_ENABLED is required")
	}

	rcloneEnabled, err := strconv.ParseBool(rcloneEnabledRaw)
	if err != nil {
		return Config{}, fmt.Errorf(
			"invalid RCLONE_ENABLED: %w",
			err,
		)
	}

	startBackupAtStartRaw := os.Getenv("START_BACKUP_AT_START")
	if startBackupAtStartRaw == "" {
		return Config{}, errors.New("START_BACKUP_AT_START is required")
	}

	startBackupAtStart, err := strconv.ParseBool(startBackupAtStartRaw)
	if err != nil {
		return Config{}, fmt.Errorf(
			"invalid START_BACKUP_AT_START: %w",
			err,
		)
	}

	rcloneProfilesRaw := os.Getenv("RCLONE_PROFILES")
	if rcloneProfilesRaw == "" {
		return Config{}, errors.New("RCLONE_PROFILES is required")
	}

	rcloneProfiles := strings.Split(rcloneProfilesRaw, ",")

	for i, profile := range rcloneProfiles {
		rcloneProfiles[i] = strings.TrimSpace(profile)

		if rcloneProfiles[i] == "" {
			return Config{}, errors.New(
				"RCLONE_PROFILES contains empty profile",
			)
		}
	}

	backupRoot := os.Getenv("BACKUP_ROOT")
	if backupRoot == "" {
		return Config{}, errors.New("BACKUP_ROOT is required")
	}

	rcloneConfigPath := os.Getenv("RCLONE_CONFIG_PATH")
	if rcloneConfigPath == "" {
		return Config{}, errors.New("RCLONE_CONFIG_PATH is required")
	}

	rcloneCron := os.Getenv("RCLONE_CRON")
	if rcloneCron == "" {
		return Config{}, errors.New("RCLONE_CRON is required")
	}

	statsIntervalRaw := os.Getenv("STATS_INTERVAL")
	if statsIntervalRaw == "" {
		return Config{}, errors.New("STATS_INTERVAL is required")
	}

	statsInterval, err := time.ParseDuration(statsIntervalRaw)
	if err != nil {
		return Config{}, fmt.Errorf(
			"invalid STATS_INTERVAL: %w",
			err,
		)
	}

	if statsInterval <= 0 {
		return Config{}, errors.New(
			"STATS_INTERVAL must be greater than zero",
		)
	}

	checkersRaw := os.Getenv("CHECKERS")
	if checkersRaw == "" {
		return Config{}, errors.New("CHECKERS is required")
	}

	checkers, err := strconv.Atoi(checkersRaw)
	if err != nil {
		return Config{}, fmt.Errorf(
			"invalid CHECKERS: %w",
			err,
		)
	}

	if checkers <= 0 {
		return Config{}, errors.New(
			"CHECKERS must be greater than zero",
		)
	}

	transfersRaw := os.Getenv("TRANSFERS")
	if transfersRaw == "" {
		return Config{}, errors.New("TRANSFERS is required")
	}

	transfers, err := strconv.Atoi(transfersRaw)
	if err != nil {
		return Config{}, fmt.Errorf(
			"invalid TRANSFERS: %w",
			err,
		)
	}

	if transfers <= 0 {
		return Config{}, errors.New(
			"TRANSFERS must be greater than zero",
		)
	}

	cfg := Config{
		TelegramToken:  token,
		TelegramChatID: chatID,

		RcloneEnabled:      rcloneEnabled,
		StartBackupAtStart: startBackupAtStart,
		RcloneProfiles:     rcloneProfiles,
		BackupRoot:         backupRoot,
		RcloneConfigPath:   rcloneConfigPath,
		RcloneCron:         rcloneCron,

		StatsInterval: statsInterval,
		Checkers:      checkers,
		Transfers:     transfers,
	}

	return cfg, nil
}
