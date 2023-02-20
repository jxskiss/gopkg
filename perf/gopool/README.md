# gopool

## Introduction

`gopool` is a high-performance goroutine pool which aims to reuse goroutines and limit the number of goroutines.

It is an alternative to the `go` keyword.

This package is a fork of `github.com/bytedance/gopkg/util/gopool`.

## Features

- High performance
- Auto-recovering panics
- Limit goroutine numbers
- Reuse goroutine stack

## QuickStart

Just replace your `go func(){...}` with `gopool.Go(func(){...})`.

old:
```go
go func() {
	// do your job
}()
```

new:
```go
gopool.Go(func() {
	// do your job
})

// or with context
gopool.CtxGo(ctx, func() {
	// do your job
})
```

Or create a dedicated pool for specific workload:
```go
myPool := gopool.NewPool(&gopool.Config{
	// configuration
})

myPool.Go(func() {
	// do your job
})
myPool.CtxGo(ctx, func() {
	// do your job
})
```

Or create a pool to execute a handler to process values of a specific type:
```go
myHandler := func(ctx context.Context, arg SomeType) {
	// do your job
}
myPool := gopool.NewTypedPool(myHandler, &gopool.Config{
	// configuration
})

myPool.Go(varOfSomeType)
myPool.CtxGo(ctx, varOfSomeType)
```

See package doc for more information.
