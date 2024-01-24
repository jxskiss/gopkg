package logid

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/jxskiss/gopkg/v2/internal/machineid"
	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
)

var _ Generator = &V1Gen{}
var _ V1Info = &v1Info{}

const (
	v1Version = '1'
	v1Length  = 36
)

// NewV1Gen creates a new v1 log ID generator.
func NewV1Gen() *V1Gen {
	return &V1Gen{
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

// V1Gen is a v1 log ID generator.
//
// A v1 log ID is consisted of the following parts:
//
//   - 9 bytes milli timestamp, in base32 form
//   - 16 bytes hash of the machine ID of current host if available,
//     else 16 bytes random data
//   - 10 bytes random data
//   - 1 byte version flag "1"
//
// e.g. "1HMZ5YAD5M0RY2MKE72XWXGSW140NFEAD8J1"
type V1Gen struct {
	machineID [16]byte
}

// Gen generates a new log ID string.
func (p *V1Gen) Gen() string {
	buf := make([]byte, v1Length)
	buf[len(buf)-1] = v1Version

	// milli timestamp, fixed length, 9 bytes
	encodeBase32(buf[0:9], time.Now().UnixMilli())

	// random bytes, fixed length, 10 bytes
	randNum := rand50bits()
	encodeBase32(buf[25:35], randNum)

	// machine ID, fixed length, 16 bytes
	copy(buf[9:25], p.machineID[:])

	return unsafeheader.BytesToString(buf)
}

func decodeV1Info(s string) (info *v1Info) {
	info = &v1Info{}
	if len(s) != v1Length {
		return
	}
	mID := s[9:25]
	r := s[25:35]
	tMsec, err := decodeBase32(s[:9])
	if err != nil {
		return
	}
	*info = v1Info{
		valid:     true,
		time:      time.UnixMilli(tMsec),
		machineID: mID,
		random:    r,
	}
	return
}

// V1Info holds parsed information of a v1 log ID string.
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
