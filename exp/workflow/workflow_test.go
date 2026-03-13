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

	"github.com/jxskiss/gopkg/v2/easy/ezmap"
)

func TestBuildFluentAndRun(t *testing.T) {
	wf := New("test")
	require.NoError(t, wf.AddTask("A", func(ctx context.Context, in TaskInput) (any, error) {
		return ezmap.Map{"v": 1}, nil
	}, ezmap.Map{"p": "x"}))
	require.NoError(t, wf.AddTask("B", func(ctx context.Context, in TaskInput) (any, error) {
		aOut := in.UpstreamOutputs.Get("A")
		require.NotNil(t, aOut)
		assert.Equal(t, "x", (aOut.(ezmap.Map)).GetOr("p", "x"))
		return ezmap.Map{"vb": (aOut.(ezmap.Map)).GetInt("v") + 1}, nil
	}, nil))
	require.NoError(t, wf.DependsOn("B", "A"))

	res, err := wf.Run(context.Background(), RunOptions{MaxConcurrency: 2, FailurePolicy: FailFast})
	require.NoError(t, err)
	require.False(t, res.StartedAt.IsZero())
	require.False(t, res.EndedAt.IsZero())
	require.False(t, res.EndedAt.Before(res.StartedAt))
	require.Equal(t, Succeeded, res.Tasks["A"].State)
	require.Equal(t, Succeeded, res.Tasks["B"].State)
	require.Equal(t, 2, (res.Tasks["B"].Output.(ezmap.Map)).GetInt("vb"))
}

func TestRunDefault(t *testing.T) {
	wf := New("default")
	require.NoError(t, wf.AddTask("A", func(ctx context.Context, in TaskInput) (any, error) {
		return ezmap.Map{"v": 1}, nil
	}, nil))

	res, err := wf.RunDefault(context.Background())
	require.NoError(t, err)
	require.Equal(t, Succeeded, res.Tasks["A"].State)
	require.Equal(t, 1, (res.Tasks["A"].Output.(ezmap.Map)).GetInt("v"))
}

func TestBuildDeclarative(t *testing.T) {
	wf, err := Build(Spec{
		Name: "s",
		Tasks: []TaskSpec{
			{Name: "A", Action: func(ctx context.Context, in TaskInput) (any, error) { return ezmap.Map{"x": 1}, nil }},
			{Name: "B", DependsOn: []string{"A"}, Action: func(ctx context.Context, in TaskInput) (any, error) {
				aOut := in.UpstreamOutputs.Get("A")
				require.NotNil(t, aOut)
				return ezmap.Map{"x": (aOut.(ezmap.Map)).GetInt("x") + 1}, nil
			}},
		},
	})
	require.NoError(t, err)
	res, err := wf.Run(context.Background(), RunOptions{FailurePolicy: FailFast})
	require.NoError(t, err)
	assert.Equal(t, 2, (res.Tasks["B"].Output.(ezmap.Map)).GetInt("x"))
}

func TestRunContextSharedData(t *testing.T) {
	wf := New("runctx")
	require.NoError(t, wf.AddTask("A", func(ctx context.Context, in TaskInput) (any, error) {
		in.RunCtx.Store("token", "abc")
		in.RunCtx.Store("n", 3)
		return nil, nil
	}, nil))
	require.NoError(t, wf.AddTask("B", func(ctx context.Context, in TaskInput) (any, error) {
		token, ok := in.RunCtx.Load("token")
		require.True(t, ok)
		n, ok := in.RunCtx.Load("n")
		require.True(t, ok)
		return ezmap.Map{
			"token": token.(string),
			"n2":    n.(int) * 2,
		}, nil
	}, nil))
	require.NoError(t, wf.DependsOn("B", "A"))

	res, err := wf.Run(context.Background(), RunOptions{FailurePolicy: FailFast})
	require.NoError(t, err)
	out := res.Tasks["B"].Output.(ezmap.Map)
	assert.Equal(t, "abc", out.GetString("token"))
	assert.Equal(t, 6, out.GetInt("n2"))
}

func TestRunContextFromOptions(t *testing.T) {
	wf := New("runctx-options")
	require.NoError(t, wf.AddTask("A", func(ctx context.Context, in TaskInput) (any, error) {
		v, ok := in.RunCtx.Load("seed")
		require.True(t, ok)
		return ezmap.Map{"seed": v.(int) + 1}, nil
	}, nil))

	shared := NewRunContext()
	shared.Store("seed", 41)
	res, err := wf.Run(context.Background(), RunOptions{
		FailurePolicy: FailFast,
		RunContext:    shared,
	})
	require.NoError(t, err)
	assert.Equal(t, 42, (res.Tasks["A"].Output.(ezmap.Map)).GetInt("seed"))
	require.Same(t, shared, res.RunCtx)
}

func TestRunContextAutoCreateWhenNil(t *testing.T) {
	wf := New("runctx-nil")
	require.NoError(t, wf.AddTask("A", func(ctx context.Context, in TaskInput) (any, error) {
		require.NotNil(t, in.RunCtx)
		in.RunCtx.Store("ok", true)
		v, ok := in.RunCtx.Load("ok")
		require.True(t, ok)
		return ezmap.Map{"ok": v.(bool)}, nil
	}, nil))

	res, err := wf.Run(context.Background(), RunOptions{
		FailurePolicy: FailFast,
		RunContext:    nil,
	})
	require.NoError(t, err)
	require.NotNil(t, res.RunCtx)
	v, ok := res.RunCtx.Load("ok")
	require.True(t, ok)
	require.Equal(t, true, v)
	assert.True(t, (res.Tasks["A"].Output.(ezmap.Map)).GetBool("ok"))
}

func TestFailFast(t *testing.T) {
	wf := New("failfast")
	startC := make(chan struct{})
	require.NoError(t, wf.AddTask("A", func(ctx context.Context, in TaskInput) (any, error) {
		close(startC)
		return nil, errors.New("boom")
	}, nil))
	require.NoError(t, wf.AddTask("B", func(ctx context.Context, in TaskInput) (any, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}, nil))
	require.NoError(t, wf.AddTask("C", func(ctx context.Context, in TaskInput) (any, error) {
		<-startC
		time.Sleep(30 * time.Millisecond)
		return ezmap.Map{"ok": true}, nil
	}, nil))
	require.NoError(t, wf.DependsOn("B", "A"))

	res, err := wf.Run(context.Background(), RunOptions{MaxConcurrency: 2, FailurePolicy: FailFast})
	require.Error(t, err)
	assert.Equal(t, Failed, res.Tasks["A"].State)
	assert.Equal(t, Canceled, res.Tasks["B"].State)
	assert.Contains(t, []TaskState{Succeeded, Canceled}, res.Tasks["C"].State)
}

func TestBestEffort(t *testing.T) {
	wf := New("besteffort")
	require.NoError(t, wf.AddTask("A", func(ctx context.Context, in TaskInput) (any, error) {
		return nil, errors.New("boom")
	}, nil))
	require.NoError(t, wf.AddTask("B", func(ctx context.Context, in TaskInput) (any, error) {
		return ezmap.Map{"ok": true}, nil
	}, nil))
	require.NoError(t, wf.AddTask("C", func(ctx context.Context, in TaskInput) (any, error) {
		bOut := in.UpstreamOutputs.Get("B")
		require.NotNil(t, bOut)
		return ezmap.Map{"ok": (bOut.(ezmap.Map)).GetBool("ok")}, nil
	}, nil))
	require.NoError(t, wf.AddTask("D", func(ctx context.Context, in TaskInput) (any, error) {
		return ezmap.Map{"never": true}, nil
	}, nil))
	require.NoError(t, wf.DependsOn("C", "B"))
	require.NoError(t, wf.DependsOn("D", "A"))

	res, err := wf.Run(context.Background(), RunOptions{MaxConcurrency: 4, FailurePolicy: BestEffort})
	require.Error(t, err)
	assert.Equal(t, Failed, res.Tasks["A"].State)
	assert.Equal(t, SkippedDependencyFailed, res.Tasks["D"].State)
	assert.Equal(t, Succeeded, res.Tasks["B"].State)
	assert.Equal(t, Succeeded, res.Tasks["C"].State)
}

func TestMaxConcurrency(t *testing.T) {
	wf := New("concurrency")
	var current int64
	var peak int64
	mkTask := func(name string) TaskFunc {
		return func(ctx context.Context, in TaskInput) (any, error) {
			n := atomic.AddInt64(&current, 1)
			for {
				p := atomic.LoadInt64(&peak)
				if n <= p || atomic.CompareAndSwapInt64(&peak, p, n) {
					break
				}
			}
			time.Sleep(20 * time.Millisecond)
			atomic.AddInt64(&current, -1)
			return ezmap.Map{"name": name}, nil
		}
	}

	for _, name := range []string{"A", "B", "C", "D", "E"} {
		require.NoError(t, wf.AddTask(name, mkTask(name), nil))
	}
	res, err := wf.Run(context.Background(), RunOptions{MaxConcurrency: 2, FailurePolicy: FailFast})
	require.NoError(t, err)
	assert.LessOrEqual(t, peak, int64(2))
	for _, r := range res.Tasks {
		assert.Equal(t, Succeeded, r.State)
	}
}

func TestValidation(t *testing.T) {
	wf := New("validation")
	err := wf.AddTask("", func(ctx context.Context, in TaskInput) (any, error) { return nil, nil }, nil)
	require.Error(t, err)
	err = wf.AddTask("A", nil, nil)
	require.Error(t, err)
	require.NoError(t, wf.AddTask("A", func(ctx context.Context, in TaskInput) (any, error) { return nil, nil }, nil))
	err = wf.AddTask("A", func(ctx context.Context, in TaskInput) (any, error) { return nil, nil }, nil)
	require.Error(t, err)
	err = wf.DependsOn("B", "A")
	require.Error(t, err)
	err = wf.DependsOn("A", "A")
	require.Error(t, err)
}

func TestCancelContext(t *testing.T) {
	wf := New("ctx")
	var once sync.Once
	require.NoError(t, wf.AddTask("A", func(ctx context.Context, in TaskInput) (any, error) {
		once.Do(func() { time.Sleep(50 * time.Millisecond) })
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return ezmap.Map{"ok": true}, nil
		}
	}, nil))
	require.NoError(t, wf.AddTask("B", func(ctx context.Context, in TaskInput) (any, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}, nil))

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	res, err := wf.Run(ctx, RunOptions{MaxConcurrency: 1, FailurePolicy: FailFast})
	require.Error(t, err)
	assert.Contains(t, []TaskState{Failed, Canceled}, res.Tasks["A"].State)
	assert.Equal(t, Canceled, res.Tasks["B"].State)
}

func TestResultHelpers(t *testing.T) {
	res := &Result{
		TopoOrder: []string{"ok", "failed", "skipped", "canceled"},
		Tasks: map[string]*TaskResult{
			"ok": {
				Name:  "ok",
				State: Succeeded,
				Err:   nil,
			},
			"failed": {
				Name:  "failed",
				State: Failed,
				Err:   fmt.Errorf("boom"),
			},
			"skipped": {
				Name:  "skipped",
				State: SkippedDependencyFailed,
				Err:   fmt.Errorf("dependency failed"),
			},
			"canceled": {
				Name:  "canceled",
				State: Canceled,
				Err:   context.Canceled,
			},
		},
	}

	errs := res.FailedTasks()
	require.Len(t, errs, 3)
	require.ErrorContains(t, errs["failed"], "boom")
	require.ErrorContains(t, errs["skipped"], "dependency failed")
	require.ErrorIs(t, errs["canceled"], context.Canceled)

	require.ErrorContains(t, res.TaskErrors(), "boom")
	require.ErrorContains(t, res.TaskErrors(), "context canceled")

	var nilRes *Result
	require.Nil(t, nilRes.FailedTasks())
	require.Nil(t, nilRes.TaskErrors())
}

func TestUpstreamOutputsIncludeAllAncestors(t *testing.T) {
	wf := New("upstream-ancestors")

	require.NoError(t, wf.AddTask("A", func(ctx context.Context, in TaskInput) (any, error) {
		return ezmap.Map{"a": 1}, nil
	}, nil))
	require.NoError(t, wf.AddTask("B", func(ctx context.Context, in TaskInput) (any, error) {
		aOut := in.UpstreamOutputs.Get("A")
		require.NotNil(t, aOut)
		return ezmap.Map{"b": (aOut.(ezmap.Map)).GetInt("a") + 1}, nil
	}, nil))
	require.NoError(t, wf.AddTask("C", func(ctx context.Context, in TaskInput) (any, error) {
		aOut := in.UpstreamOutputs.Get("A")
		require.NotNil(t, aOut)
		bOut := in.UpstreamOutputs.Get("B")
		require.NotNil(t, bOut)

		// C only depends on B, but it should still see A as an upstream ancestor.
		return ezmap.Map{
			"sum": (aOut.(ezmap.Map)).GetInt("a") + (bOut.(ezmap.Map)).GetInt("b"),
		}, nil
	}, nil))

	require.NoError(t, wf.DependsOn("B", "A"))
	require.NoError(t, wf.DependsOn("C", "B"))

	res, err := wf.Run(context.Background(), RunOptions{FailurePolicy: FailFast})
	require.NoError(t, err)
	assert.Equal(t, 3, (res.Tasks["C"].Output.(ezmap.Map)).GetInt("sum"))
}
