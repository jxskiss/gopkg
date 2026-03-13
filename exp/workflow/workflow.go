package workflow

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sort"
	"time"

	"github.com/jxskiss/gopkg/v2/collection/dag"
)

// TaskFunc is the function signature of a workflow task.
type TaskFunc func(ctx context.Context, in TaskInput) (out any, err error)

// TaskInput contains the input data for a task.
type TaskInput struct {
	RunCtx *RunContext
	Params any

	// To make the graph easier to construct, we provide all ancestor task
	// outputs to downstream tasks, instead of just the ones that
	// the current task directly depends on.
	UpstreamOutputs TaskOutputs
}

// TaskOutputs is a read-only view of task outputs keyed by task name.
// The view itself is immutable after creation and safe for concurrent reads.
type TaskOutputs struct {
	data map[string]any
}

func newTaskOutputs(data map[string]any) TaskOutputs {
	return TaskOutputs{data: data}
}

// Get returns the output of a task, or nil if not found.
func (o TaskOutputs) Get(name string) any {
	return o.data[name]
}

// FailurePolicy controls how workflow runner reacts to task failures.
type FailurePolicy int

const (
	// FailFast cancels not-yet-started tasks once any task fails.
	FailFast FailurePolicy = iota
	// BestEffort continues unrelated branches and skips tasks depending on failed tasks.
	BestEffort
)

// RunOptions contains runtime options for a workflow run.
type RunOptions struct {
	MaxConcurrency int
	FailurePolicy  FailurePolicy
	RunContext     *RunContext
}

func (opt *RunOptions) setDefaults() {
	if opt.MaxConcurrency <= 0 {
		opt.MaxConcurrency = runtime.GOMAXPROCS(0)
	}
	if opt.MaxConcurrency < 1 {
		opt.MaxConcurrency = 1
	}
	if opt.FailurePolicy < FailFast || opt.FailurePolicy > BestEffort {
		opt.FailurePolicy = FailFast
	}
}

// TaskState is the terminal or intermediate state of a task.
type TaskState int

const (
	Pending TaskState = iota
	Running
	Succeeded
	Failed
	SkippedDependencyFailed
	Canceled
)

// TaskResult describes execution result of one task.
type TaskResult struct {
	Name      string
	State     TaskState
	Output    any
	Err       error
	StartedAt time.Time
	EndedAt   time.Time
}

// Result contains execution results for all tasks.
type Result struct {
	RunCtx    *RunContext
	Tasks     map[string]*TaskResult
	TopoOrder []string
	StartedAt time.Time
	EndedAt   time.Time
}

// FailedTasks returns a map of task errors for all non-success terminal states
// with non-nil errors (e.g. Failed, SkippedDependencyFailed, Canceled).
func (r *Result) FailedTasks() map[string]error {
	if r == nil || len(r.Tasks) == 0 {
		return nil
	}
	out := make(map[string]error)
	for name, task := range r.Tasks {
		if task.Err != nil && task.State != Succeeded {
			out[name] = task.Err
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// TaskErrors returns an error that wraps all task errors in result.
// If result is nil or empty, or all tasks are succeeded, nil is returned.
func (r *Result) TaskErrors() error {
	if r == nil || len(r.Tasks) == 0 {
		return nil
	}
	var errs []error
	for _, task := range r.TopoOrder {
		tr := r.Tasks[task]
		if tr == nil || tr.Err == nil || tr.State == Succeeded {
			continue
		}
		errs = append(errs, fmt.Errorf("task %s: %w", tr.Name, tr.Err))
	}
	return errors.Join(errs...)
}

type taskDef struct {
	name   string
	action TaskFunc
	params any
}

// Workflow manages a DAG of tasks and executes them concurrently.
//
// It is not concurrent-safe for graph mutation methods.
type Workflow struct {
	name  string
	graph dag.DAG[string]
	tasks map[string]taskDef
}

// New creates a new workflow with the given name.
func New(name string) *Workflow {
	return &Workflow{
		name:  name,
		tasks: make(map[string]taskDef),
	}
}

func (w *Workflow) Name() string {
	return w.name
}

// AddTask adds a task node to workflow.
func (w *Workflow) AddTask(name string, action TaskFunc, params any) error {
	if name == "" {
		return fmt.Errorf("workflow: task name cannot be empty")
	}
	if action == nil {
		return fmt.Errorf("workflow: task %s has nil TaskFunc", name)
	}
	if _, ok := w.tasks[name]; ok {
		return fmt.Errorf("workflow: task %s already exists", name)
	}
	w.tasks[name] = taskDef{
		name:   name,
		action: action,
		params: params,
	}
	w.graph.AddVertex(name)
	return nil
}

// DependsOn adds dependency edges from deps to task.
func (w *Workflow) DependsOn(task string, deps ...string) error {
	if _, ok := w.tasks[task]; !ok {
		return fmt.Errorf("workflow: task %s not found", task)
	}
	for _, dep := range deps {
		if dep == task {
			return fmt.Errorf("workflow: task %s cannot depend on itself", task)
		}
		if _, ok := w.tasks[dep]; !ok {
			return fmt.Errorf("workflow: dependency task %s not found", dep)
		}
		if cyclic := w.graph.AddEdge(dep, task); cyclic {
			return fmt.Errorf("workflow: adding dependency %s -> %s forms cycle", dep, task)
		}
	}
	return nil
}

type taskDone struct {
	name   string
	output any
	err    error
	endAt  time.Time
}

type runState struct {
	w      *Workflow
	opt    RunOptions
	result *Result

	cancel context.CancelFunc
	runCtx *RunContext

	topoOrder []string
	orderPos  map[string]int

	remainingDeps map[string]int
	failedDep     map[string]bool
	outputs       map[string]any
	ancestors     map[string][]string

	ready             []string
	doneCh            chan taskDone
	running           int
	finished          int
	failFastTriggered bool
}

// RunDefault executes the workflow with default run options.
func (w *Workflow) RunDefault(ctx context.Context) (*Result, error) {
	return w.Run(ctx, RunOptions{})
}

// Run executes the workflow and returns all task details.
func (w *Workflow) Run(ctx context.Context, opt RunOptions) (*Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	opt.setDefaults()
	state, result, err := w.prepareRun(ctx, cancel, opt)
	defer func() {
		result.EndedAt = time.Now()
	}()
	if err != nil {
		return result, err
	}

	for state.finished < len(w.tasks) {
		state.dispatchReadyTasks(ctx)
		if state.running == 0 {
			break
		}
		td := <-state.doneCh
		state.handleTaskDone(ctx, td)
	}
	state.finalizePending(ctx)

	return result, resultError(ctx.Err(), result)
}

func (w *Workflow) prepareRun(_ context.Context, cancel context.CancelFunc, opt RunOptions) (*runState, *Result, error) {
	runCtx := opt.RunContext
	if runCtx == nil {
		runCtx = NewRunContext()
	}

	startAt := time.Now()
	topoOrder := w.graph.TopoSort()
	result := &Result{
		RunCtx:    runCtx,
		Tasks:     make(map[string]*TaskResult, len(w.tasks)),
		TopoOrder: topoOrder,
		StartedAt: startAt,
	}
	state := &runState{
		w:             w,
		opt:           opt,
		result:        result,
		cancel:        cancel,
		runCtx:        runCtx,
		topoOrder:     topoOrder,
		remainingDeps: make(map[string]int, len(w.tasks)),
		failedDep:     make(map[string]bool, len(w.tasks)),
		outputs:       make(map[string]any, len(w.tasks)),
		ancestors:     make(map[string][]string, len(w.tasks)),
		ready:         make([]string, 0, len(w.tasks)),
		doneCh:        make(chan taskDone, len(w.tasks)),
	}
	if len(w.tasks) == 0 {
		return state, result, nil
	}

	for name := range w.tasks {
		result.Tasks[name] = &TaskResult{Name: name, State: Pending}
	}
	if len(state.topoOrder) != len(w.tasks) {
		return state, result, fmt.Errorf("workflow: graph is invalid")
	}

	state.orderPos = make(map[string]int, len(state.topoOrder))
	for i, name := range state.topoOrder {
		state.orderPos[name] = i
		state.remainingDeps[name] = len(w.graph.GetReverseNeighbors(name))
		state.ancestors[name] = state.collectAncestors(name)
	}
	for _, name := range state.topoOrder {
		if state.remainingDeps[name] == 0 {
			state.enqueueReady(name)
		}
	}

	return state, result, nil
}

func (s *runState) dispatchReadyTasks(ctx context.Context) {
	for s.running < s.opt.MaxConcurrency && len(s.ready) > 0 {
		name := s.ready[0]
		s.ready = s.ready[1:]
		if s.result.Tasks[name].State != Pending {
			continue
		}
		if s.failFastTriggered {
			s.markResult(name, Canceled, context.Canceled)
			continue
		}
		s.startTask(ctx, name)
	}
}

func (s *runState) handleTaskDone(ctx context.Context, td taskDone) {
	s.running--
	r := s.result.Tasks[td.name]
	r.EndedAt = td.endAt
	r.Output = td.output

	if td.err != nil {
		r.State = Failed
		r.Err = td.err
		s.result.Tasks[td.name] = r
		s.finished++
		if s.opt.FailurePolicy == FailFast && !s.failFastTriggered {
			s.failFastTriggered = true
			s.cancel()
		}
		for _, child := range s.w.graph.GetNeighbors(td.name) {
			s.failedDep[child] = true
			s.remainingDeps[child]--
			if s.remainingDeps[child] == 0 {
				if s.opt.FailurePolicy == BestEffort {
					s.markSkippedRecursively(child)
				} else if s.failFastTriggered {
					s.markResult(child, Canceled, context.Canceled)
				}
			}
		}
		return
	}

	if ctxErr := ctx.Err(); ctxErr != nil && s.failFastTriggered {
		r.State = Canceled
		r.Err = context.Canceled
		s.result.Tasks[td.name] = r
		s.finished++
		for _, child := range s.w.graph.GetNeighbors(td.name) {
			s.remainingDeps[child]--
			if s.remainingDeps[child] == 0 {
				s.markResult(child, Canceled, context.Canceled)
			}
		}
		return
	}

	r.State = Succeeded
	r.Err = nil
	s.result.Tasks[td.name] = r
	s.finished++
	s.outputs[td.name] = td.output
	for _, child := range s.w.graph.GetNeighbors(td.name) {
		s.remainingDeps[child]--
		if s.remainingDeps[child] == 0 {
			if s.failedDep[child] {
				if s.opt.FailurePolicy == BestEffort {
					s.markSkippedRecursively(child)
				} else {
					s.markResult(child, Canceled, context.Canceled)
				}
			} else {
				s.enqueueReady(child)
			}
		}
	}
}

func (s *runState) finalizePending(ctx context.Context) {
	for _, name := range s.topoOrder {
		if s.result.Tasks[name].State == Pending {
			r := s.result.Tasks[name]
			r.State = Canceled
			if ctxErr := ctx.Err(); ctxErr != nil {
				r.Err = ctxErr
			} else {
				r.Err = context.Canceled
			}
			r.EndedAt = time.Now()
			s.result.Tasks[name] = r
		}
	}
}

func (s *runState) enqueueReady(name string) {
	s.ready = append(s.ready, name)
	sort.SliceStable(s.ready, func(i, j int) bool {
		return s.orderPos[s.ready[i]] < s.orderPos[s.ready[j]]
	})
}

func (s *runState) collectAncestors(name string) []string {
	directDeps := s.w.graph.GetReverseNeighbors(name)
	if len(directDeps) == 0 {
		return nil
	}

	ancestorSet := make(map[string]struct{}, len(directDeps))
	for _, dep := range directDeps {
		ancestorSet[dep] = struct{}{}
		for _, ancestor := range s.ancestors[dep] {
			ancestorSet[ancestor] = struct{}{}
		}
	}

	out := make([]string, 0, len(ancestorSet))
	for _, taskName := range s.topoOrder {
		if taskName == name {
			break
		}
		if _, ok := ancestorSet[taskName]; ok {
			out = append(out, taskName)
		}
	}
	return out
}

func (s *runState) markResult(name string, state TaskState, err error) {
	r := s.result.Tasks[name]
	r.State = state
	r.Err = err
	r.EndedAt = time.Now()
	s.result.Tasks[name] = r
	s.finished++
}

func (s *runState) markSkippedRecursively(name string) {
	r := s.result.Tasks[name]
	if r.State != Pending {
		return
	}
	s.markResult(name, SkippedDependencyFailed, fmt.Errorf("dependency failed"))
	for _, child := range s.w.graph.GetNeighbors(name) {
		s.failedDep[child] = true
		s.remainingDeps[child]--
		if s.remainingDeps[child] == 0 {
			s.markSkippedRecursively(child)
		}
	}
}

func (s *runState) startTask(ctx context.Context, name string) {
	task := s.w.tasks[name]
	ancestors := s.ancestors[name]
	upstream := make(map[string]any, len(ancestors))
	for _, ancestor := range ancestors {
		if out, ok := s.outputs[ancestor]; ok {
			upstream[ancestor] = out
		}
	}
	r := s.result.Tasks[name]
	r.State = Running
	r.StartedAt = time.Now()
	s.result.Tasks[name] = r
	s.running++

	go func() {
		var out any
		var err error
		defer func() {
			if rec := recover(); rec != nil {
				err = fmt.Errorf("task %s panic: %v", name, rec)
			}
			s.doneCh <- taskDone{name: name, output: out, err: err, endAt: time.Now()}
		}()
		out, err = task.action(ctx, TaskInput{
			RunCtx:          s.runCtx,
			Params:          task.params,
			UpstreamOutputs: newTaskOutputs(upstream),
		})
	}()
}

func resultError(ctxErr error, result *Result) error {
	failed := 0
	skipped := 0
	canceled := 0
	for _, r := range result.Tasks {
		switch r.State {
		case Failed:
			failed++
		case SkippedDependencyFailed:
			skipped++
		case Canceled:
			canceled++
		}
	}
	if failed == 0 && skipped == 0 && canceled == 0 {
		return nil
	}
	if failed == 0 && skipped == 0 && ctxErr != nil {
		return ctxErr
	}
	return fmt.Errorf("workflow failed: %d failed, %d skipped, %d canceled", failed, skipped, canceled)
}
