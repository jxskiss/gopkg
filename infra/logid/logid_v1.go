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
	v1Length  = 44
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
//   - 17 bytes milli timestamp, in UTC timezone
//   - 16 bytes hash of the machine ID of current host if available,
//     else 16 bytes random data
//   - 10 bytes random data, with 1 bit to mark UTC timezone
//   - 1 byte version flag "1"
type V1Gen struct {
	machineID [16]byte
	useUTC    bool
}

// UseUTC sets the generator to format timestamp with location time.UTC.
// By default, it formats timestamp with location time.Local.
func (p *V1Gen) UseUTC() *V1Gen {
	p.useUTC = true
	return p
}

// Gen generates a new log ID string.
func (p *V1Gen) Gen() string {
	buf := make([]byte, v1Length)
	buf[len(buf)-1] = v1Version

	// milli timestamp, fixed length, 17 bytes
	appendTime(buf[:0], time.Now(), p.useUTC)

	// random bytes, fixed length, 10 bytes
	randNum := rand50bitsWithUTCMark(p.useUTC)
	encodeBase32(buf[33:43], randNum)

	// machine ID, fixed length, 16 bytes
	copy(buf[17:33], p.machineID[:])

	return unsafeheader.BytesToString(buf)
}

func decodeV1Info(s string) (info *v1Info) {
	info = &v1Info{}
	if len(s) != v1Length {
		return
	}
	mID := s[17:33]
	r := s[33:43]
	isUTC := checkUTCMark(r)
	t, err := decodeTime(s[:17], isUTC)
	if err != nil {
		return
	}
	*info = v1Info{
		valid:     true,
		time:      t,
		machineID: mID,
		random:    r,
	}
	return
}

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
