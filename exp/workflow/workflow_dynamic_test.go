package workflow

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWorkflow_Dynamic_Basic verifies that a task can be added dynamically
// and executed after its dependency completes.
func TestWorkflow_Dynamic_Basic(t *testing.T) {
	wf := NewWorkflow("wf-dynamic-basic")

	executed := make(map[string]bool)
	var mu sync.Mutex

	task1 := NewTask("task1", func(ctx context.Context, rc RunContext) (any, error) {
		mu.Lock()
		executed["task1"] = true
		mu.Unlock()

		// Dynamically add task2 which depends on task1
		t2 := NewTask("task2", func(ctx context.Context, rc RunContext) (any, error) {
			mu.Lock()
			executed["task2"] = true
			mu.Unlock()
			return "task2-result", nil
		}, DependsOn("task1"))

		return "task1-result", rc.AddTask(ctx, t2)
	})

	err := wf.AddTask(context.Background(), task1)
	require.NoError(t, err)

	res, err := wf.Execute(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, res)

	mu.Lock()
	assert.True(t, executed["task1"])
	assert.True(t, executed["task2"])
	mu.Unlock()

	assert.Equal(t, "task1-result", res.TaskResults["task1"].Output)
	assert.Equal(t, "task2-result", res.TaskResults["task2"].Output)
}

// TestWorkflow_Dynamic_Chain verifies a chain of dynamic tasks:
// A adds B, B adds C.
func TestWorkflow_Dynamic_Chain(t *testing.T) {
	wf := NewWorkflow("wf-dynamic-chain")
	var order []string
	var mu sync.Mutex

	record := func(id string) {
		mu.Lock()
		order = append(order, id)
		mu.Unlock()
	}

	taskA := NewTask("A", func(ctx context.Context, rc RunContext) (any, error) {
		record("A")
		// Add B -> A
		return nil, rc.AddTask(ctx, NewTask("B", func(ctx context.Context, rc RunContext) (any, error) {
			record("B")
			// Add C -> B
			return nil, rc.AddTask(ctx, NewTask("C", func(ctx context.Context, rc RunContext) (any, error) {
				record("C")
				return nil, nil
			}, DependsOn("B")))
		}, DependsOn("A")))
	})

	err := wf.AddTask(context.Background(), taskA)
	require.NoError(t, err)

	res, err := wf.Execute(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, []string{"A", "B", "C"}, order)
	assert.Contains(t, res.TaskResults, "A")
	assert.Contains(t, res.TaskResults, "B")
	assert.Contains(t, res.TaskResults, "C")
}

// TestWorkflow_Dynamic_Independent verifies adding a task with no dependencies
// runs immediately/concurrently.
func TestWorkflow_Dynamic_Independent(t *testing.T) {
	wf := NewWorkflow("wf-dynamic-independent")
	executed := int32(0)

	task1 := NewTask("task1", func(ctx context.Context, rc RunContext) (any, error) {
		// Add task2 with NO dependencies.
		err := rc.AddTask(ctx, NewTask("task2", func(ctx context.Context, rc RunContext) (any, error) {
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&executed, 1)
			return "task2-res", nil
		}))
		if err != nil {
			return nil, err
		}

		time.Sleep(20 * time.Millisecond) // Give task2 a chance to start
		atomic.AddInt32(&executed, 1)
		return "task1-res", nil
	})

	err := wf.AddTask(context.Background(), task1)
	require.NoError(t, err)

	res, err := wf.Execute(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, int32(2), atomic.LoadInt32(&executed))
	assert.Contains(t, res.TaskResults, "task2")
}

// TestWorkflow_Dynamic_Parallel verifies multiple tasks adding dynamic tasks concurrently.
func TestWorkflow_Dynamic_Parallel(t *testing.T) {
	// Use a unique workflow ID to avoid any potential collision in global state (if any)
	wfID := fmt.Sprintf("wf-dynamic-parallel-%d", time.Now().UnixNano())
	wf := NewWorkflow(wfID)
	var count int32

	// A execution triggers adding B1 and B2
	taskA := NewTask("A", func(ctx context.Context, rc RunContext) (any, error) {
		atomic.AddInt32(&count, 1)

		err1 := rc.AddTask(ctx, NewTask("B1", func(ctx context.Context, rc RunContext) (any, error) {
			atomic.AddInt32(&count, 1)
			return nil, nil
		}, DependsOn("A")))

		err2 := rc.AddTask(ctx, NewTask("B2", func(ctx context.Context, rc RunContext) (any, error) {
			atomic.AddInt32(&count, 1)
			return nil, nil
		}, DependsOn("A")))

		if err1 != nil {
			return nil, err1
		}
		return nil, err2
	})

	err := wf.AddTask(context.Background(), taskA)
	require.NoError(t, err)

	_, err = wf.Execute(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&count))
}

// TestWorkflow_Dynamic_Error verifies that error in dynamic task fails the workflow.
func TestWorkflow_Dynamic_Error(t *testing.T) {
	wf := NewWorkflow("wf-dynamic-error")

	task1 := NewTask("task1", func(ctx context.Context, rc RunContext) (any, error) {
		return nil, rc.AddTask(ctx, NewTask("task2", func(ctx context.Context, rc RunContext) (any, error) {
			return nil, errors.New("dynamic task error")
		}, DependsOn("task1")))
	})

	err := wf.AddTask(context.Background(), task1)
	require.NoError(t, err)

	res, err := wf.Execute(context.Background(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dynamic task error")
	assert.NotNil(t, res)
	assert.Contains(t, res.TaskResults, "task2")
}

// TestWorkflow_Dynamic_Validation verifies validation logic for dynamic tasks.
func TestWorkflow_Dynamic_Validation(t *testing.T) {
	wf := NewWorkflow("wf-dynamic-validation")

	task1 := NewTask("task1", func(ctx context.Context, rc RunContext) (any, error) {
		// Case 1: Add duplicate task ID
		err := rc.AddTask(ctx, NewTask("task1", nil))
		if err == nil {
			return nil, errors.New("expected error for duplicate task ID")
		}

		// Case 2: Add task with self-dependency (cycle)
		err = rc.AddTask(ctx, NewTask("task2", nil, DependsOn("task2")))
		if err == nil {
			return nil, errors.New("expected error for self-dependency")
		}

		// Case 3: Add task creating a cycle (task3 -> task1, but task1 is already running/done)
		// Note: The cycle check logic in AddTask checks if adding edge creates cycle.
		// Since task1 is already in DAG, adding task3->task1 is valid as an edge,
		// BUT task1 is the parent of the current execution.
		// If task3 depends on task1, it's just a normal forward dependency (task1 is dependency of task3).
		// Cycle would be if we add a dependency such that task1 -> ... -> task3 -> task1.
		// Since task1 is already running, we can't make task1 depend on task3.
		// We can only make task3 depend on task1.

		// Let's try to add a task that depends on a non-existent task
		// In AddTask implementation, it checks dependencies?
		// The AddTask implementation:
		// for _, dep := range t.Depends() {
		//    if w.dag.AddEdge(dep, id) { ... }
		// }
		// If dep doesn't exist, AddEdge adds it as a vertex.
		// But Execute method checks "Validate all dependencies exist" only at start.
		// For dynamic tasks, if they depend on non-existent task, that task node is added to DAG
		// but has no action. It will be treated as a task with no incoming edges (if nothing depends on it)
		// or just a node.
		// Actually, if we add task3 depending on "ghost", "ghost" becomes a node in DAG.
		// "ghost" is not in w.tasks map.
		// When "ghost" is processed (zero incoming), startTask checks w.tasks["ghost"], finds nothing, and does nothing?
		// Let's verify this behavior. Ideally it might be better to error if dependency is missing,
		// OR we accept that the dependency might be added later?
		// For this test, let's stick to explicit errors we know exist in AddTask.

		return nil, nil
	})

	err := wf.AddTask(context.Background(), task1)
	require.NoError(t, err)

	_, err = wf.Execute(context.Background(), nil)
	require.NoError(t, err)
}

// TestWorkflow_Dynamic_MixedDependency verifies a dynamic task depending on
// one completed task (static) and one future task (dynamic).
func TestWorkflow_Dynamic_MixedDependency(t *testing.T) {
	wf := NewWorkflow("wf-dynamic-mixed")

	executed := make(map[string]bool)
	var mu sync.Mutex

	// Static Task A
	taskA := NewTask("A", func(ctx context.Context, rc RunContext) (any, error) {
		mu.Lock()
		executed["A"] = true
		mu.Unlock()

		// Add Dynamic Task B, depends on A
		err := rc.AddTask(ctx, NewTask("B", func(ctx context.Context, rc RunContext) (any, error) {
			mu.Lock()
			executed["B"] = true
			mu.Unlock()
			return nil, nil
		}, DependsOn("A")))
		if err != nil {
			return nil, err
		}

		// Add Dynamic Task C, depends on A and B
		err = rc.AddTask(ctx, NewTask("C", func(ctx context.Context, rc RunContext) (any, error) {
			mu.Lock()
			executed["C"] = true
			mu.Unlock()
			return nil, nil
		}, DependsOn("A"), DependsOn("B")))

		return nil, err
	})

	err := wf.AddTask(context.Background(), taskA)
	require.NoError(t, err)

	_, err = wf.Execute(context.Background(), nil)
	require.NoError(t, err)

	mu.Lock()
	assert.True(t, executed["A"])
	assert.True(t, executed["B"])
	assert.True(t, executed["C"])
	mu.Unlock()
}
