package main

import (
	"fmt"
	"os/exec"
	"rasberry_api_management_go/internal/rclone"
	"time"
)

func main() {
	printStage()

	opts := rclone.CopyOptions{
		BackupRoot:    "/home/admin/immich-app/immich-data/library/",
		StatsInterval: 5 * time.Second,
		Checkers:      8,
		Transfers:     4,
	}

	args, err := rclone.BuildCopyArgs(opts, "alice", "yandex_disk")

	if err != nil {
		fmt.Println("error create BuildCopyArgs", err)
		return
	}

	path, err := exec.LookPath("rclone")

	if err != nil {
		fmt.Println("rclone not found", err)
		return
	}

	cmd := exec.Command("rclone", args...)

	fmt.Printf("%q\n", cmd.Args)
	fmt.Println(path)

}

func printStage() {
	fmt.Println("rasberry: подготовка аргументов rclone")
}
