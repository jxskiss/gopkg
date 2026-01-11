package workflow

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/jxskiss/gopkg/v2/perf/json"
)

// Option configures the workflow.
type Option func(*workflow)

// WithInput sets the initial input for the workflow.
// It merges the given input into the existing input.
func WithInput(input map[string]any) Option {
	return func(w *workflow) {
		if w.initErr != nil {
			return
		}
		if w.input == nil {
			w.input = input
		} else {
			for k, v := range input {
				w.input[k] = v
			}
		}
	}
}

// WithJSONInput parses the JSON data and sets it as the initial input.
func WithJSONInput(data []byte) Option {
	return func(w *workflow) {
		if w.initErr != nil {
			return
		}
		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			w.initErr = fmt.Errorf("unmarshal JSON input: %w", err)
			return
		}
		WithInput(m)(w)
	}
}

// WithYAMLInput parses the YAML data and sets it as the initial input.
func WithYAMLInput(data []byte) Option {
	return func(w *workflow) {
		if w.initErr != nil {
			return
		}
		var m map[string]any
		if err := yaml.Unmarshal(data, &m); err != nil {
			w.initErr = fmt.Errorf("unmarshal YAML input: %w", err)
			return
		}
		WithInput(m)(w)
	}
}

// WithObserver sets the observer for the workflow.
func WithObserver(obs Observer) Option {
	return func(w *workflow) {
		if w.initErr != nil {
			return
		}
		w.observer = obs
	}
}
