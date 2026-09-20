package config

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/pelletier/go-toml/v2"
)

// configWarningWriter receives unknown-key warnings. Tests redirect it.
var configWarningWriter io.Writer = os.Stderr

// warnedConfigFiles keeps each config file's warnings to a single emission,
// because a run loads the same file once per analysis.
var warnedConfigFiles sync.Map

// UnknownTomlKeys returns the dotted paths of keys in data that no field of
// dest accepts, so a typo such as `[clone]` instead of `[clones]` is visible
// rather than silently dropped. root navigates into a sub-table first
// (e.g. "tool", "pyscn"); keys outside it are not inspected.
func UnknownTomlKeys(data []byte, dest interface{}, root ...string) []string {
	var parsed map[string]interface{}
	if err := toml.Unmarshal(data, &parsed); err != nil {
		return nil
	}

	node := parsed
	for _, key := range root {
		child, ok := node[key].(map[string]interface{})
		if !ok {
			return nil
		}
		node = child
	}

	var unknown []string
	collectUnknownTomlKeys(node, reflect.TypeOf(dest), strings.Join(root, "."), &unknown)
	return unknown
}

// collectUnknownTomlKeys walks a decoded TOML table against the struct type
// that receives it, appending every key with no destination field.
func collectUnknownTomlKeys(node map[string]interface{}, structType reflect.Type, prefix string, unknown *[]string) {
	structType = derefType(structType)
	if structType.Kind() != reflect.Struct {
		return
	}
	if prefix != "" {
		prefix += "."
	}

	keys := make([]string, 0, len(node))
	for key := range node {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		field, ok := lookupTomlField(structType, key)
		if !ok {
			*unknown = append(*unknown, prefix+key)
			continue
		}
		switch value := node[key].(type) {
		case map[string]interface{}:
			collectUnknownTomlKeys(value, field.Type, prefix+key, unknown)
		case []interface{}:
			elemType := derefType(field.Type)
			if elemType.Kind() != reflect.Slice {
				continue
			}
			for _, elem := range value {
				table, ok := elem.(map[string]interface{})
				if !ok {
					continue
				}
				collectUnknownTomlKeys(table, elemType.Elem(), prefix+key, unknown)
			}
		}
	}
}

// lookupTomlField finds the field that go-toml would decode key into: an
// explicit `toml` tag first, then a case-insensitive field-name match.
func lookupTomlField(structType reflect.Type, key string) (reflect.StructField, bool) {
	var fallback reflect.StructField
	var hasFallback bool
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.PkgPath != "" {
			continue // unexported
		}
		tag := strings.Split(field.Tag.Get("toml"), ",")[0]
		if tag == "-" {
			continue
		}
		if tag != "" {
			if tag == key {
				return field, true
			}
			continue
		}
		if strings.EqualFold(field.Name, key) {
			fallback = field
			hasFallback = true
		}
	}
	return fallback, hasFallback
}

func derefType(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

// warnUnknownTomlKeys reports unknown keys once per config file.
func warnUnknownTomlKeys(filePath string, data []byte, dest interface{}, root ...string) {
	unknown := UnknownTomlKeys(data, dest, root...)
	if len(unknown) == 0 {
		return
	}
	if _, loaded := warnedConfigFiles.LoadOrStore(filePath+"\x00"+strings.Join(unknown, ","), struct{}{}); loaded {
		return
	}
	for _, key := range unknown {
		fmt.Fprintf(configWarningWriter, "Warning: %s: unknown configuration key %q\n", filePath, key)
	}
}
