package logid

import (
	"context"
	"net/http"
)

type ctxKey struct{}

// NewCtx returns a new context with a log ID associated.
func NewCtx() context.Context {
	return SetLogID(context.Background(), Gen())
}

// AddLogIDIfNotExists checks log ID from ctx, if there is no log ID
// associated with it, it returns a new context with a new log ID.
func AddLogIDIfNotExists(ctx context.Context) context.Context {
	logId := GetLogID(ctx)
	if logId == "" {
		ctx = SetLogID(ctx, Gen())
	}
	return ctx
}

// GetLogID returns the associated log ID with ctx.
func GetLogID(ctx context.Context) string {
	x, _ := ctx.Value(ctxKey{}).(string)
	return x
}

// SetLogID sets logId to ctx.
func SetLogID(ctx context.Context, logId string) context.Context {
	return context.WithValue(ctx, ctxKey{}, logId)
}

// SetRequestID gets log ID from ctx if available, or generates a new one,
// it sets the log ID to req as header "X-Request-ID".
func SetRequestID(ctx context.Context, req *http.Request) string {
	logId := GetLogID(ctx)
	if logId == "" {
		logId = Gen()
	}
	req.Header.Set("X-Request-ID", logId)
	return logId
}
