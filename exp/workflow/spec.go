package workflow

import "fmt"

// TaskSpec defines one task in declarative workflow spec.
type TaskSpec struct {
	Name      string
	DependsOn []string
	Action    TaskFunc
	Params    any
}

// Spec is a declarative definition for building a workflow.
type Spec struct {
	Name  string
	Tasks []TaskSpec
}

// Build creates a workflow from declarative spec.
func Build(spec Spec) (*Workflow, error) {
	w := New(spec.Name)
	for _, t := range spec.Tasks {
		if err := w.AddTask(t.Name, t.Action, t.Params); err != nil {
			return nil, fmt.Errorf("build spec task %s: %w", t.Name, err)
		}
	}
	for _, t := range spec.Tasks {
		if len(t.DependsOn) == 0 {
			continue
		}
		if err := w.DependsOn(t.Name, t.DependsOn...); err != nil {
			return nil, fmt.Errorf("build spec deps for task %s: %w", t.Name, err)
		}
	}
	return w, nil
}
