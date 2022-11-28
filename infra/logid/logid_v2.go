package logid

import (
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
	"github.com/jxskiss/gopkg/v2/perf/fastrand"
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
	ipstr := IPUnknown
	if len(ip) > 0 {
		ipstr = hex.EncodeToString(ip.To16())
	}
	return &v2Gen{
		ipstr: ipstr,
	}
}

type v2Gen struct {
	ipstr string
}

func (p *v2Gen) Gen() string {
	buf := make([]byte, 1, v2Length)
	buf[0] = v2Version

	// milli timestamp, fixed length, 9 bytes
	t := time.Now().UnixMilli()
	buf = strconv.AppendInt(buf, t, 32)

	// ip address, fixed length, 32 bytes
	buf = append(buf, p.ipstr...)

	// random number, fixed length, 5 bytes
	r := fastrand.Int31n(v2RandN) + 1<<20
	buf = strconv.AppendInt(buf, int64(r), 32)

	return unsafeheader.BytesToString(buf)
}

func decodeV2(s string) (info *v2Info) {
	info = &v2Info{}
	if len(s) != v2Length {
		return
	}
	t, err := strconv.ParseInt(s[1:10], 32, 64)
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

var _ infoInterface = &v2Info{}

type v2Info struct {
	valid  bool
	time   time.Time
	ip     net.IP
	random string
}

func (info *v2Info) Valid() bool {
	return info != nil && info.valid
}

func (info *v2Info) Version() string { return "2" }

func (info *v2Info) Time() time.Time { return info.time }

func (info *v2Info) IP() net.IP { return info.ip }

func (info *v2Info) Random() string { return info.random }

func (info *v2Info) String() string {
	if !info.Valid() {
		return "2|invalid"
	}
	return fmt.Sprintf("2|%s|%s|%s", info.time.Format(strTimeMilli), info.ip.String(), info.random)
}
