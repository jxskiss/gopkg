package workflow

import (
	"context"
	"fmt"

	"github.com/jxskiss/gopkg/v2/collection/dag"
	"github.com/jxskiss/gopkg/v2/perf/gopool"
)

type Workflow struct {
	dag    *dag.DAG[string]
	runCtx *RunContext
	tasks  map[string]Task

	goFunc func(ctx context.Context, task func())
	done   chan struct{}
	errCh  chan error
	sigCh  chan string
}

func NewWorkflow() *Workflow {
	return &Workflow{
		dag:    dag.NewDAG[string](),
		runCtx: NewRunContext(),
		tasks:  make(map[string]Task),
		done:   make(chan struct{}),
		errCh:  make(chan error),
		sigCh:  make(chan string),
	}
}

func (w *Workflow) AddTasks(tasks ...Task) error {
	for _, t := range tasks {
		name := t.Name()
		w.dag.AddVertex(name)
		w.tasks[name] = t
		for _, dep := range t.Depends() {
			isCyclic := w.dag.AddEdge(dep, name)
			if isCyclic {
				return fmt.Errorf("tasks are cyclic: %s", name)
			}
		}
	}
	return nil
}

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
		ready   = make(map[string]bool)
		running = make(map[string]bool)
	)

	// run tasks without deps first
	firstTasks := w.dag.ListZeroIncomingVertices()
	for _, taskName := range firstTasks {
		task, ok := w.tasks[taskName]
		if !ok {
			return fmt.Errorf("task %s not found", taskName)
		}
		if n := len(task.Depends()); n > 0 {
			return fmt.Errorf("task %s has %d deps, expect zero", taskName, n)
		}
		running[taskName] = true
		w.runTask(ctx, task)
	}

	// wait for all tasks to run and finish
	for {
		select {
		case err := <-w.errCh:
			return err
		case taskName := <-w.sigCh:
			delete(running, taskName)
			ready[taskName] = true
			w.dag.VisitNeighbors(taskName, func(next string) {
				if !ready[next] && !running[next] {
					running[next] = true
					w.runTask(ctx, w.tasks[next])
				}
			})
			if len(ready) == len(w.tasks) {
				if len(running) > 0 {
					panic("workflow is in invalid state: some tasks are still running")
				}
				return nil
			}
		}
	}
}

func (w *Workflow) runTask(ctx context.Context, task Task) {
	w.goFunc(ctx, func() {
		err := task.Execute(ctx, w.runCtx)
		if err != nil {
			err = fmt.Errorf("task %s failed: %w", task.Name(), err)
			select {
			case <-w.done:
			case w.errCh <- err:
			}
		}
		// task done, notify to start next tasks
		select {
		case <-w.done:
		case w.sigCh <- task.Name():
		}
	})
}
