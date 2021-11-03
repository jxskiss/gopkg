package zlog

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func testHelperReplaceGlobalsToStdout(ctxFunc func(ctx context.Context, args CtxArgs) CtxResult) func() {
	cfg := &Config{
		Level:             "trace",
		Format:            "json",
		DisableTimestamp:  true,
		DisableCaller:     true,
		DisableStacktrace: true,
		GlobalConfig: GlobalConfig{
			CtxFunc: ctxFunc,
		},
	}
	l, p, err := NewWithOutput(cfg, zapcore.AddSync(os.Stdout))
	if err != nil {
		panic(err)
	}
	return ReplaceGlobals(l, p)
}

func ExampleBuilder() {
	defer testHelperReplaceGlobalsToStdout(nil)()

	logger := B(context.TODO()).
		Named("example_builder").
		Method().
		With(zap.String("k1", "v1"), zap.Int64("k2", 54321)).
		Build()
	logger.Info("example builder")

	// Output:
	// {"level":"info","logger":"example_builder","msg":"example builder","methodName":"zlog.ExampleBuilder","k1":"v1","k2":54321}
}

func ExampleBuilder_namespace() {
	defer testHelperReplaceGlobalsToStdout(nil)()

	builder := B(nil).
		With(zap.String("k1", "v1"), zap.String("k2", "v2")).
		With(zap.Namespace("subns"))
	builder = builder.With(zap.String("k1", "sub1"), zap.String("k2", "sub2"))
	builder.Build().Info("example builder namespace")

	// Output:
	// {"level":"info","msg":"example builder namespace","k1":"v1","k2":"v2","subns":{"k1":"sub1","k2":"sub2"}}
}

func ExampleBuilder_newNamespace() {
	defer testHelperReplaceGlobalsToStdout(nil)()

	builder := B(nil)
	builder = builder.With(zap.String("k1", "v1"), zap.String("k2", "v2"))
	builder = builder.With(
		zap.String("k1", "override"),
		zap.Namespace("subns"),
		zap.String("k1", "sub1"), zap.String("k2", "sub2"))
	builder.Build().Info("example builder new namespace")

	// Output:
	// {"level":"info","msg":"example builder new namespace","k1":"override","k2":"v2","subns":{"k1":"sub1","k2":"sub2"}}
}

func ExampleWithBuilder() {
	defer testHelperReplaceGlobalsToStdout(nil)()

	// Make a Builder.
	builder := B(context.TODO()).
		Method().
		With(zap.String("k1", "v1"), zap.Int64("k2", 54321))
	builder.Build().Info("with builder")

	// Pass it to another function or goroutine.
	ctx := WithBuilder(context.Background(), builder)

	func(ctx context.Context) {
		builder := B(ctx). // get Builder from ctx
					Method().                       // override the method name
					With(zap.String("k1", "inner")) // override "k1"

		// do something

		builder.Build().Info("another function")
	}(ctx)

	// Output:
	// {"level":"info","msg":"with builder","methodName":"zlog.ExampleWithBuilder","k1":"v1","k2":54321}
	// {"level":"info","msg":"another function","methodName":"zlog.ExampleWithBuilder.func1","k1":"inner","k2":54321}
}

func ExampleWith() {
	defer testHelperReplaceGlobalsToStdout(nil)()

	With(zap.String("k1", "v1"), zap.Int64("k2", 54321)).
		Info("example with")

	// Output:
	// {"level":"info","msg":"example with","k1":"v1","k2":54321}
}

func ExampleWithCtx() {

	demoCtxFunc := func(ctx context.Context, args CtxArgs) CtxResult {
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
