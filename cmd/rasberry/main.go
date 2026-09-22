package main

import (
	"fmt"
	"os"
	"os/exec"
	"rasberry_api_management_go/internal/rclone"
	"time"
)

func main() {
	printStage()

	envs := append(
		os.Environ(),
		"RCLONE_CONFIG=/home/shooroop/.config/rclone/rclone.conf",
	)

	opts := rclone.CopyOptions{
		BackupRoot:    "/home/admin/immich-app/immich-data/library/",
		StatsInterval: 5 * time.Second,
		Checkers:      8,
		Transfers:     4,
	}

	rclonePath, err := exec.LookPath("rclone")

	if err != nil {
		fmt.Println("rclone not found", err)
		return
	}

	profiles := []string{
		"yandex_disk",
		"ssd_toshiba",
	}

	folders, err := rclone.ListFolders(opts.BackupRoot)
	if err != nil {
		fmt.Println("failed to list folders:", err)
		return
	}

	for _, folder := range folders {
		for _, profile := range profiles {
			args, err := rclone.BuildCopyArgs(opts, folder, profile)
			if err != nil {
				fmt.Printf("failed build copy args: folder = %q, profile = %q: %v\n", folder, profile, err)
				return
			}
			cmd := exec.Command(rclonePath, args...)
			cmd.Env = envs

			err = rclone.Run(cmd)
			if err != nil {
				fmt.Printf("failed to run rclone: folder = %q, profile = %q: %v\n", folder, profile, err)
				return
			}
		}
	}
}

func printStage() {
	fmt.Println("rasberry: подготовка аргументов rclone")
}
