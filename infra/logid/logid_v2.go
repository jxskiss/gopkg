package logid

import (
	"fmt"
	"net"
	"time"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
)

var _ Generator = &V2Gen{}
var _ V2Info = &v2Info{}

// IPUnknown represents unknown IP address.
const IPUnknown = "AAAAAAA"

const (
	v2Version    = '2'
	v2IPv4Length = 35
	v2IPv6Length = 54
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
//
// A v2 log ID is consisted by the following parts:
//
//   - 17 bytes milli timestamp, in UTC timezone
//   - 26 bytes IPv6 address, in base32 form,
//     or 7 bytes IPv4 address, in base32 form
//   - 10 bytes random data
//   - 1 byte version flag "2"
type V2Gen struct {
	ipStr  string
	useUTC bool
}

// UseUTC sets the generator to format timestamp with location time.UTC.
// By default, it formats timestamp with location time.Local.
func (p *V2Gen) UseUTC() *V2Gen {
	p.useUTC = true
	return p
}

// Gen generates a new log ID string.
func (p *V2Gen) Gen() string {
	buf := make([]byte, 28+len(p.ipStr))
	buf[len(buf)-1] = v2Version

	// milli timestamp, fixed length, 17 bytes
	appendTime(buf[:0], time.Now(), p.useUTC)

	// random bytes, fixed length, 10 bytes
	randNum := rand50bitsWithUTCMark(p.useUTC)
	encodeBase32(buf[len(buf)-11:len(buf)-1], randNum)

	// ip address, fixed length, 32 bytes
	copy(buf[17:], p.ipStr)

	return unsafeheader.BytesToString(buf)
}

func decodeV2Info(s string) (info *v2Info) {
	info = &v2Info{}
	if len(s) != v2IPv4Length && len(s) != v2IPv6Length {
		return
	}
	r := s[len(s)-11 : len(s)-1]
	isUTC := checkUTCMark(r)
	t, err := decodeTime(s[:17], isUTC)
	if err != nil {
		return
	}
	ip, err := b32Enc.DecodeString(s[17 : len(s)-11])
	if err != nil {
		return
	}
	*info = v2Info{
		valid:  true,
		time:   t,
		ip:     ip,
		random: r,
	}
	return
}

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
