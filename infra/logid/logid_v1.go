package logid

import (
	"crypto/md5"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/jxskiss/gopkg/v2/internal/machineid"
	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
	"github.com/jxskiss/gopkg/v2/perf/fastrand"
)

const (
	v1Version = '1'
	v1Length  = 34
)

// NewV1Gen creates a new v1 log ID generator.
//
// A v1 log ID is consisted of the following parts:
//
//   - 1 byte version flag "1"
//   - 9 bytes milli timestamp, in base32 form
//   - 16 bytes hash of the machine ID of current host if available,
//     else 16 bytes random data
//   - 8 bytes random data
func NewV1Gen() Generator {
	return &v1Gen{
		machineID: getMachineID(),
	}
}

func getMachineID() [16]byte {
	var machineID [16]byte
	var mID [10]byte
	if x, err := machineid.ID(); err == nil {
		sum := md5.Sum([]byte(x))
		copy(mID[:], sum[:])
	} else {
		_, _ = fastrand.Read(mID[:])
	}
	b32Enc.Encode(machineID[:], mID[:])
	return machineID
}

type v1Gen struct {
	machineID [16]byte
}

func (p *v1Gen) Gen() string {
	buf := make([]byte, 1, v1Length)
	buf[0] = v1Version

	// milli timestamp, fixed length, 9 bytes
	t := time.Now().UnixMilli()
	buf = strconv.AppendInt(buf, t, 32)

	// random bytes, fixed length, 8 bytes
	// 5*8 -> 8*5, use buf[10:15] as temporary buffer
	b := buf[10:15]
	_, _ = fastrand.Read(b)
	b32Enc.Encode(buf[26:34], b)

	// machine ID, fixed length, 16 bytes
	copy(buf[10:26], p.machineID[:])

	buf = buf[:v1Length]
	return unsafeheader.BytesToString(buf)
}

func decodeV1(s string) (info *v1Info) {
	info = &v1Info{}
	if len(s) != v1Length {
		return
	}
	t, err := strconv.ParseInt(s[1:10], 32, 64)
	if err != nil {
		return
	}
	mID := s[10:26]
	r := s[26:v1Length]
	*info = v1Info{
		valid:     true,
		time:      time.UnixMilli(t),
		machineID: mID,
		random:    r,
	}
	return
}

var _ infoInterface = &v1Info{}

type v1Info struct {
	valid     bool
	time      time.Time
	machineID string
	random    string
}

func (info *v1Info) Valid() bool {
	return info != nil && info.valid
}

func (info *v1Info) Version() string { return "1" }

func (info *v1Info) Time() time.Time { return info.time }

func (info *v1Info) IP() net.IP { return nil }

func (info *v1Info) Random() string { return info.random }

func (info *v1Info) String() string {
	if !info.Valid() {
		return "1|invalid"
	}
	return fmt.Sprintf("1|%s|%s|%s", info.time.Format(strTimeMilli), info.machineID, info.random)
}
