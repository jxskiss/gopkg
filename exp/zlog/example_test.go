package zlog

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func testHelperReplaceGlobalToStdout(ctxFunc func(ctx context.Context, args CtxArgs) CtxResult) func() {
	oldL, oldP := gL, gP
	cfg := &Config{
		Level:             "trace",
		Format:            "logfmt",
		DisableTimestamp:  true,
		DisableCaller:     true,
		DisableStacktrace: true,
		CtxFunc:           ctxFunc,
	}
	l, p, err := NewLoggerWithSyncer(cfg, zapcore.AddSync(os.Stdout))
	if err != nil {
		panic(err)
	}
	ReplaceGlobals(l, p)
	return func() {
		ReplaceGlobals(oldL, oldP)
	}
}

func ExampleBuilder() {
	defer testHelperReplaceGlobalToStdout(nil)()

	logger := B().
		Named("example_builder").
		Method().
		With(zap.String("k1", "v1"), zap.Int64("k2", 54321)).
		Build()
	logger.Info("example builder")

	// Output:
	// level=info logger=example_builder msg="example builder" method=zlog.ExampleBuilder k1=v1 k2=54321
}

func ExampleWithBuilder() {
	defer testHelperReplaceGlobalToStdout(nil)()

	// Make a Builder.
	builder := B().
		Method().
		With(zap.String("k1", "v1"), zap.Int64("k2", 54321))
	builder.Build().Info("with builder")

	// Pass it to another function or goroutine.
	ctx := WithBuilder(context.Background(), builder)

	func(ctx context.Context) {
		builder := B().
			Ctx(ctx).                       // get Builder from ctx
			Method().                       // override the method name
			With(zap.String("k1", "inner")) // override "k1"

		// do something

		builder.Build().Info("another function")
	}(ctx)

	// Output:
	// level=info msg="with builder" method=zlog.ExampleWithBuilder k1=v1 k2=54321
	// level=info msg="another function" method=zlog.ExampleWithBuilder.func1 k1=inner k2=54321
}

func ExampleWith() {
	defer testHelperReplaceGlobalToStdout(nil)()

	With(zap.String("k1", "v1"), zap.Int64("k2", 54321)).
		Info("example with")

	// Output:
	// level=info msg="example with" k1=v1 k2=54321
}

func ExampleWithCtx() {

	demoCtxFunc := func(ctx context.Context, args CtxArgs) CtxResult {
		return CtxResult{
			Fields: []zap.Field{zap.String("ctx1", "v1"), zap.Int64("ctx2", 123)},
		}
	}
	defer testHelperReplaceGlobalToStdout(demoCtxFunc)()

	logger := WithCtx(context.Background(),
		zap.String("k3", "v3"),         // add a new field
		zap.String("ctx2", "override"), // override "ctx2" from context
	)
	logger.Info("example with ctx")

	// Output:
	// level=info msg="example with ctx" ctx1=v1 ctx2=override k3=v3
}
