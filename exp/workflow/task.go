package workflow

import "context"

// Action 定义任务的执行逻辑。
type Action func(ctx context.Context, rc RunContext) (any, error)

// BaseTask 是 Task 接口的一个基础实现。
type BaseTask struct {
	id      string
	depends []string
	action  Action
}

// NewTask 创建一个新的任务。
func NewTask(id string, action Action, options ...TaskOption) *BaseTask {
	t := &BaseTask{
		id:     id,
		action: action,
	}
	for _, opt := range options {
		opt(t)
	}
	return t
}

// ID 返回任务的唯一标识。
func (t *BaseTask) ID() string {
	return t.id
}

// Depends 返回任务的依赖列表。
func (t *BaseTask) Depends() []string {
	return t.depends
}

// Run 执行任务。
func (t *BaseTask) Run(ctx context.Context, rc RunContext) (any, error) {
	if t.action == nil {
		return nil, nil
	}
	return t.action(ctx, rc)
}

// TaskOption 配置任务。
type TaskOption func(*BaseTask)

// WithDepends 设置任务的依赖。
func WithDepends(depends ...string) TaskOption {
	return func(t *BaseTask) {
		t.depends = append(t.depends, depends...)
	}
}

// DependsOn 是 WithDepends 的别名，更加语义化。
func DependsOn(tasks ...string) TaskOption {
	return WithDepends(tasks...)
}
