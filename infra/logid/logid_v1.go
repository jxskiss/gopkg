package logid

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
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
		_, err = rand.Read(mID[:])
		if err != nil {
			panic("error calling crypto/rand.Read: " + err.Error())
		}
	}
	b32Enc.Encode(machineID[:], mID[:])
	return machineID
}

type v1Gen struct {
	machineID [16]byte
}

func (p *v1Gen) Gen() string {
	buf := make([]byte, v1Length)
	buf[0] = v1Version

	// milli timestamp, fixed length, 9 bytes
	t := time.Now().UnixMilli()
	encodeBase32(buf[1:10], t)

	// random bytes, fixed length, 8 bytes
	// 5*8 -> 8*5, use buf[10:18] as temporary buffer
	b := buf[10:18]
	*(*uint64)(unsafeheader.SliceData(b)) = fastrand.Uint64()
	b32Enc.Encode(buf[26:34], b[:5])

	// machine ID, fixed length, 16 bytes
	copy(buf[10:26], p.machineID[:])

	return unsafeheader.BytesToString(buf)
}

func decodeV1Info(s string) (info *v1Info) {
	info = &v1Info{}
	if len(s) != v1Length {
		return
	}
	t, err := decodeBase32(s[1:10])
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

var _ V1Info = &v1Info{}

type V1Info interface {
	Info
	Time() time.Time
	MachineID() string
	Random() string
}

type v1Info struct {
	valid     bool
	time      time.Time
	machineID string
	random    string
}

func (info *v1Info) Valid() bool       { return info != nil && info.valid }
func (info *v1Info) Version() byte     { return v1Version }
func (info *v1Info) Time() time.Time   { return info.time }
func (info *v1Info) MachineID() string { return info.machineID }
func (info *v1Info) Random() string    { return info.random }

func (info *v1Info) String() string {
	if !info.Valid() {
		return "1|invalid"
	}
	return fmt.Sprintf("1|%s|%s|%s", formatTime(info.time), info.machineID, info.random)
}
