package acache

type Fetcher interface {
	Fetch(key string) (any, error)
}

type BatchFetcher interface {
	Fetcher
	BatchSize() int
	BatchFetch(keys []string) (map[string]any, error)
}

// FuncFetcher is a function that implements the interface Fetcher.
type FuncFetcher func(key string) (any, error)

func (f FuncFetcher) Fetch(key string) (any, error) {
	return f(key)
}
