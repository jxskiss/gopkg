package ezkv

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jxskiss/gopkg/v2/utils/compress"
)

func writeCompressionModels[K comparable, V Model](t *testing.T, method string, cache *ModelCache[K, V], models []V) {
	t.Helper()
	ctx := context.Background()
	switch method {
	case "Set":
		for _, model := range models {
			require.NoError(t, cache.Set(ctx, cache.config.IDFunc(model), model, 0))
		}
	case "BatchSetSlice":
		require.NoError(t, cache.BatchSetSlice(ctx, models, 0))
	case "BatchSetMap":
		values := make(map[K]V, len(models))
		for _, model := range models {
			values[cache.config.IDFunc(model)] = model
		}
		require.NoError(t, cache.BatchSetMap(ctx, values, 0))
	default:
		t.Fatalf("unknown write method %q", method)
	}
}

func TestModelCacheCompressionWrites(t *testing.T) {
	ctx := context.Background()
	models := []*TestModel{{IntId: 111, StrId: "aaa"}, testModelList[1]}
	for _, method := range []string{"Set", "BatchSetSlice", "BatchSetMap"} {
		for _, mode := range []string{"default", "strict", "disabled", "nil"} {
			t.Run(method+"/"+mode, func(t *testing.T) {
				calls := 0
				c, err := compress.NewCompressor(compress.CompressorConfig{
					Threshold:        64,
					DisableUnframed:  mode == "strict",
					CompressCallback: func(context.Context, compress.CompressionStats) { calls++ },
				})
				require.NoError(t, err)
				cache := makeTestingCache(t.Name(), func(m *TestModel) int64 { return m.IntId })
				cache.config.Compressor = c
				cache.config.DisableCompressionWrites = mode == "disabled"
				if mode == "nil" {
					cache.config.Compressor = nil
				}
				writeCompressionModels(t, method, cache, models)
				wantCalls := len(models)
				if mode == "disabled" || mode == "nil" {
					wantCalls = 0
				}
				require.Equal(t, wantCalls, calls)
				for i, model := range models {
					raw, err := model.MarshalBinary()
					require.NoError(t, err)
					stored, err := cache.config.Storage(ctx).Get(ctx, cache.config.KeyFunc(model.IntId))
					require.NoError(t, err)
					require.Len(t, stored, 1)
					if mode == "disabled" || mode == "nil" || (mode == "default" && i == 0) {
						require.Equal(t, raw, stored[0])
					} else {
						require.NotEqual(t, raw, stored[0])
						decoded, alg, err := c.Decompress(ctx, stored[0])
						require.NoError(t, err)
						require.Equal(t, raw, decoded)
						if i == 1 {
							require.Equal(t, compress.TypeGzip, alg)
						} else {
							require.Equal(t, compress.TypeNone, alg)
						}
					}
					got, err := cache.Get(ctx, model.IntId)
					require.NoError(t, err)
					require.Equal(t, model, got)
				}
				got, err := cache.BatchGetSlice(ctx, []int64{111, 112})
				require.NoError(t, err)
				require.ElementsMatch(t, models, got)
			})
		}
	}
}

func TestModelCacheCompressionSwitch(t *testing.T) {
	ctx := context.Background()
	base := makeTestingCache(t.Name(), func(m *TestModel) int64 { return m.IntId })
	require.NoError(t, base.Set(ctx, 111, testModelList[0], 0))
	c, err := compress.NewCompressor(compress.CompressorConfig{})
	require.NoError(t, err)
	cfg := *base.config
	cfg.Compressor = c
	writer := NewModelCache(&cfg)
	require.NoError(t, writer.Set(ctx, 112, testModelList[1], 0))
	rawCfg := cfg
	rawCfg.DisableCompressionWrites = true
	reader := NewModelCache(&rawCfg)
	for _, cache := range []*ModelCache[int64, *TestModel]{writer, reader} {
		for _, model := range testModelList {
			got, err := cache.Get(ctx, model.IntId)
			require.NoError(t, err)
			require.Equal(t, model, got)
		}
		got, err := cache.BatchGetMap(ctx, []int64{111, 112})
		require.NoError(t, err)
		require.Equal(t, map[int64]*TestModel{111: testModelList[0], 112: testModelList[1]}, got)
	}
	require.NoError(t, reader.Set(ctx, 111, testModelList[0], 0))
	got, err := writer.Get(ctx, 111)
	require.NoError(t, err)
	require.Equal(t, testModelList[0], got)
}

type compressionBlob struct {
	data    []byte
	decoded bool
}

func (m *compressionBlob) MarshalBinary() ([]byte, error) { return m.data, nil }
func (m *compressionBlob) UnmarshalBinary(data []byte) error {
	m.data = bytes.Clone(data)
	m.decoded = true
	return nil
}

func TestModelCacheCompressionEmptyAndFramedInput(t *testing.T) {
	ctx := context.Background()
	frame := compress.DefaultCompressor.Compress(ctx, nil)
	require.NoError(t, frame.Err)
	require.NotEmpty(t, frame.Data)
	for _, method := range []string{"Set", "BatchSetSlice", "BatchSetMap"} {
		for _, mode := range []string{"enabled", "disabled", "nil"} {
			for _, input := range []struct {
				name string
				data []byte
			}{
				{"nil", nil}, {"empty", []byte{}},
				{"framed", frame.Data},
			} {
				t.Run(method+"/"+mode+"/"+input.name, func(t *testing.T) {
					cache := makeTestingCache(t.Name(), func(*compressionBlob) int { return 1 })
					cache.config.Compressor = compress.DefaultCompressor
					cache.config.DisableCompressionWrites = mode == "disabled"
					if mode == "nil" {
						cache.config.Compressor = nil
					}
					writeCompressionModels(t, method, cache, []*compressionBlob{{data: input.data}})
					stored, err := cache.config.Storage(ctx).Get(ctx, cache.config.KeyFunc(1))
					require.NoError(t, err)
					require.Len(t, stored, 1)
					if mode == "enabled" {
						require.NotEmpty(t, stored[0])
						got, err := cache.Get(ctx, 1)
						require.NoError(t, err)
						require.True(t, got.decoded)
						require.True(t, bytes.Equal(input.data, got.data))
						batch, err := cache.BatchGetMap(ctx, []int{1})
						require.NoError(t, err)
						require.Equal(t, map[int]*compressionBlob{1: got}, batch)
					} else {
						require.Equal(t, input.data, stored[0])
						if len(input.data) == 0 {
							got, err := cache.Get(ctx, 1)
							require.ErrorIs(t, err, ErrDataNotFound)
							require.Nil(t, got)
							batch, err := cache.BatchGetSlice(ctx, []int{1})
							require.NoError(t, err)
							require.Empty(t, batch)
						}
					}
				})
			}
		}
	}
}

type compressionTestStorage struct {
	mu sync.Mutex
	memoryStorage
	written chan struct{}
}

func (s *compressionTestStorage) Get(ctx context.Context, keys ...string) ([][]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.memoryStorage.Get(ctx, keys...)
}

func (s *compressionTestStorage) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return s.BatchSet(ctx, []string{key}, [][]byte{value}, expiration)
}

func (s *compressionTestStorage) BatchSet(ctx context.Context, keys []string, values [][]byte, expiration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.memoryStorage.BatchSet(ctx, keys, values, expiration)
	s.written <- struct{}{}
	return err
}

func (s *compressionTestStorage) Delete(ctx context.Context, keys ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.memoryStorage.Delete(ctx, keys...)
}

func TestModelCacheCompressionLoader(t *testing.T) {
	ctx := context.Background()
	for _, async := range []bool{false, true} {
		for _, batch := range []bool{false, true} {
			for _, disabled := range []bool{false, true} {
				t.Run(fmt.Sprintf("async=%v/batch=%v/disabled=%v", async, batch, disabled), func(t *testing.T) {
					stor := &compressionTestStorage{memoryStorage: memoryStorage{data: make(map[string][]byte)}, written: make(chan struct{}, 1)}
					cache := makeTestingCache(t.Name(), func(m *TestModel) int64 { return m.IntId })
					cache.config.Storage = func(context.Context) Storage { return stor }
					c, err := compress.NewCompressor(compress.CompressorConfig{})
					require.NoError(t, err)
					cache.config.Compressor = c
					cache.config.DisableCompressionWrites = disabled
					cache.config.CacheLoaderResultAsync = async
					cache.config.Loader = func(context.Context, []int64) (map[int64]*TestModel, error) {
						return map[int64]*TestModel{111: testModelList[0]}, nil
					}
					if batch {
						got, err := cache.BatchGetMap(ctx, []int64{111})
						require.NoError(t, err)
						require.Equal(t, map[int64]*TestModel{111: testModelList[0]}, got)
					} else {
						got, err := cache.Get(ctx, 111)
						require.NoError(t, err)
						require.Equal(t, testModelList[0], got)
					}
					select {
					case <-stor.written:
					case <-time.After(5 * time.Second):
						t.Fatal("loader result was not stored")
					}
					stored, err := stor.Get(ctx, cache.config.KeyFunc(111))
					require.NoError(t, err)
					raw, err := testModelList[0].MarshalBinary()
					require.NoError(t, err)
					if disabled {
						require.Equal(t, raw, stored[0])
					} else {
						decoded, alg, err := cache.config.Compressor.Decompress(ctx, stored[0])
						require.NoError(t, err)
						require.Equal(t, compress.TypeGzip, alg)
						require.Equal(t, raw, decoded)
					}
				})
			}
		}
	}
}

func TestModelCacheCompressionInvalidFrame(t *testing.T) {
	ctx := context.Background()
	for _, batch := range []bool{false, true} {
		for _, loader := range []bool{false, true} {
			t.Run(fmt.Sprintf("batch=%v/loader=%v", batch, loader), func(t *testing.T) {
				cache := makeTestingCache(t.Name(), func(m *TestModel) int64 { return m.IntId })
				cache.config.Compressor = compress.DefaultCompressor
				cache.config.DisableCompressionWrites = true
				var logged []error
				cache.config.ErrorLogger = func(_ context.Context, err error, _ string) { logged = append(logged, err) }
				loads := 0
				if loader {
					cache.config.Loader = func(_ context.Context, pks []int64) (map[int64]*TestModel, error) {
						loads++
						require.Equal(t, []int64{111}, pks)
						return map[int64]*TestModel{111: testModelList[0]}, nil
					}
				}
				frame := compress.DefaultCompressor.Compress(ctx, nil).Data
				require.NoError(t, cache.config.Storage(ctx).Set(ctx, cache.config.KeyFunc(111), frame[:len(frame)-1], 0))
				if batch {
					got, err := cache.BatchGetMap(ctx, []int64{111})
					require.NoError(t, err)
					if loader {
						require.Equal(t, map[int64]*TestModel{111: testModelList[0]}, got)
					} else {
						require.Empty(t, got)
					}
				} else {
					got, err := cache.Get(ctx, 111)
					if loader {
						require.NoError(t, err)
						require.Equal(t, testModelList[0], got)
					} else {
						require.ErrorIs(t, err, compress.ErrInvalidFrame)
						require.Nil(t, got)
					}
				}
				if batch || loader {
					require.Len(t, logged, 1)
					require.ErrorIs(t, logged[0], compress.ErrInvalidFrame)
				}
				if loader {
					require.Equal(t, 1, loads)
				} else {
					require.Zero(t, loads)
				}
			})
		}
	}
}
