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
const IPUnknown = "0000000"

const (
	v2Version    = '2'
	v2IPv4Length = 27
	v2IPv6Length = 46
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
//   - 9 bytes milli timestamp, in base32 form
//   - 7 bytes IPv4 address, or 26 bytes IPv6 address, in base32 form
//   - 10 bytes random data
//   - 1 byte version flag "2"
//
// e.g.
//   - "1HMZ5YAD6041061072VFVV7C2J2"
//   - "1HMZ5YAD6ZPYXR0802R01C00000000000JGCBDDWZEJH42"
type V2Gen struct {
	ipStr string
}

// Gen generates a new log ID string.
func (p *V2Gen) Gen() string {
	buf := make([]byte, 20+len(p.ipStr))
	buf[len(buf)-1] = v2Version

	// milli timestamp, fixed length, 9 bytes
	encodeBase32(buf[0:9], time.Now().UnixMilli())

	// random bytes, fixed length, 10 bytes
	randNum := rand50bits()
	encodeBase32(buf[len(buf)-11:len(buf)-1], randNum)

	// ip address, 7 bytes for IPv4 or 26 bytes for IPv6
	copy(buf[9:], p.ipStr)

	return unsafeheader.BytesToString(buf)
}

func decodeV2Info(s string) (info *v2Info) {
	info = &v2Info{}
	if len(s) != v2IPv4Length && len(s) != v2IPv6Length {
		return
	}
	r := s[len(s)-11 : len(s)-1]
	tMsec, err := decodeBase32(s[:9])
	if err != nil {
		return
	}
	ip, err := b32Enc.DecodeString(s[9 : len(s)-11])
	if err != nil {
		return
	}
	*info = v2Info{
		valid:  true,
		time:   time.UnixMilli(tMsec),
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
