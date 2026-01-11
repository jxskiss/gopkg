package workflow

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/jxskiss/gopkg/v2/collection/dag"
	"github.com/jxskiss/gopkg/v2/easy/ezmap"
	"github.com/jxskiss/gopkg/v2/perf/gopool"
)

type workflow struct {
	id    string
	tasks map[string]Task
	dag   *dag.DAG[string]

	observer Observer
	input    ezmap.Map

	runCtx  *runContext
	initErr error
}

// NewWorkflow creates a new workflow instance.
func NewWorkflow(id string, opts ...Option) Workflow {
	w := &workflow{
		id:    id,
		tasks: make(map[string]Task),
		dag:   dag.New[string](),
	}
	for _, opt := range opts {
		opt(w)
	}
	w.runCtx = newRunContext(id, w.input, w.observer)
	return w
}

func (w *workflow) ID() string { return w.id }

func (w *workflow) Tasks() []Task {
	tasks := make([]Task, 0, len(w.tasks))
	taskIDs := w.dag.TopoSort()
	for _, id := range taskIDs {
		tasks = append(tasks, w.tasks[id])
	}
	return tasks
}

func (w *workflow) AddTask(_ context.Context, tasks ...Task) error {
	for _, t := range tasks {
		id := t.ID()
		if id == "" {
			return errors.New("task ID cannot be empty")
		}
		if _, exists := w.tasks[id]; exists {
			return fmt.Errorf("task %s already exists", id)
		}

		w.tasks[id] = t
		w.dag.AddVertex(id)

		for _, dep := range t.Depends() {
			if dep == id {
				return fmt.Errorf("task %s depends on itself", id)
			}
			if w.dag.AddEdge(dep, id) {
				return fmt.Errorf("cycle detected: task %s depends on %s", id, dep)
			}
		}
	}
	return nil
}

func (w *workflow) RunContext() RunContext {
	return w.runCtx
}

func (w *workflow) Execute(ctx context.Context, gp *gopool.GoPool) (*WorkflowResult, error) {
	if w.initErr != nil {
		return nil, w.initErr
	}

	// Validate all dependencies exist
	for id, t := range w.tasks {
		for _, dep := range t.Depends() {
			if _, ok := w.tasks[dep]; !ok {
				return nil, fmt.Errorf("task %s depends on non-existent task %s", id, dep)
			}
		}
	}

	// Update context with current observer/input if they changed (though currently no setter exposed)
	// We reuse the existing runCtx instance to maintain state visibility via RunContext() method.
	w.runCtx.observer = w.observer
	if w.input != nil {
		w.runCtx.input = w.input
	}
	runCtx := w.runCtx

	if w.observer != nil {
		w.observer.OnWorkflowStarted(ctx, w.id)
	}

	// Create a cancelable context to control task execution
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	result := &WorkflowResult{
		ID:          w.id,
		StartTime:   time.Now(),
		TaskResults: make(map[string]*TaskResult),
	}

	completed := make(map[string]bool)
	running := make(map[string]bool)

	type taskDoneEvent struct {
		taskID string
		res    *TaskResult
	}
	doneCh := make(chan taskDoneEvent, len(w.tasks))

	var goFunc = gopool.CtxGo
	if gp != nil {
		goFunc = gp.CtxGo
	}

	startTask := func(t Task) {
		running[t.ID()] = true
		if w.observer != nil {
			w.observer.OnTaskStarted(ctx, t.ID())
		}

		goFunc(ctx, func() {
			var tr *TaskResult
			var start time.Time
			defer func() {
				// If tr is nil, it means the task panicked.
				// We don't recover the panic here, let gopool handle it.
				// But we must notify the channel to avoid deadlock.
				if tr == nil {
					err := fmt.Errorf("task %s panicked", t.ID())
					tr = &TaskResult{
						ID:        t.ID(),
						StartTime: start,
						EndTime:   time.Now(),
						Error:     err,
					}
					doneCh <- taskDoneEvent{taskID: t.ID(), res: tr}
				} else {
					doneCh <- taskDoneEvent{taskID: t.ID(), res: tr}
				}
			}()

			start = time.Now()
			out, err := t.Run(ctx, runCtx)
			end := time.Now()

			tr = &TaskResult{
				ID:        t.ID(),
				StartTime: start,
				EndTime:   end,
				Output:    out,
				Error:     err,
			}
			if err == nil {
				runCtx.setTaskOutput(t.ID(), out)
			}
		})
	}

	// Start initial tasks (zero incoming edges)
	zeroIn := w.dag.ListZeroIncomingVertices()
	slices.Sort(zeroIn) // Ensure deterministic order

	startedCount := 0
	for _, id := range zeroIn {
		if t, ok := w.tasks[id]; ok {
			startTask(t)
			startedCount++
		}
	}

	if startedCount == 0 && len(w.tasks) > 0 {
		// Should not happen if cycle check passed and deps exist
		return nil, errors.New("no tasks can be started (possible cycle or logic error)")
	}

	if len(w.tasks) == 0 {
		goto Finish
	}

	for len(completed) < len(w.tasks) {
		select {
		case <-ctx.Done():
			result.Error = ctx.Err()
			goto Finish
		case evt := <-doneCh:
			tr := evt.res
			result.TaskResults[tr.ID] = tr
			completed[tr.ID] = true
			delete(running, tr.ID)

			if w.observer != nil {
				w.observer.OnTaskCompleted(ctx, tr.ID, tr)
			}

			if tr.Error != nil {
				result.Error = tr.Error
				cancel() // Cancel other running tasks
				goto Finish
			}

			// Check and start next tasks
			w.dag.VisitNeighbors(tr.ID, func(nextID string) {
				if completed[nextID] || running[nextID] {
					return
				}

				allDepsDone := true
				w.dag.VisitReverseNeighbors(nextID, func(depID string) {
					if !completed[depID] {
						allDepsDone = false
					}
				})

				if allDepsDone {
					if nextTask, ok := w.tasks[nextID]; ok {
						startTask(nextTask)
					}
				}
			})
		}
	}

Finish:
	result.EndTime = time.Now()

	if w.observer != nil {
		w.observer.OnWorkflowCompleted(ctx, w.id, result)
	}

	return result, result.Error
}
