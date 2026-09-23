package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"rasberry_api_management_go/internal/config"
	"rasberry_api_management_go/internal/rclone"
	"rasberry_api_management_go/internal/telegram"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {

	if err := godotenv.Load(); err != nil {
		fmt.Printf("failed to load .env: %v\n", err)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("config failed: %v\n", err)
		return
	}

	if !cfg.RcloneEnabled {
		fmt.Println("rclone backup is disabled")
		return
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	scheduler := cron.New(
		cron.WithSeconds(),
		cron.WithChain(
			cron.SkipIfStillRunning(cron.DefaultLogger)),
	)

	_, err = scheduler.AddFunc(
		cfg.RcloneCron,
		func() {
			if err := runBackup(cfg, client); err != nil {
				fmt.Printf("scheduled backup failed: %v\n", err)
			}
		},
	)

	if err != nil {
		fmt.Printf("failed to add cron job: %v\n", err)
		return
	}

	if cfg.StartBackupAtStart {
		if err := runBackup(cfg, client); err != nil {
			fmt.Printf("backup failed: %v\n", err)
		}
	}

	scheduler.Run()
}

func runBackup(cfg config.Config, client *http.Client) error {
	envs := append(
		os.Environ(),
		"RCLONE_CONFIG="+cfg.RcloneConfigPath,
	)

	opts := rclone.CopyOptions{
		BackupRoot:    cfg.BackupRoot,
		StatsInterval: cfg.StatsInterval,
		Checkers:      cfg.Checkers,
		Transfers:     cfg.Transfers,
	}

	rclonePath, err := exec.LookPath("rclone")
	if err != nil {
		return fmt.Errorf("rclone not found: %w", err)
	}

	folders, err := rclone.ListFolders(opts.BackupRoot)
	if err != nil {
		return fmt.Errorf("failed to list folders: %w", err)
	}

	for _, folder := range folders {
		for _, profile := range cfg.RcloneProfiles {
			args, err := rclone.BuildCopyArgs(
				opts,
				folder,
				profile,
			)
			if err != nil {
				return fmt.Errorf(
					"failed to build copy args: folder=%q, profile=%q: %w",
					folder,
					profile,
					err,
				)
			}

			cmd := exec.Command(rclonePath, args...)
			cmd.Env = envs

			var messageID int

			err = rclone.Run(cmd, func(progress string) {
				msg := rclone.FormatProgressMessage(
					folder,
					profile,
					progress,
				)

				if messageID == 0 {
					id, err := telegram.SendMessage(
						client,
						cfg.TelegramToken,
						cfg.TelegramChatID,
						msg,
					)
					if err != nil {
						fmt.Printf(
							"failed to send telegram progress message: %v\n",
							err,
						)
						return
					}

					messageID = id
					return
				}

				if err := telegram.EditMessageText(
					client,
					cfg.TelegramToken,
					cfg.TelegramChatID,
					msg,
					messageID,
				); err != nil {
					fmt.Printf(
						"failed to edit telegram progress message: %v\n",
						err,
					)
				}
			})

			if err != nil {
				return fmt.Errorf(
					"failed to run rclone: folder=%q, profile=%q: %w",
					folder,
					profile,
					err,
				)
			}
		}
	}

	return nil
}
