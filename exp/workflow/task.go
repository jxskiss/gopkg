package workflow

import (
	"context"
)

// Task is a unit of work in a workflow.
type Task interface {
	// ID returns the unique identifier of the task,
	// the returned task ID must be unique in all tasks.
	ID() string
	// Depends returns the task IDs that this task depends on.
	Depends() []string
	// Execute executes the task.
	// If a task depends on other tasks, the other tasks will be executed
	// before this task.
	// If the task execution fails, the workflow will be stopped.
	Execute(ctx context.Context, rc *RunContext) error
}

// NewFuncTask creates a new Task from the given function.
func NewFuncTask(id string, f func(ctx context.Context, rc *RunContext) error, depends ...string) Task {
	return &funcTask{
		id:      id,
		f:       f,
		depends: depends,
	}
}

type funcTask struct {
	id      string
	depends []string
	f       func(ctx context.Context, rc *RunContext) error
}

func (t *funcTask) ID() string {
	return t.id
}
func (t *funcTask) Depends() []string {
	return t.depends
}
func (t *funcTask) Execute(ctx context.Context, rc *RunContext) error {
	return t.f(ctx, rc)
}

// NewEmptyTask creates a new empty Task with the given ID.
// An empty task does nothing when executed.
func NewEmptyTask(id string) Task {
	return &emptyTask{id: id}
}

type emptyTask struct {
	id string
}

func (t *emptyTask) ID() string {
	return t.id
}
func (t *emptyTask) Depends() []string {
	return nil
}
func (t *emptyTask) Execute(_ context.Context, _ *RunContext) error {
	return nil
}
