package workflow

import "github.com/jxskiss/gopkg/v2/easy/ezmap"

type RunContext struct {
	data *ezmap.SafeMap
}

func NewRunContext() *RunContext {
	return &RunContext{
		data: ezmap.NewSafeMap(),
	}
}

func (rc *RunContext) Data() *ezmap.SafeMap {
	return rc.data
}
