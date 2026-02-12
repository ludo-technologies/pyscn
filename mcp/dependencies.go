package mcp

import (
	"github.com/ludo-technologies/pyscn/app"
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
	"github.com/ludo-technologies/pyscn/service"
)

// Dependencies aggregates the shared services required by MCP handlers.
type Dependencies struct {
	fileReader domain.FileReader
	config     *config.Config
	configPath string
}

// NewDependencies constructs the dependency set with sane defaults.
func NewDependencies(cfg *config.Config, configPath string) *Dependencies {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	return &Dependencies{
		fileReader: service.NewFileReader(),
		config:     cfg,
		configPath: configPath,
	}
}

// Config exposes the loaded configuration snapshot.
func (d *Dependencies) Config() *config.Config {
	return d.config
}

// ConfigPath returns the configured config file path (may be empty to trigger discovery).
func (d *Dependencies) ConfigPath() string {
	return d.configPath
}

// BuildAnalyzeUseCase assembles a fresh AnalyzeUseCase with injected dependencies.
func (d *Dependencies) BuildAnalyzeUseCase() (*app.AnalyzeUseCase, error) {
	return buildAnalyzeUseCase(d.fileReader)
}
