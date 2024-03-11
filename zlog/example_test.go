package zlog

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func testHelperReplaceGlobalsToStdout(ctxFunc func(ctx context.Context) CtxResult) func() {
	cfg := &Config{
		Level:             "trace",
		Format:            "json",
		DisableTimestamp:  true,
		DisableCaller:     true,
		DisableStacktrace: true,
		GlobalConfig: GlobalConfig{
			CtxHandler: CtxHandler{
				WithCtx: ctxFunc,
			},
		},
	}
	l, p, err := NewWithOutput(cfg, zapcore.AddSync(os.Stdout))
	if err != nil {
		panic(err)
	}
	return ReplaceGlobals(l, p)
}

func ExampleWith() {
	defer testHelperReplaceGlobalsToStdout(nil)()

	With(zap.String("k1", "v1"), zap.Int64("k2", 54321)).
		Info("example with")

	// Output:
	// {"level":"info","msg":"example with","k1":"v1","k2":54321}
}

func ExampleWithCtx() {

	demoCtxFunc := func(ctx context.Context) CtxResult {
		return CtxResult{
			Fields: []zap.Field{zap.String("ctx1", "v1"), zap.Int64("ctx2", 123)},
		}
	}
	defer testHelperReplaceGlobalsToStdout(demoCtxFunc)()

	logger := WithCtx(context.Background(),
		zap.String("k3", "v3"),         // add a new field
		zap.String("ctx2", "override"), // override "ctx2" from context
	)
	logger.Info("example with ctx")

	// Output:
	// {"level":"info","msg":"example with ctx","ctx1":"v1","ctx2":"override","k3":"v3"}
}

func ExampleAddFields() {
	defer testHelperReplaceGlobalsToStdout(nil)()

	ctx := context.Background()
	ctx = AddFields(ctx,
		zap.String("ctx1", "v1"),
		zap.Int64("ctx2", 123))
	ctx = AddFields(ctx,
		zap.String("k3", "v3"),         // add a new field
		zap.String("ctx2", "override"), // override "ctx2"
	)
	logger := WithCtx(ctx)
	logger.Info("example AddFields")

	// Output:
	// {"level":"info","msg":"example AddFields","ctx1":"v1","ctx2":"override","k3":"v3"}
}
