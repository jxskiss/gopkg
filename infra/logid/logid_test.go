package logid

import (
	"net"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefault(t *testing.T) {
	id1 := Gen()
	assert.Len(t, id1, v1Length)

	SetDefault(NewV2Gen(nil))
	defer SetDefault(NewV1Gen())
	id2 := Gen()
	assert.Len(t, id2, v2IPv4Length)
}

func Test_encodeBase32(t *testing.T) {
	for i := 0; i < 1000; i++ {
		var buf = make([]byte, 13)
		x := rand64bits()
		encodeBase32(buf, x)
		got, err := decodeBase32(string(buf))
		assert.Nil(t, err)
		assert.Equal(t, x, got)
	}
}

func Test_encodeBase32FullWidth(t *testing.T) {
	for _, tc := range []struct {
		x    uint64
		want string
	}{
		{0, "0000000000000"},
		{1<<63 - 1, "7zzzzzzzzzzzz"},
		{1 << 63, "8000000000000"},
		{1<<64 - 1, "fzzzzzzzzzzzz"},
	} {
		buf := make([]byte, 13)
		encodeBase32(buf, tc.x)
		assert.Equal(t, tc.want, string(buf))
		got, err := decodeBase32(string(buf))
		require.NoError(t, err)
		assert.Equal(t, tc.x, got)
	}
}

func TestGeneratorsLowercase(t *testing.T) {
	for _, tc := range []struct {
		name   string
		gen    Generator
		length int
	}{
		{"v1", NewV1Gen(), 39},
		{"v2_unknown", NewV2Gen(nil), 30},
		{"v2_ipv4", NewV2Gen(net.ParseIP("192.168.254.253")), 30},
		{"v2_ipv6", NewV2Gen(net.ParseIP("fdbd:dc01:16:16::94")), 49},
	} {
		t.Run(tc.name, func(t *testing.T) {
			id := tc.gen.Gen()
			assert.Len(t, id, tc.length)
			assert.Equal(t, strings.ToLower(id), id)
			assert.True(t, Decode(id).Valid())
		})
	}
}

func TestDecodeCaseInsensitive(t *testing.T) {
	for _, tc := range []struct {
		name string
		id   string
		ip   string
	}{
		{"v1", "1HMZ5YAD5M0RY2MKE72XWXGSWFZZZZZZZZZZZZ1", ""},
		{"v2_ipv4", "1HMZ5YAD5R2MFXZ8FZZZZZZZZZZZZ2", "192.168.254.253"},
		{"v2_ipv6", "1HMZ5YAD5ZPYXR0802R01C00000000000JGFZZZZZZZZZZZZ2", "fdbd:dc01:16:16::94"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mixed := []byte(tc.id)
			for i := 0; i < len(mixed); i += 2 {
				if mixed[i] >= 'A' && mixed[i] <= 'Z' {
					mixed[i] += 'a' - 'A'
				}
			}
			for _, id := range []string{tc.id, strings.ToLower(tc.id), string(mixed)} {
				info := Decode(id)
				require.True(t, info.Valid(), "id: %s", id)
				if tc.ip == "" {
					v1 := info.(V1Info)
					assert.Equal(t, int64(1706148440485), v1.Time().UnixMilli())
					assert.Equal(t, id[9:25], v1.MachineID())
					assert.Equal(t, id[25:38], v1.Random())
				} else {
					v2 := info.(V2Info)
					assert.Equal(t, int64(1706148440485), v2.Time().UnixMilli())
					assert.Equal(t, tc.ip, v2.IP().String())
					assert.Equal(t, id[len(id)-14:len(id)-1], v2.Random())
				}
			}
		})
	}
}

func TestUniqueness(t *testing.T) {
	var count int32 = -1
	var got = make([]string, 1000)

	var wg sync.WaitGroup
	var n = runtime.GOMAXPROCS(0)
	for j := 0; j < n; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				i := atomic.AddInt32(&count, 1)
				if int(i) < len(got) {
					got[i] = Gen()
					continue
				}
				break
			}
		}()
	}
	wg.Wait()

	sort.Strings(got)
	dupCount := 0
	for i := 0; i < len(got)-1; i++ {
		if got[i] == got[i+1] {
			dupCount++
		}
	}
	assert.True(t, dupCount < 2)
}
