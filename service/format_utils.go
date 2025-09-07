package service

import (
    "encoding/json"
    "io"

    "github.com/ludo-technologies/pyscn/domain"
    "gopkg.in/yaml.v3"
)

// EncodeJSON returns an indented JSON string for the given value.
func EncodeJSON(v interface{}) (string, error) {
    data, err := json.MarshalIndent(v, "", "  ")
    if err != nil {
        return "", domain.NewOutputError("failed to marshal JSON", err)
    }
    return string(data), nil
}

// WriteJSON writes indented JSON for the given value to the writer.
func WriteJSON(w io.Writer, v interface{}) error {
    enc := json.NewEncoder(w)
    enc.SetIndent("", "  ")
    if err := enc.Encode(v); err != nil {
        return domain.NewOutputError("failed to encode JSON", err)
    }
    return nil
}

// EncodeYAML returns a YAML string for the given value.
func EncodeYAML(v interface{}) (string, error) {
    data, err := yaml.Marshal(v)
    if err != nil {
        return "", domain.NewOutputError("failed to marshal YAML", err)
    }
    return string(data), nil
}

// WriteYAML writes YAML for the given value to the writer.
func WriteYAML(w io.Writer, v interface{}) error {
    enc := yaml.NewEncoder(w)
    defer enc.Close()
    enc.SetIndent(2)
    if err := enc.Encode(v); err != nil {
        return domain.NewOutputError("failed to encode YAML", err)
    }
    return nil
}

