package zlog

import (
	"context"
	"log/slog"
	"slices"
)

type ctxAttrKey int

const (
	prependKey ctxAttrKey = 1
	appendKey  ctxAttrKey = 2
)

// PrependAttrs adds the attribute arguments to the end of the group that
// will be prepended to the start of the log record when it is handled.
// The attributes will be at the root level, and not in any groups.
func PrependAttrs(parent context.Context, args ...any) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	if v, ok := parent.Value(prependKey).([]slog.Attr); ok {
		attrs := append(slices.Clip(v), ArgsToAttrSlice(args...)...)
		return context.WithValue(parent, prependKey, attrs)
	}
	return context.WithValue(parent, prependKey, ArgsToAttrSlice(args...))
}

// ExtractPrepended is an AttrExtractor that returns the prepended attributes
// stored in the context. The returned attr should not be modified in any way,
// doing so will cause a race condition.
func ExtractPrepended(ctx context.Context, _ *slog.Record) slog.Attr {
	if v, ok := ctx.Value(prependKey).([]slog.Attr); ok {
		return slog.Attr{Value: slog.GroupValue(v...)}
	}
	return slog.Attr{}
}

// AppendAttrs adds the attribute arguments to the end of the group that
// will be appended to the end of the log record when it is handled.
// The attributes could be in a group or subgroup, if the log has used
// WithGroup at some point.
func AppendAttrs(parent context.Context, args ...any) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	if v, ok := parent.Value(appendKey).([]slog.Attr); ok {
		attrs := append(slices.Clip(v), ArgsToAttrSlice(args...)...)
		return context.WithValue(parent, appendKey, attrs)
	}
	return context.WithValue(parent, appendKey, ArgsToAttrSlice(args...))
}

// ExtractAppended is an AttrExtractor that returns the appended attributes
// stored in the context. The returned attr should not be modified in any way,
// doing so will cause a race condition.
func ExtractAppended(ctx context.Context, _ *slog.Record) slog.Attr {
	if v, ok := ctx.Value(appendKey).([]slog.Attr); ok {
		return slog.Attr{Value: slog.GroupValue(v...)}
	}
	return slog.Attr{}
}

// This is copied from golang package log/slog.
const badKey = "!BADKEY"

// ArgsToAttrSlice turns a slice of arguments, some of which pairs of primitives,
// some might be attributes already, into a slice of attributes.
func ArgsToAttrSlice(args ...any) []slog.Attr {
	if len(args) == 1 {
		if attrs, ok := args[0].([]slog.Attr); ok {
			return attrs
		}
		if arg0Slice, ok := args[0].([]any); ok {
			return ArgsToAttrSlice(arg0Slice...)
		}
	}

	var (
		attr  slog.Attr
		attrs []slog.Attr
	)
	for len(args) > 0 {
		attr, args = argsToAttr(args)
		attrs = append(attrs, attr)
	}
	return attrs
}

// argsToAttr turns a prefix of the nonempty args slice into an Attr
// and returns the unconsumed portion of the slice.
// If args[0] is a slog.Attr, it returns it.
// if args[0] is []slog.Attr, it returns a group Attr.
// If args[0] is a string, it treats the first two elements as
// a key-value pair.
// Otherwise, it treats args[0] as a value with a missing key.
func argsToAttr(args []any) (slog.Attr, []any) {
	switch x := args[0].(type) {
	case string:
		if len(args) == 1 {
			return slog.String(badKey, x), nil
		}
		return slog.Any(x, args[1]), args[2:]

	case slog.Attr:
		return x, args[1:]

	case []slog.Attr:
		return slog.Attr{Value: slog.GroupValue(x...)}, args[1:]

	default:
		return slog.Any(badKey, x), args[1:]
	}
}
