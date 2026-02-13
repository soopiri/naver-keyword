package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func getLogPath(filename string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filename, err
	}
	logDir := filepath.Join(homeDir, ".pick-keyword", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return filename, err
	}
	return filepath.Join(logDir, filename), nil
}

func logLine(filename, message string) {
	logPath, err := getLogPath(filename)
	if err != nil {
		return
	}

	line := fmt.Sprintf("%s %s\n", time.Now().Format(time.RFC3339), message)
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	file.WriteString(line)
}

func readLogTail(filename string, maxChars int) (string, error) {
	if maxChars == 0 {
		maxChars = 8000
	}

	logPath, err := getLogPath(filename)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	if len(data) <= maxChars {
		return string(data), nil
	}

	return string(data[len(data)-maxChars:]), nil
}
