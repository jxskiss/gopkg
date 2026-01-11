package workflow

import (
	"context"
	"time"

	"github.com/jxskiss/gopkg/v2/easy/ezmap"
	"github.com/jxskiss/gopkg/v2/perf/gopool"
)

// TaskResult 任务执行结果
type TaskResult struct {
	ID        string
	StartTime time.Time
	EndTime   time.Time
	Output    any
	Error     error
}

// Task is a unit of work in a Workflow.
type Task interface {
	// ID 任务唯一标识
	ID() string
	// Depends 依赖的任务列表
	Depends() []string
	// Run 执行任务
	Run(ctx context.Context, rc RunContext) (output any, err error)
}

// WorkflowResult 工作流执行结果
//
//nolint:revive
type WorkflowResult struct {
	ID          string
	StartTime   time.Time
	EndTime     time.Time
	TaskResults map[string]*TaskResult
	Output      any
	Error       error
}

type Workflow interface {
	ID() string
	Tasks() []Task
	AddTask(ctx context.Context, tasks ...Task) error
	Execute(ctx context.Context, gp *gopool.GoPool) (*WorkflowResult, error)
	RunContext() RunContext
}

type Observer interface {
	OnEvent(ctx context.Context, taskID, eventName string, data any)
	OnTaskStarted(ctx context.Context, taskID string)
	OnTaskCompleted(ctx context.Context, taskID string, result *TaskResult)
	OnWorkflowStarted(ctx context.Context, workflowID string)
	OnWorkflowCompleted(ctx context.Context, workflowID string, result *WorkflowResult)
}

// RunContext 贯穿 Workflow 整个执行过程，提供 Workflow 上下文信息和跨任务的协作能力
type RunContext interface {
	WorkflowID() string
	WorkflowInput() ezmap.Map
	GetTaskOutput(taskID string) (any, bool)
	SharedData() *ezmap.SafeMap
	AddTask(ctx context.Context, tasks ...Task) error
	EmitEvent(ctx context.Context, taskID, eventName string, data any)
}
