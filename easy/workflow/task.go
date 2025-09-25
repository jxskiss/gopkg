package workflow

import (
	"context"
)

type Task interface {
	Name() string
	Depends() []string
	Execute(ctx context.Context, rc *RunContext) error
}

func NewTask(name string, f func(ctx context.Context, rc *RunContext) error, depends ...string) Task {
	return &funcTask{
		name:    name,
		f:       f,
		depends: depends,
	}
}

type funcTask struct {
	name    string
	depends []string
	f       func(ctx context.Context, rc *RunContext) error
}

func (t *funcTask) Name() string {
	return t.name
}

func (t *funcTask) Depends() []string {
	return t.depends
}

func (t *funcTask) Execute(ctx context.Context, rc *RunContext) error {
	return t.f(ctx, rc)
}
