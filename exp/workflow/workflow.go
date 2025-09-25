package workflow

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/jxskiss/gopkg/v2/collection/dag"
	"github.com/jxskiss/gopkg/v2/perf/gopool"
)

// Workflow is a directed acyclic graph (DAG) of tasks.
// A task is a unit of work in a workflow.
// A task can have zero or more dependencies, and a task with dependencies
// will not be executed until all its dependencies are ready.
//
// Tasks managed by a workflow executes in parallel,
// but the workflow itself is not concurrent-safe.
type Workflow struct {
	dag    *dag.DAG[string]
	runCtx *RunContext
	tasks  map[string]Task

	goFunc func(ctx context.Context, task func())
	done   chan struct{}
	errCh  chan error
	sigCh  chan string

	// shared state
	isRunning atomic.Bool
	ready     map[string]bool
}

// NewWorkflow creates a new workflow.
func NewWorkflow() *Workflow {
	return &Workflow{
		dag:    dag.NewDAG[string](),
		runCtx: NewRunContext(),
		tasks:  make(map[string]Task),
		done:   make(chan struct{}),
		errCh:  make(chan error),
		sigCh:  make(chan string),
		ready:  make(map[string]bool),
	}
}

// RunContext returns the RunContext of the workflow.
func (w *Workflow) RunContext() *RunContext {
	return w.runCtx
}

// AddTasks adds the given tasks to the workflow.
// If a task with the same ID already exists, it will be overwritten.
// If a task causes a cycle in the DAG, an error will be returned.
func (w *Workflow) AddTasks(tasks ...Task) error {
	for _, t := range tasks {
		id := t.ID()
		w.dag.AddVertex(id)
		w.tasks[id] = t
		for _, dep := range t.Depends() {
			isCyclic := w.dag.AddEdge(dep, id)
			if isCyclic {
				return fmt.Errorf("task %s causes cycle", id)
			}
		}
	}
	return nil
}

// SetReady sets the tasks with the given IDs as ready.
// This method can only be called before Run, calling it after Run
// causes a panic.
func (w *Workflow) SetReady(taskIDs ...string) {
	if w.isRunning.Load() {
		panic("SetReady can only be called before Run")
	}
	for _, id := range taskIDs {
		w.ready[id] = true
	}
}

// Run runs the workflow.
// It returns when all tasks are done or an error occurs.
//
// If the given pool is nil, the default goroutine pool will be used.
func (w *Workflow) Run(ctx context.Context, pool *gopool.Pool) error {
	var goFunc = gopool.CtxGo
	if pool != nil {
		goFunc = pool.CtxGo
	}
	w.goFunc = goFunc
	defer close(w.done)

	// prepare a cancelable context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		ready   = w.ready
		running = make(map[string]bool)
	)

	// run tasks without deps first
	w.isRunning.Store(true)
	firstTasks := w.dag.ListZeroIncomingVertices()
	for _, taskID := range firstTasks {
		task, ok := w.tasks[taskID]
		if !ok {
			return fmt.Errorf("task %s not found", taskID)
		}
		if n := len(task.Depends()); n > 0 {
			return fmt.Errorf("task %s has %d deps, expect zero", taskID, n)
		}
		running[taskID] = true
		w.runTask(ctx, task)
	}

	// wait for all tasks to run and finish
waitTasks:
	for {
		select {
		case err := <-w.errCh:
			return err
		case taskID := <-w.sigCh:
			delete(running, taskID)
			ready[taskID] = true
			w.dag.VisitNeighbors(taskID, func(next string) {
				if !ready[next] && !running[next] {
					running[next] = true
					w.runTask(ctx, w.tasks[next])
				}
			})
			if len(ready) == len(w.tasks) {
				if len(running) > 0 {
					return fmt.Errorf("workflow is in invalid state: some tasks are still running")
				}
				break waitTasks
			}
		}
	}

	return nil
}

func (w *Workflow) runTask(ctx context.Context, task Task) {
	w.goFunc(ctx, func() {
		err := task.Execute(ctx, w.runCtx)
		if err != nil {
			err = fmt.Errorf("task %s failed: %w", task.ID(), err)
			select {
			case <-w.done:
			case w.errCh <- err:
			}
		}
		// task done, notify to start next tasks
		select {
		case <-w.done:
		case w.sigCh <- task.ID():
		}
	})
}
