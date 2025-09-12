package main

import (
	"testing"
)

func TestDepsCommandInterface(t *testing.T) {
	cmd := NewDepsCmd()
	if cmd == nil {
		t.Fatal("NewDepsCmd should return a valid command")
	}
	if cmd.Use != "deps [paths...]" {
		t.Errorf("expected Use to be 'deps [paths...]', got %s", cmd.Use)
	}

	flags := cmd.Flags()
	for _, name := range []string{"json", "yaml", "csv", "dot", "html", "no-open", "config"} {
		if flags.Lookup(name) == nil {
			t.Errorf("expected flag '%s' to be defined", name)
		}
	}
}
