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

var localIPStr string

func init() {
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
	v1Version = '1'
	v1Length  = 47
	v1RandN   = 1<<25 - 1<<20
)

type v1Gen struct{}

func (p *v1Gen) Gen() string {
	buf := make([]byte, 1, v1Length)
	buf[0] = v1Version

	// milli timestamp, fixed length, 9 bytes
	t := time.Now().UnixMilli()
	buf = strconv.AppendInt(buf, t, 32)

	// ip address, fixed length, 32 bytes
	buf = append(buf, localIPStr...)

	// random number, fixed length, 5 bytes
	r := fastrand.Int31n(v1RandN) + 1<<20
	buf = strconv.AppendInt(buf, int64(r), 32)

	return unsafeheader.BytesToString(buf)
}

func decodeV1(s string) (info *v1Info) {
	info = &v1Info{}
	if len(s) != 47 {
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
	r, err := strconv.ParseInt(s[42:47], 32, 64)
	if err != nil {
		return
	}
	*info = v1Info{
		valid:  true,
		time:   time.UnixMilli(t),
		ip:     ip,
		random: int(r),
	}
	return
}

var _ infoInterface = &v1Info{}

type v1Info struct {
	valid  bool
	time   time.Time
	ip     net.IP
	random int
}

func (info *v1Info) Valid() bool { return info != nil && info.valid }

func (info *v1Info) Version() string { return "1" }

func (info *v1Info) Time() time.Time { return info.time }

func (info *v1Info) IP() net.IP { return info.ip }

func (info *v1Info) Random() int { return info.random }

func (info *v1Info) String() string {
	if !info.Valid() {
		return "1|invalid"
	}
	return fmt.Sprintf("1|%s|%s|%d", info.time.Format(strTimeMilli), info.ip.String(), info.random)
}
