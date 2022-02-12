package obscure

import (
	"crypto/md5"
	"encoding/base32"
	"encoding/binary"
	"errors"

	"github.com/jxskiss/gopkg/v2/fastrand"
	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
)

const (
	idxlen  = 61
	encbase = 32
	chars62 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

var ErrInvalidInput = errors.New("obscure: invalid input")

var binEnc = binary.BigEndian

func getRandomChars(rand *fastrand.PCG64, dst []byte) {
	chars := []byte(chars62)
	rand.Shuffle(len(chars), func(i, j int) {
		chars[i], chars[j] = chars[j], chars[i]
	})
	copy(dst, chars)
}

func fnvHash32(buf []byte) uint32 {
	const offset32 = 2166136261
	const prime32 = 16777619

	var hash uint32 = offset32
	for _, c := range buf {
		hash *= prime32
		hash ^= uint32(c)
	}
	return hash
}

type Obscure struct {
	idxChars  [idxlen]byte
	idxdec    [128]int
	table     [idxlen][encbase]byte
	encodings [idxlen]*base32.Encoding
	rand      *fastrand.PCG64
}

func New(key []byte) *Obscure {
	hash := md5.Sum(key)
	rand := fastrand.NewPCG64()
	hi, lo := binEnc.Uint64(hash[:8]), binEnc.Uint64(hash[8:16])
	rand.Seed(hi, lo)
	obs := &Obscure{
		rand: rand,
	}
	getRandomChars(rand, obs.idxChars[:])
	for i := 0; i < idxlen; i++ {
		obs.idxdec[obs.idxChars[i]] = i
		getRandomChars(rand, obs.table[i][:])
		obs.encodings[i] = base32.NewEncoding(string(obs.table[i][:])).WithPadding(base32.NoPadding)
	}
	return obs
}

func (p *Obscure) Index() string {
	return string(p.idxChars[:])
}

func (p *Obscure) Table() [61]string {
	var out [61]string
	for i := 0; i < idxlen; i++ {
		out[i] = string(p.table[i][:])
	}
	return out
}

func (p *Obscure) EncodedLen(n int) int {
	if n <= 0 {
		return 0
	}
	return 1 + p.encodings[0].EncodedLen(n)
}

func (p *Obscure) Encode(dst, src []byte) {
	if len(src) == 0 {
		return
	}
	idx := fnvHash32(src) % idxlen
	idxchar := p.idxChars[idx]
	enc := p.encodings[idx]
	dst[0] = idxchar
	enc.Encode(dst[1:], src)
}

func (p *Obscure) EncodeToBytes(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}
	dst := make([]byte, p.EncodedLen(len(src)))
	p.Encode(dst, src)
	return dst
}

func (p *Obscure) EncodeToString(src []byte) string {
	if len(src) == 0 {
		return ""
	}
	dst := p.EncodeToBytes(src)
	return unsafeheader.BytesToString(dst)
}

func (p *Obscure) DecodedLen(n int) int {
	if n <= 1 {
		return 0
	}
	return p.encodings[0].DecodedLen(n - 1)
}

func (p *Obscure) Decode(dst, src []byte) (n int, err error) {
	if len(src) == 0 {
		return 0, nil
	}
	idxchar := src[0]
	if idxchar >= 128 {
		return 0, ErrInvalidInput
	}
	idx := p.idxdec[idxchar]
	if idx == 0 && idxchar != p.idxChars[0] {
		return 0, ErrInvalidInput
	}
	enc := p.encodings[idx]
	return enc.Decode(dst, src[1:])
}

func (p *Obscure) DecodeBytes(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return nil, nil
	}
	dst := make([]byte, p.DecodedLen(len(src)))
	n, err := p.Decode(dst, src)
	if err != nil {
		return nil, err
	}
	return dst[:n], nil
}

func (p *Obscure) DecodeString(src string) ([]byte, error) {
	buf := unsafeheader.StringToBytes(src)
	return p.DecodeBytes(buf)
}
