package zlog

import "go.uber.org/zap"

// Field is an alias type of zap.Field.
type Field = zap.Field

// Any is an alias function to zap.Any.
// See zap.Any for detailed documentation.
func Any(key string, value any) Field {
	return zap.Any(key, value)
}
