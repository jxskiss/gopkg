package workflow

import (
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/easy/ezmap"
)

func TestWorkflow(t *testing.T) {
	t.Run("cyclic graph", func(t *testing.T) {
		w := NewWorkflow()
		err := w.AddTasks(
			NewFuncTask("1", func(ctx context.Context, rc *RunContext) error {
				return nil
			}, "2"),
			NewFuncTask("2", func(ctx context.Context, rc *RunContext) error {
				return nil
			}, "3"),
			NewFuncTask("3", func(ctx context.Context, rc *RunContext) error {
				return nil
			}, "1"),
		)
		assert.Error(t, err)
	})

	t.Run("simple", func(t *testing.T) {
		addState := func(rc *RunContext, value string) {
			rc.Data().WithLock(func(m ezmap.Map) {
				state := m.GetOr("state", []string{}).([]string)
				state = append(state, value)
				m.Set("state", state)
			})
		}

		w := NewWorkflow()
		err := w.AddTasks(
			NewFuncTask("1", func(ctx context.Context, rc *RunContext) error {
				addState(rc, "1")
				log.Printf("task 1: %v", rc.Data().MustGet("state"))
				return nil
			}, "2", "3"),
			NewFuncTask("2", func(ctx context.Context, rc *RunContext) error {
				addState(rc, "2")
				log.Printf("task 2: %v", rc.Data().MustGet("state"))
				return nil
			}, "3", "4"),
			NewFuncTask("3", func(ctx context.Context, rc *RunContext) error {
				addState(rc, "3")
				log.Printf("task 3: %v", rc.Data().MustGet("state"))
				return nil
			}, "4"),
			NewFuncTask("4", func(ctx context.Context, rc *RunContext) error {
				addState(rc, "4")
				log.Printf("task 4: %v", rc.Data().MustGet("state"))
				return nil
			}, "5"),
			NewFuncTask("5", func(ctx context.Context, rc *RunContext) error {
				addState(rc, "5")
				log.Printf("task 5: %v", rc.Data().MustGet("state"))
				return nil
			}, "6", "7"),
			NewFuncTask("6", func(ctx context.Context, rc *RunContext) error {
				addState(rc, "6")
				log.Printf("task 6: %v", rc.Data().MustGet("state"))
				return nil
			}, "7", "8"),
			NewFuncTask("7", func(ctx context.Context, rc *RunContext) error {
				addState(rc, "7")
				log.Printf("task 7: %v", rc.Data().MustGet("state"))
				return nil
			}, "8"),
			NewEmptyTask("8"),
		)
		assert.NoError(t, err)
		w.SetReady("8")

		err = w.Run(context.Background(), nil)
		assert.NoError(t, err)

		state := w.RunContext().Data().MustGet("state").([]string)
		assert.Equal(t, []string{"7", "6", "5", "4", "3", "2", "1"}, state)
	})
}
