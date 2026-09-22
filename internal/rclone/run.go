package rclone

import (
	"bufio"
	"fmt"
	"os/exec"
)

func Run(cmd *exec.Cmd) error {

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return fmt.Errorf("stdout pipe failed: %w", err)
	}

	cmd.Stderr = cmd.Stdout

	scanner := bufio.NewScanner(stdout)

	err = cmd.Start()

	if err != nil {
		return fmt.Errorf("start failed: %w", err)
	}

	var parseErr error

	for scanner.Scan() {
		line := scanner.Text()
		msg, err := ParseLogLine(line)

		if err != nil {
			parseErr = err
			break
		}

		transferred := findTransferred(msg)

		if transferred != "" {
			fmt.Println(transferred)
		}
	}

	scanErr := scanner.Err()

	if scanErr != nil || parseErr != nil {
		_ = cmd.Process.Kill()
	}

	waitErr := cmd.Wait()

	if scanErr != nil {
		return fmt.Errorf("read rclone output: %w", scanErr)
	}

	if parseErr != nil {
		return fmt.Errorf("parse rclone output: %w", parseErr)
	}

	if waitErr != nil {
		return fmt.Errorf("wait rclone process: %w", waitErr)
	}

	return nil
}
