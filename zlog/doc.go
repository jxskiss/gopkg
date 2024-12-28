/*
Package zlog lets you use golang structured logging (slog) with context.
Add and retrieve logger to and from context.
Add and retrieve attributes to and from context.
Automatically read any custom context values, such as OpenTelemetry TraceID.

This package borrows code from package `github.com/veqryn/slog-context`.
The reason of forking is that I want opinionated different API design,
the code is simple, it's easy to maintain a different package.
*/
package zlog
