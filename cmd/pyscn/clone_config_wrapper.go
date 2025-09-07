package main

import (
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/service"
	"github.com/spf13/cobra"
)

// CloneConfigWrapper wraps clone configuration loading with explicit flag tracking
type CloneConfigWrapper struct {
	loader  *service.CloneConfigurationLoaderWithFlags
	request *domain.CloneRequest
}

// NewCloneConfigWrapper creates a new clone configuration wrapper
func NewCloneConfigWrapper(cmd *cobra.Command, request *domain.CloneRequest) *CloneConfigWrapper {
	// Track which flags were explicitly set by the user
	explicitFlags := GetExplicitFlags(cmd)

	return &CloneConfigWrapper{
		loader:  service.NewCloneConfigurationLoaderWithFlags(explicitFlags),
		request: request,
	}
}

// MergeWithConfig merges the request with configuration from file
func (w *CloneConfigWrapper) MergeWithConfig() *domain.CloneRequest {
	if w.request.ConfigPath == "" {
		// Try to load default config
		defaultConfig := w.loader.GetDefaultCloneConfig()
		if defaultConfig != nil {
			return w.loader.MergeConfig(defaultConfig, w.request)
		}
		return w.request
	}

	// Load specified config
	configReq, err := w.loader.LoadCloneConfig(w.request.ConfigPath)
	if err != nil {
		// If config loading fails, return original request
		return w.request
	}

	return w.loader.MergeConfig(configReq, w.request)
}