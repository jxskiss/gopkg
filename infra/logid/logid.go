package logid

import (
	"encoding/hex"
	"net"
	"strconv"
	"time"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
	"github.com/jxskiss/gopkg/v2/perf/fastrand"
)

// Gen generates a new log ID string using the default generator.
func Gen() string {
	return defaultV1Gen.Gen()
}

var defaultV1Gen *v1Gen
var localIPStr string

func init() {
	defaultV1Gen = &v1Gen{}
	localIPStr = getIPADdr()
}

func getIPADdr() string {
	conn, err := net.Dial("udp", "10.20.30.40:56789")
	if err != nil {
		return IPUnknown
	}
	defer conn.Close()
	ip := conn.LocalAddr().(*net.UDPAddr).IP.To16()
	return hex.EncodeToString(ip)
}

const (
	// IPUnknown represents unknown IP address.
	IPUnknown = "00000000000000000000000000000000"
)

const (
	v1Version = '1'
	v1Length  = 51
	v1RandN   = 1<<24 - 1<<20
)

type v1Gen struct{}

func (p *v1Gen) Gen() string {
	buf := make([]byte, 1, v1Length)
	buf[0] = v1Version

	// milli timestamp, variadic length, max 13 bytes
	t := time.Now().UnixMilli()
	buf = strconv.AppendInt(buf, t, 32)

	// ip address, fixed length, 32 bytes
	buf = append(buf, localIPStr...)

	// random number, fixed length, 5 bytes
	r := fastrand.Int31n(v1RandN) + 1<<20
	buf = strconv.AppendInt(buf, int64(r), 32)

	return unsafeheader.BytesToString(buf)
}
