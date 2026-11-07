package logid

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
)

var _ Generator = &V2Gen{}
var _ V2Info = &v2Info{}

// IPUnknown represents unknown IP address.
const IPUnknown = "0000000"

const V2Version byte = '2'

const (
	v2BaseLength = timestampLength + randomSuffixLength
	v2IPv4Length = v2BaseLength + 7
	v2IPv6Length = v2BaseLength + 26
)

// NewV2Gen creates a new v2 log ID generator.
func NewV2Gen(ip net.IP) *V2Gen {
	ipStr := IPUnknown
	if v4 := ip.To4(); len(v4) > 0 {
		ipStr = b32Enc.EncodeToString(v4)
	} else if len(ip) > 0 {
		ipStr = b32Enc.EncodeToString(ip)
	}
	return &V2Gen{
		ipStr: ipStr,
	}
}

// V2Gen is a v2 log ID generator.
// A v2 log ID is 30 characters long for IPv4 or 49 characters long for IPv6.
//
// A v2 log ID is consisted by the following parts:
//
//   - 9 bytes milli timestamp, in base32 form
//   - 7 bytes IPv4 address, or 26 bytes IPv6 address, in base32 form
//   - 13 bytes base32-encoded 64-bit random data
//   - 1 byte version flag "2"
//
// e.g.
//   - "1hmz5yad60410610fzzzzzzzzzzzz2"
//   - "1hmz5yad6zpyxr0802r01c00000000000jg80000000000002"
type V2Gen struct {
	ipStr string
}

// Gen generates a new log ID string.
func (p *V2Gen) Gen() string {
	buf := make([]byte, v2BaseLength+len(p.ipStr))
	versionOffset := len(buf) - 1
	randomOffset := len(buf) - randomSuffixLength
	buf[versionOffset] = V2Version

	// milli timestamp, fixed length, 9 bytes
	encodeBase32(buf[:timestampLength], uint64(time.Now().UnixMilli()))

	encodeBase32(buf[randomOffset:versionOffset], rand64bits())

	// ip address, 7 bytes for IPv4 or 26 bytes for IPv6
	copy(buf[timestampLength:randomOffset], p.ipStr)

	return unsafeheader.BytesToString(buf)
}

func decodeV2Info(s string) (info *v2Info) {
	info = &v2Info{}
	if len(s) != v2IPv4Length && len(s) != v2IPv6Length {
		return
	}
	versionOffset := len(s) - 1
	randomOffset := len(s) - randomSuffixLength
	r := s[randomOffset:versionOffset]
	tMsec, err := decodeBase32(s[:timestampLength])
	if err != nil {
		return
	}
	ip, err := b32Enc.DecodeString(strings.ToLower(s[timestampLength:randomOffset]))
	if err != nil {
		return
	}
	*info = v2Info{
		valid:  true,
		time:   time.UnixMilli(int64(tMsec)),
		ip:     ip,
		random: r,
	}
	return
}

// V2Info holds parsed information of a v2 log ID string.
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
func (info *v2Info) Version() byte   { return V2Version }
func (info *v2Info) Time() time.Time { return info.time }
func (info *v2Info) IP() net.IP      { return info.ip }
func (info *v2Info) Random() string  { return info.random }

func (info *v2Info) String() string {
	if !info.Valid() {
		return "2|invalid"
	}
	return fmt.Sprintf("2|%s|%s|%s", formatTime(info.time), info.ip.String(), info.random)
}
