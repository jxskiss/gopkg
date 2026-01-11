package workflow

import (
	"context"
	"sync"

	"github.com/jxskiss/gopkg/v2/easy/ezmap"
)

type runContext struct {
	workflowID string
	input      ezmap.Map
	sharedData *ezmap.SafeMap
	outputs    sync.Map // map[string]any
	observer   Observer
}

func newRunContext(wfID string, input ezmap.Map, observer Observer) *runContext {
	if input == nil {
		input = make(ezmap.Map)
	}
	return &runContext{
		workflowID: wfID,
		input:      input,
		sharedData: ezmap.NewSafeMap(),
		observer:   observer,
	}
}

func (c *runContext) WorkflowID() string {
	return c.workflowID
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

func (c *runContext) EmitEvent(ctx context.Context, taskID, eventName string, data any) {
	if c.observer != nil {
		c.observer.OnEvent(ctx, taskID, eventName, data)
	}
}

func (c *runContext) setTaskOutput(taskID string, output any) {
	c.outputs.Store(taskID, output)
}
