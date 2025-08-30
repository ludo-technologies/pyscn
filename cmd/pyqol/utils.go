package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/pyqol/pyqol/internal/config"
)


// generateFileNameWithTarget generates an automatic filename with target path context
func generateFileNameWithTarget(command, extension string, targetPath string) string {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.%s", command, timestamp, extension)
	
	// Priority:
	// 1. Config file output.directory setting (search from target path)
	// 2. Current directory (default)
	
	// Try to load output directory from config file
	cfg, err := config.LoadConfigWithTarget("", targetPath)
	if err == nil && cfg != nil && cfg.Output.Directory != "" {
		return filepath.Join(cfg.Output.Directory, filename)
	}
	
	// Default to current directory
	return filename
}

