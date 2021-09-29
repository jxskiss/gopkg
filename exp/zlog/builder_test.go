package zlog

import (
	"testing"

	"go.uber.org/zap"
)

func BenchmarkZapLoggerWith(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		method, _ := getFunctionName(0)
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
		_ = B().Method().With(
			zap.Int64("some1", 1),
			zap.String("some2", "value"),
			zap.String("some3", "value"),
		).Build()
	}
}
