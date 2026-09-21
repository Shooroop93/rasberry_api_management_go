package rclone

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type CopyOptions struct {
	BackupRoot    string
	StatsInterval time.Duration
	Checkers      int
	Transfers     int
}

func BuildCopyArgs(opts CopyOptions, folderName, profile string) ([]string, error) {

	if err := validateString(opts.BackupRoot, "backupRoot"); err != nil {
		return nil, err
	}

	if err := validateString(folderName, "folderName"); err != nil {
		return nil, err
	}

	if err := validateString(profile, "profile"); err != nil {
		return nil, err
	}

	if opts.Checkers <= 0 {
		return nil, errors.New("checkers must be greater than zero")
	}

	if opts.Transfers <= 0 {
		return nil, errors.New("transfers must be greater than zero")
	}

	if opts.StatsInterval <= 0 {
		return nil, errors.New("stats interval must be greater than zero")
	}

	source := filepath.Join(opts.BackupRoot, folderName)
	destination := profile + ":" + folderName

	args := []string{
		"copy",
		source,
		destination,
		"--use-json-log",
		"--log-level",
		"INFO",
		"--stats=" + opts.StatsInterval.String(),
		"--create-empty-src-dirs",
		"--checkers=" + strconv.Itoa(opts.Checkers),
		"--transfers=" + strconv.Itoa(opts.Transfers),
	}

	return args, nil
}

func validateString(value, paramName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s must not be empty", paramName)
	}

	return nil
}
