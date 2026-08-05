package service

import "github.com/ludo-technologies/pyscn/internal/config"

// discoverConfigFile returns the config file governing targetPath, or "" when
// none applies.
//
// Discovery starts at the analyzed path, never at the working directory: a
// repository analyzed from outside its own tree - a nested checkout, or a
// package in a monorepo - must not be scored against the config of whatever
// project the command happened to be run from (issue #666).
func discoverConfigFile(targetPath string) string {
	tomlLoader := config.NewTomlConfigLoader()
	return tomlLoader.FindConfigFileFromPath(targetPath)
}
