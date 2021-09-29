package zlog

import (
	"testing"

	"go.uber.org/zap"
)

func BenchmarkWithMethods(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = WithMethod(
			zap.Int64("some1", 1),
			zap.String("some2", "value"),
			zap.String("some3", "value"),
		)
	}
}

func BenchmarkBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = B().Method().With(
			zap.Int64("some1", 1),
			zap.String("some2", "value"),
			zap.String("some3", "value"),
		).Build()
	}
}
