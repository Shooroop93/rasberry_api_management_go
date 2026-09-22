package rclone

import (
	"encoding/json"
	"fmt"
	"strings"
)

type logEntry struct {
	Msg string `json:"msg"`
}

func ParseLogLine(line string) (string, error) {

	var entry logEntry

	err := json.Unmarshal([]byte(line), &entry)
	if err != nil {
		return "", fmt.Errorf("failed unmarshal json: %w", err)
	}

	return entry.Msg, nil
}

func findTransferred(msg string) string {
	splitMsg := strings.Split(msg, "\n")

	for _, line := range splitMsg {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Transferred:") {
			return line
		}
	}

	return ""
}
