package logid

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/jxskiss/gopkg/v2/internal/fastrand"
	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
)

const (
	// IPUnknown represents unknown IP address.
	IPUnknown = "00000000000000000000000000000000"
)

const (
	v2Version = '2'
	v2Length  = 47
	v2RandN   = 1<<25 - 1<<20
)

// NewV2Gen creates a new v2 log ID generator.
//
// A v2 log ID is consisted of the following parts:
//
//   - 1 byte version flag "2"
//   - 9 bytes milli timestamp, in base32 form
//   - 32 bytes IP address, in hex form
//   - 5 bytes random data
func NewV2Gen(ip net.IP) Generator {
	ipStr := IPUnknown
	if ip = ip.To16(); len(ip) > 0 {
		ipStr = strings.ToUpper(hex.EncodeToString(ip))
	}
	return &v2Gen{
		ipStr: ipStr,
	}
}

type v2Gen struct {
	ipStr string
}

func (p *v2Gen) Gen() string {
	buf := make([]byte, v2Length)
	buf[0] = v2Version

	// milli timestamp, fixed length, 9 bytes
	t := time.Now().UnixMilli()
	encodeBase32(buf[1:10], t)

	// ip address, fixed length, 32 bytes
	copy(buf[10:42], p.ipStr)

	// random number, fixed length, 5 bytes
	r := fastrand.N(int64(v2RandN)) + 1<<20
	encodeBase32(buf[42:47], r)

	return unsafeheader.BytesToString(buf)
}

func decodeV2Info(s string) (info *v2Info) {
	info = &v2Info{}
	if len(s) != v2Length {
		return
	}
	t, err := decodeBase32(s[1:10])
	if err != nil {
		return
	}
	ip, err := hex.DecodeString(s[10:42])
	if err != nil {
		return
	}
	r := s[42:47]
	*info = v2Info{
		valid:  true,
		time:   time.UnixMilli(t),
		ip:     ip,
		random: r,
	}
	return
}

var _ V2Info = &v2Info{}

type V2Info interface {
	Info
	Time() time.Time
	IP() net.IP
	Random() string
}

type v2Info struct {
	valid  bool
	time   time.Time
	ip     net.IP
	random string
}

func (info *v2Info) Valid() bool     { return info != nil && info.valid }
func (info *v2Info) Version() byte   { return v2Version }
func (info *v2Info) Time() time.Time { return info.time }
func (info *v2Info) IP() net.IP      { return info.ip }
func (info *v2Info) Random() string  { return info.random }

func (info *v2Info) String() string {
	if !info.Valid() {
		return "2|invalid"
	}
	return fmt.Sprintf("2|%s|%s|%s", formatTime(info.time), info.ip.String(), info.random)
}
