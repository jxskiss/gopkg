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

	"github.com/jxskiss/gopkg/v2/perf/gopool"
)

func TestWorkflow_ComplexDAG(t *testing.T) {
	// 模拟任务执行记录，用于验证顺序
	var executionLog []string
	var mu sync.Mutex
	logExec := func(id string) {
		mu.Lock()
		defer mu.Unlock()
		executionLog = append(executionLog, id)
	}

	// 定义任务
	tasks := []Task{
		&mockTask{
			id:     "FetchData",
			output: "raw_data",
			action: func() { logExec("FetchData") },
		},
		&mockTask{
			id:      "ValidateData",
			depends: []string{"FetchData"},
			action:  func() { logExec("ValidateData") },
			fn: func(in any) (any, error) {
				s, _ := in.(string)
				return s + "_valid", nil
			},
		},
		&mockTask{
			id:      "LogMetric",
			depends: []string{"FetchData"},
			action:  func() { logExec("LogMetric") },
			output:  "metric_done",
			sleep:   50 * time.Millisecond,
		},
		&mockTask{
			id:      "EnrichData",
			depends: []string{"ValidateData"},
			action:  func() { logExec("EnrichData") },
			fn: func(in any) (any, error) {
				s, _ := in.(string)
				return s + "_enriched", nil
			},
		},
		&mockTask{
			id:      "SaveToDB",
			depends: []string{"EnrichData"},
			action:  func() { logExec("SaveToDB") },
			output:  "db_id_123",
			sleep:   10 * time.Millisecond,
		},
		&mockTask{
			id:      "SendEmail",
			depends: []string{"EnrichData"},
			action:  func() { logExec("SendEmail") },
			output:  "email_sent",
			sleep:   10 * time.Millisecond,
		},
		&mockTask{
			id:      "Report",
			depends: []string{"SaveToDB", "SendEmail", "LogMetric"},
			action:  func() { logExec("Report") },
			output:  "report_generated",
		},
	}

	wf := NewWorkflow("wf-complex-dag")
	ctx := context.Background()

	err := wf.AddTask(ctx, tasks...)
	require.NoError(t, err)

	start := time.Now()
	res, err := wf.Execute(ctx, nil)
	require.NoError(t, err)
	duration := time.Since(start)

	t.Logf("Workflow finished in %v", duration)

	// 验证结果
	require.NotNil(t, res)
	require.NotNil(t, res.TaskResults)

	taskRes, ok := res.TaskResults["Report"]
	require.True(t, ok)
	assert.Equal(t, "report_generated", taskRes.Output)

	// 验证执行顺序 (基于时间戳)
	tr := res.TaskResults

	checkOrder := func(prev, next string) {
		p := tr[prev]
		n := tr[next]
		require.NotNil(t, p, prev)
		require.NotNil(t, n, next)
		assert.True(t, !p.EndTime.After(n.StartTime),
			fmt.Sprintf("%s finished at %v, but %s started at %v", prev, p.EndTime.Format(time.StampMilli), next, n.StartTime.Format(time.StampMilli)))
	}

	checkOrder("FetchData", "ValidateData")
	checkOrder("FetchData", "LogMetric")
	checkOrder("ValidateData", "EnrichData")
	checkOrder("EnrichData", "SaveToDB")
	checkOrder("EnrichData", "SendEmail")
	checkOrder("SaveToDB", "Report")
	checkOrder("SendEmail", "Report")
	checkOrder("LogMetric", "Report")

	// 验证数据流
	assert.Equal(t, "raw_data_valid", tr["ValidateData"].Output)
	assert.Equal(t, "raw_data_valid_enriched", tr["EnrichData"].Output)
}

func TestWorkflow_WithInput(t *testing.T) {
	inputJSON := `{"env": "test", "retries": 3}`
	wf := NewWorkflow("wf-input", WithJSONInput([]byte(inputJSON)))

	// Check if input is accessible via RunContext
	rc := wf.RunContext()
	require.NotNil(t, rc)
	input := rc.WorkflowInput()

	assert.Equal(t, "test", input.GetString("env"))
	assert.Equal(t, 3, input.GetInt("retries"))

	// Also test execution with init error
	wfErr := NewWorkflow("wf-err", WithJSONInput([]byte(`invalid-json`)))
	res, err := wfErr.Execute(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "unmarshal JSON input")
}

func TestNewTask(t *testing.T) {
	// Simple task execution
	wf := NewWorkflow("wf-new-task")

	task1 := NewTask("task1", func(ctx context.Context, rc RunContext) (any, error) {
		return 1, nil
	})

	task2 := NewTask("task2", func(ctx context.Context, rc RunContext) (any, error) {
		v1, ok := rc.GetTaskOutput("task1")
		if !ok {
			return nil, fmt.Errorf("task1 output not found")
		}
		return v1.(int) + 1, nil
	}, DependsOn("task1"))

	err := wf.AddTask(context.Background(), task1, task2)
	require.NoError(t, err)

	res, err := wf.Execute(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, 2, res.TaskResults["task2"].Output)
}

func TestWorkflow_TaskPanic(t *testing.T) {
	// Setup gopool with custom panic handler
	var panicVal any
	var panicMu sync.Mutex

	pool := gopool.New("test-pool", &gopool.Option{
		PanicHandler: func(ctx context.Context, r any) {
			panicMu.Lock()
			panicVal = r
			panicMu.Unlock()
		},
		MaxIdleWorkers: 10,
		WorkerMaxAge:   time.Minute,
		TaskChanBuffer: 10,
	})

	wf := NewWorkflow("wf-panic")

	// Create a task that panics
	task1 := NewTask("panic-task", func(ctx context.Context, rc RunContext) (any, error) {
		panic("something went wrong")
	})

	// A downstream task that should not run
	task2 := NewTask("downstream-task", func(ctx context.Context, rc RunContext) (any, error) {
		return "ok", nil
	}, DependsOn("panic-task"))

	err := wf.AddTask(context.Background(), task1, task2)
	require.NoError(t, err)

	res, err := wf.Execute(context.Background(), pool)

	// Workflow should return error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "task panic-task panicked")

	// Verify result
	require.NotNil(t, res)
	require.NotNil(t, res.TaskResults["panic-task"])
	// "downstream-task" should not be in results as it didn't run
	assert.Nil(t, res.TaskResults["downstream-task"])

	// Verify gopool captured the panic
	panicMu.Lock()
	defer panicMu.Unlock()
	assert.Equal(t, "something went wrong", panicVal)
}

func TestWorkflow_CancellationOnError(t *testing.T) {
	wf := NewWorkflow("test-cancellation")

	taskFail := NewTask("fail-task", func(ctx context.Context, rc RunContext) (interface{}, error) {
		time.Sleep(50 * time.Millisecond)
		return nil, errors.New("deliberate error")
	})

	longTaskCanceled := int32(0)
	taskLong := NewTask("long-task", func(ctx context.Context, rc RunContext) (interface{}, error) {
		select {
		case <-time.After(500 * time.Millisecond):
			return "finished", nil
		case <-ctx.Done():
			atomic.StoreInt32(&longTaskCanceled, 1)
			return nil, ctx.Err()
		}
	})

	err := wf.AddTask(context.Background(), taskFail, taskLong)
	assert.NoError(t, err)

	start := time.Now()
	res, err := wf.Execute(context.Background(), nil)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Equal(t, "deliberate error", err.Error())
	assert.Less(t, duration, 400*time.Millisecond, "Workflow should finish quickly upon error")

	// Give a bit of time for the goroutine to process the cancellation signal
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&longTaskCanceled), "Long running task should be canceled")

	// Ensure result contains the executed task results even on error
	// The failed task should be in results
	if res != nil {
		assert.Contains(t, res.TaskResults, "fail-task")
		assert.Equal(t, "deliberate error", res.TaskResults["fail-task"].Error.Error())
	}
}

// mockTask 用于测试的 Task 实现
type mockTask struct {
	id      string
	depends []string
	sleep   time.Duration
	output  any
	err     error
	action  func()
	fn      func(input any) (any, error)
}

func (t *mockTask) ID() string        { return t.id }
func (t *mockTask) Depends() []string { return t.depends }
func (t *mockTask) Run(_ context.Context, rc RunContext) (any, error) {
	if t.action != nil {
		t.action()
	}
	if t.sleep > 0 {
		time.Sleep(t.sleep)
	}
	if t.err != nil {
		return nil, t.err
	}

	if t.fn != nil {
		var input any
		if len(t.depends) > 0 {
			// 默认取第一个依赖的输出
			if val, ok := rc.GetTaskOutput(t.depends[0]); ok {
				input = val
			}
		}
		return t.fn(input)
	}
	return t.output, nil
}
