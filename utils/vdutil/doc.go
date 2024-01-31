// Package vdutil helps to do data validation.
//
// It defines an abstract Rule interface to run validating rules,
// some commonly used rules are implemented in this package.
// User can define their own validating rules, for example wrapping
// "github.com/go-playground/validator" and
// "github.com/bytedance/go-tagexpr" to validate structs,
// or validating data passed through context.Context and save the
// result to Result.Data for further accessing.
package vdutil
