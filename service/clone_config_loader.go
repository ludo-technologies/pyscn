package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
	"github.com/ludo-technologies/pyscn/internal/constants"
)

// CloneConfigurationLoader implements the domain.CloneConfigurationLoader interface
type CloneConfigurationLoader struct{}

// NewCloneConfigurationLoader creates a new clone configuration loader
func NewCloneConfigurationLoader() *CloneConfigurationLoader {
	return &CloneConfigurationLoader{}
}

// LoadCloneConfig loads clone detection configuration from file
func (c *CloneConfigurationLoader) LoadCloneConfig(configPath string) (*domain.CloneRequest, error) {
	// Load the full configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Convert to clone request
	request := c.configToCloneRequest(&cfg.CloneDetection)
	return request, nil
}

// SaveCloneConfig saves clone detection configuration to file
func (c *CloneConfigurationLoader) SaveCloneConfig(cloneConfig *domain.CloneRequest, configPath string) error {
	// Load existing config or create new one
	var cfg *config.Config
	if _, err := os.Stat(configPath); err == nil {
		// File exists, load it
		loadedCfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load existing config: %w", err)
		}
		cfg = loadedCfg
	} else {
		// File doesn't exist, create default config
		cfg = config.DefaultConfig()
	}

	// Update clone detection section
	c.updateConfigFromCloneRequest(cfg, cloneConfig)

	// Save the updated configuration
	return config.SaveConfig(cfg, configPath)
}

// GetDefaultCloneConfig returns default clone detection configuration, first checking for .pyscn.yaml
func (c *CloneConfigurationLoader) GetDefaultCloneConfig() *domain.CloneRequest {
	// First, try to find and load a config file in the current directory
	configFile := c.FindDefaultConfigFile()
	if configFile != "" {
		if configReq, err := c.LoadCloneConfig(configFile); err == nil {
			return configReq
		}
		// If loading failed, fall back to hardcoded defaults
	}

	// Fall back to hardcoded default configuration
	defaultConfig := config.DefaultConfig()
	return c.configToCloneRequest(&defaultConfig.CloneDetection)
}

// configToCloneRequest converts a config.CloneDetectionConfig to domain.CloneRequest
func (c *CloneConfigurationLoader) configToCloneRequest(cfg *config.CloneDetectionConfig) *domain.CloneRequest {
	// Convert string clone types to domain clone types
	cloneTypes := make([]domain.CloneType, 0, len(cfg.CloneTypes))
	for _, typeStr := range cfg.CloneTypes {
		switch typeStr {
		case "type1":
			cloneTypes = append(cloneTypes, domain.Type1Clone)
		case "type2":
			cloneTypes = append(cloneTypes, domain.Type2Clone)
		case "type3":
			cloneTypes = append(cloneTypes, domain.Type3Clone)
		case "type4":
			cloneTypes = append(cloneTypes, domain.Type4Clone)
		}
	}

	// Convert sort criteria
	var sortBy domain.SortCriteria
	switch cfg.SortBy {
	case "similarity":
		sortBy = domain.SortBySimilarity
	case "size":
		sortBy = domain.SortBySize
	case "location":
		sortBy = domain.SortByLocation
	case "type":
		sortBy = domain.SortByComplexity // Reuse existing sort criteria
	default:
		sortBy = domain.SortBySimilarity
	}

	return &domain.CloneRequest{
		Paths:               []string{"."}, // Default to current directory
		MinLines:            cfg.MinLines,
		MinNodes:            cfg.MinNodes,
		SimilarityThreshold: cfg.SimilarityThreshold,
		MaxEditDistance:     cfg.MaxEditDistance,
		IgnoreLiterals:      cfg.IgnoreLiterals,
		IgnoreIdentifiers:   cfg.IgnoreIdentifiers,
		Type1Threshold:      cfg.Type1Threshold,
		Type2Threshold:      cfg.Type2Threshold,
		Type3Threshold:      cfg.Type3Threshold,
		Type4Threshold:      cfg.Type4Threshold,
		ShowDetails:         false, // Not stored in config, use CLI default
		ShowContent:         cfg.ShowContent,
		SortBy:              sortBy,
		GroupClones:         cfg.GroupClones,
		MinSimilarity:       cfg.MinSimilarity,
		MaxSimilarity:       cfg.MaxSimilarity,
		CloneTypes:          cloneTypes,
		OutputFormat:        domain.OutputFormatText,            // Default, overridden by CLI
		Recursive:           true,                               // Default, overridden by CLI
		IncludePatterns:     []string{"*.py"},                   // Default, overridden by CLI
		ExcludePatterns:     []string{"test_*.py", "*_test.py"}, // Default
	}
}

// updateConfigFromCloneRequest updates a config.Config from a domain.CloneRequest
func (c *CloneConfigurationLoader) updateConfigFromCloneRequest(cfg *config.Config, req *domain.CloneRequest) {
	// Convert domain clone types to string slice
	cloneTypes := make([]string, 0, len(req.CloneTypes))
	for _, cloneType := range req.CloneTypes {
		switch cloneType {
		case domain.Type1Clone:
			cloneTypes = append(cloneTypes, "type1")
		case domain.Type2Clone:
			cloneTypes = append(cloneTypes, "type2")
		case domain.Type3Clone:
			cloneTypes = append(cloneTypes, "type3")
		case domain.Type4Clone:
			cloneTypes = append(cloneTypes, "type4")
		}
	}

	// Convert sort criteria to string
	var sortBy string
	switch req.SortBy {
	case domain.SortBySimilarity:
		sortBy = "similarity"
	case domain.SortBySize:
		sortBy = "size"
	case domain.SortByLocation:
		sortBy = "location"
	default:
		sortBy = "similarity"
	}

	// Update clone detection configuration
	cfg.CloneDetection = config.CloneDetectionConfig{
		Enabled:             true, // Always enabled if we're saving config
		MinLines:            req.MinLines,
		MinNodes:            req.MinNodes,
		Type1Threshold:      req.Type1Threshold,
		Type2Threshold:      req.Type2Threshold,
		Type3Threshold:      req.Type3Threshold,
		Type4Threshold:      req.Type4Threshold,
		SimilarityThreshold: req.SimilarityThreshold,
		MaxEditDistance:     req.MaxEditDistance,
		CostModelType:       "python", // Default cost model
		IgnoreLiterals:      req.IgnoreLiterals,
		IgnoreIdentifiers:   req.IgnoreIdentifiers,
		ShowContent:         req.ShowContent,
		GroupClones:         req.GroupClones,
		SortBy:              sortBy,
		MinSimilarity:       req.MinSimilarity,
		MaxSimilarity:       req.MaxSimilarity,
		CloneTypes:          cloneTypes,
	}

	// Update analysis patterns if provided
	if len(req.IncludePatterns) > 0 {
		cfg.Analysis.IncludePatterns = req.IncludePatterns
	}
	if len(req.ExcludePatterns) > 0 {
		cfg.Analysis.ExcludePatterns = req.ExcludePatterns
	}
	cfg.Analysis.Recursive = req.Recursive
}

// LoadCloneConfigFromViper loads clone configuration using viper (for advanced config scenarios)
func (c *CloneConfigurationLoader) LoadCloneConfigFromViper(configPath string) (*domain.CloneRequest, error) {
	viper.SetConfigFile(configPath)

	// Set defaults
	c.setViperDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse clone detection section
	request := &domain.CloneRequest{}
	if err := viper.UnmarshalKey("clone_detection", request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal clone_detection config: %w", err)
	}

	return request, nil
}

// setViperDefaults sets default values in viper
func (c *CloneConfigurationLoader) setViperDefaults() {
	viper.SetDefault("clone_detection.enabled", true)
	viper.SetDefault("clone_detection.min_lines", 5)
	viper.SetDefault("clone_detection.min_nodes", 10)
	viper.SetDefault("clone_detection.type1_threshold", constants.DefaultType1CloneThreshold)
	viper.SetDefault("clone_detection.type2_threshold", constants.DefaultType2CloneThreshold)
	viper.SetDefault("clone_detection.type3_threshold", constants.DefaultType3CloneThreshold)
	viper.SetDefault("clone_detection.type4_threshold", constants.DefaultType4CloneThreshold)
	viper.SetDefault("clone_detection.similarity_threshold", 0.80)
	viper.SetDefault("clone_detection.max_edit_distance", 50.0)
	viper.SetDefault("clone_detection.cost_model_type", "python")
	viper.SetDefault("clone_detection.ignore_literals", false)
	viper.SetDefault("clone_detection.ignore_identifiers", false)
	viper.SetDefault("clone_detection.show_content", false)
	viper.SetDefault("clone_detection.group_clones", true)
	viper.SetDefault("clone_detection.sort_by", "similarity")
	viper.SetDefault("clone_detection.min_similarity", 0.0)
	viper.SetDefault("clone_detection.max_similarity", 1.0)
	viper.SetDefault("clone_detection.clone_types", []string{"type1", "type2", "type3", "type4"})
}

// SaveCloneConfigAsYAML saves clone configuration as a standalone YAML file
func (c *CloneConfigurationLoader) SaveCloneConfigAsYAML(cloneConfig *domain.CloneRequest, filePath string) error {
	// Create a simplified config structure for YAML output
	yamlConfig := map[string]interface{}{
		"clone_detection": map[string]interface{}{
			"enabled":              true,
			"min_lines":            cloneConfig.MinLines,
			"min_nodes":            cloneConfig.MinNodes,
			"type1_threshold":      cloneConfig.Type1Threshold,
			"type2_threshold":      cloneConfig.Type2Threshold,
			"type3_threshold":      cloneConfig.Type3Threshold,
			"type4_threshold":      cloneConfig.Type4Threshold,
			"similarity_threshold": cloneConfig.SimilarityThreshold,
			"max_edit_distance":    cloneConfig.MaxEditDistance,
			"cost_model_type":      "python",
			"ignore_literals":      cloneConfig.IgnoreLiterals,
			"ignore_identifiers":   cloneConfig.IgnoreIdentifiers,
			"show_content":         cloneConfig.ShowContent,
			"group_clones":         cloneConfig.GroupClones,
			"sort_by":              string(cloneConfig.SortBy),
			"min_similarity":       cloneConfig.MinSimilarity,
			"max_similarity":       cloneConfig.MaxSimilarity,
			"clone_types":          c.cloneTypesToStrings(cloneConfig.CloneTypes),
		},
		"analysis": map[string]interface{}{
			"include_patterns": cloneConfig.IncludePatterns,
			"exclude_patterns": cloneConfig.ExcludePatterns,
			"recursive":        cloneConfig.Recursive,
		},
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write YAML file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()
	encoder.SetIndent(2)

	if err := encoder.Encode(yamlConfig); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}

	return nil
}

// cloneTypesToStrings converts clone types to string slice
func (c *CloneConfigurationLoader) cloneTypesToStrings(types []domain.CloneType) []string {
	strings := make([]string, len(types))
	for i, cloneType := range types {
		switch cloneType {
		case domain.Type1Clone:
			strings[i] = "type1"
		case domain.Type2Clone:
			strings[i] = "type2"
		case domain.Type3Clone:
			strings[i] = "type3"
		case domain.Type4Clone:
			strings[i] = "type4"
		default:
			strings[i] = "type1"
		}
	}
	return strings
}

// FindDefaultConfigFile looks for .pyscn.yaml in the current directory
func (c *CloneConfigurationLoader) FindDefaultConfigFile() string {
	configFiles := []string{".pyscn.yaml", ".pyscn.yml", "pyscn.yaml"}

	for _, filename := range configFiles {
		if _, err := os.Stat(filename); err == nil {
			return filename
		}
	}

	return "" // No config file found
}
