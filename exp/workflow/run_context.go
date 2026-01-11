package workflow

import (
	"context"
	"sync"

	"github.com/jxskiss/gopkg/v2/easy/ezmap"
)

type runContext struct {
	wf         Workflow
	input      ezmap.Map
	sharedData *ezmap.SafeMap
	outputs    sync.Map // map[string]any
	observer   Observer
}

func newRunContext(wf Workflow, input ezmap.Map, observer Observer) *runContext {
	if input == nil {
		input = make(ezmap.Map)
	}
	return &runContext{
		wf:         wf,
		input:      input,
		sharedData: ezmap.NewSafeMap(),
		observer:   observer,
	}
}

func (c *runContext) WorkflowID() string {
	return c.wf.ID()
}

func (c *runContext) WorkflowInput() ezmap.Map {
	return c.input
}

func (c *runContext) GetTaskOutput(taskID string) (any, bool) {
	return c.outputs.Load(taskID)
}

func (c *runContext) SharedData() *ezmap.SafeMap {
	return c.sharedData
}

func (c *runContext) AddTask(ctx context.Context, tasks ...Task) error {
	return c.wf.AddTask(ctx, tasks...)
}

func (c *runContext) EmitEvent(ctx context.Context, taskID, eventName string, data any) {
	if c.observer != nil {
		c.observer.OnEvent(ctx, taskID, eventName, data)
	}
}

func (c *runContext) setTaskOutput(taskID string, output any) {
	c.outputs.Store(taskID, output)
}
