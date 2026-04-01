package mcp

import (
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
)

func NewTestDependencies(fr domain.FileReader, cfg *config.Config, path string) *Dependencies {
	return &Dependencies{
		fileReader: fr,
		config:     cfg,
		configPath: path,
	}
}
