package zlog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestB(t *testing.T) {

	demoCtxFunc := func(ctx context.Context, args CtxArgs) (result CtxResult) {
		if v1 := ctx.Value("k1"); v1 != nil {
			result.Fields = append(result.Fields, zap.String("k1", v1.(string)))
		}
		if v2 := ctx.Value("k2"); v2 != nil {
			result.Fields = append(result.Fields, zap.Int("k2", v2.(int)))
		}
		return result
	}
	defer testHelperReplaceGlobalsToStdout(demoCtxFunc)()

	builder0 := B()
	assert.Equal(t, baseBuilder, builder0)

	builder1 := B(context.Background())
	assert.Equal(t, baseBuilder, builder1)

	ctx2 := WithBuilder(context.Background(),
		builder1.With(zap.Int64("some1", 1), zap.String("some2", "value")))
	builder2 := B(ctx2)
	assert.NotEqual(t, builder2, builder1)
	assert.Len(t, builder2.fields, 2)

	ctx3 := WithBuilder(context.Background(), builder2)
	builder3 := B(ctx3)
	assert.Equal(t, builder2, builder3)

	ctx4 := context.WithValue(context.Background(), "k1", "v1")
	ctx4 = context.WithValue(ctx4, "k2", 123)
	builder4 := B(ctx4)
	assert.Len(t, builder4.fields, 2)
}

func BenchmarkZapLoggerWith(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		method, _, _, _ := getCaller(0)
		_ = L().With(
			zap.String("method", method),
			zap.Int64("some1", 1),
			zap.String("some2", "value"),
			zap.String("some3", "value"),
		)
	}
}

func BenchmarkWithMethod(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = WithMethod(
			zap.Int64("some1", 1),
			zap.String("some2", "value"),
			zap.String("some3", "value"),
		)
	}
}

func BenchmarkBuilder(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = B(context.TODO()).Method().With(
			zap.Int64("some1", 1),
			zap.String("some2", "value"),
			zap.String("some3", "value"),
		).Build()
	}
}
