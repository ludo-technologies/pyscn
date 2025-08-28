package main

import (
	"fmt"
	"time"
)

// generateFileName generates an automatic filename for the output
func generateFileName(command, extension string) string {
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s.%s", command, timestamp, extension)
}