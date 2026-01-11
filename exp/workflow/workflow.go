package workflow

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/jxskiss/gopkg/v2/collection/dag"
	"github.com/jxskiss/gopkg/v2/easy/ezmap"
	"github.com/jxskiss/gopkg/v2/perf/gopool"
)

type workflow struct {
	id    string
	tasks map[string]Task
	dag   *dag.DAG[string]
	mu    sync.RWMutex

	observer Observer
	input    ezmap.Map

	runCtx  *runContext
	initErr error

	// Channels for dynamic task execution
	addCh chan []Task
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
	w.runCtx = newRunContext(w, w.input, w.observer)
	return w
}

func (w *workflow) ID() string { return w.id }

func (w *workflow) Tasks() []Task {
	w.mu.RLock()
	defer w.mu.RUnlock()
	tasks := make([]Task, 0, len(w.tasks))
	taskIDs := w.dag.TopoSort()
	for _, id := range taskIDs {
		tasks = append(tasks, w.tasks[id])
	}
	return tasks
}

func (w *workflow) AddTask(ctx context.Context, tasks ...Task) error {
	var addCh chan []Task
	err := func() error {
		w.mu.Lock()
		defer w.mu.Unlock()

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

		// If workflow is running, notify execution loop
		addCh = w.addCh
		return nil
	}()
	if err != nil {
		return err
	}

	if addCh != nil {
		select {
		case addCh <- tasks:
			// sent successfully
		case <-ctx.Done():
			return ctx.Err()
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

	// Initialize channel for dynamic tasks
	w.mu.Lock()
	w.addCh = make(chan []Task, 1) // Buffer to prevent blocking AddTask
	// Note: We don't close addCh because AddTask might be called concurrently
	// even after we finish here (though checking w.addCh != nil under lock helps).
	// Ideally we should set w.addCh = nil on exit.
	defer func() {
		w.mu.Lock()
		w.addCh = nil
		w.mu.Unlock()
	}()
	w.mu.Unlock()

	completed := make(map[string]bool)
	running := make(map[string]bool)

	type taskDoneEvent struct {
		taskID string
		res    *TaskResult
	}
	// Use a larger buffer or dynamic checking to avoid deadlocks with dynamic tasks
	// But since we control the senders, a buffer of len(tasks) is only good for initial set.
	// For dynamic, we rely on the loop reading fast enough.
	doneCh := make(chan taskDoneEvent, 128)

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
	// Need lock as we access DAG
	w.mu.RLock()
	zeroIn := w.dag.ListZeroIncomingVertices()
	slices.Sort(zeroIn) // Ensure deterministic order

	startedCount := 0
	for _, id := range zeroIn {
		if t, ok := w.tasks[id]; ok {
			startTask(t)
			startedCount++
		}
	}
	currentTaskCount := len(w.tasks)
	w.mu.RUnlock()

	if startedCount == 0 && currentTaskCount > 0 {
		// Should not happen if cycle check passed and deps exist
		return nil, errors.New("no tasks can be started (possible cycle or logic error)")
	}

	if currentTaskCount == 0 {
		goto Finish
	}

	for len(completed) < currentTaskCount {
		select {
		case <-ctx.Done():
			result.Error = ctx.Err()
			goto Finish
		case newTasks := <-w.addCh:
			// Handle dynamically added tasks
			// We need to re-evaluate currentTaskCount
			w.mu.RLock()
			currentTaskCount = len(w.tasks)

			for _, t := range newTasks {
				// Check if this new task is ready to run
				// It's ready if all its dependencies are completed
				if completed[t.ID()] || running[t.ID()] {
					continue
				}

				allDepsDone := true
				for _, dep := range t.Depends() {
					if !completed[dep] {
						allDepsDone = false
						break
					}
				}
				if allDepsDone {
					startTask(t)
				}
			}
			w.mu.RUnlock()

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
			w.mu.RLock()
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
			w.mu.RUnlock()
		}
	}

Finish:
	result.EndTime = time.Now()

	if w.observer != nil {
		w.observer.OnWorkflowCompleted(ctx, w.id, result)
	}

	return result, result.Error
}
