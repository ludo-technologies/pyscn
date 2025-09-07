package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// GetExplicitFlags extracts which flags were explicitly set from a cobra command
func GetExplicitFlags(cmd *cobra.Command) map[string]bool {
	explicitFlags := make(map[string]bool)
	if cmd != nil {
		cmd.Flags().Visit(func(f *pflag.Flag) {
			explicitFlags[f.Name] = true
		})
	}
	return explicitFlags
}