package obscure

import (
	"encoding/base64"
	"github.com/jxskiss/base62"
	"sort"
	"sync"
)

const (
	seqlen  = 13
	idxlen  = 52
	letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

type Table [seqlen]string

var (
	// random sequence of 52 alphabet, used as sequence chars index
	idxcharsOnce sync.Once
	idxchars     = "vUbaKfVEexzjIrMORkDPAgtFTSZoJclGLHNBdqQXwmpsCnWiuyhY"
	idxdec       [128]int
)

func init() {
	for i, c := range idxchars {
		idxdec[c] = i
	}
}

func SetIdxChars(chars string) {
	bchars := []byte(chars)
	sort.Slice(bchars, func(i, j int) bool { return bchars[i] < bchars[j] })
	if len(chars) != 52 || string(bchars) != letters {
		panic("idx chars must be permutation of 52 letter characters")
	}
	idxcharsOnce.Do(func() {
		idxchars = chars
		idxdec = [128]int{}
		for i, c := range idxchars {
			idxdec[c] = i
		}
	})
}

var (
	// random sequence of digit number and alphabet, length 59
	idseqcharsOnce sync.Once
	idseqchars     = Table{
		"QGdaMz6yRUDO20loYLx8H4mwTcq7KIeE5rWBhPFsJCk9pvXSANfVZnjg13t", // 0
		"qzlsFXCJnGQ3yjbcWwiRudE5pO7ZUL0ktS4exvTra8H6DVgNAfMIKm9P1Bo", // 1
		"3j60lhsm2NUJwYnzZySgPVaFu8LWDQe9TqHKfotkC541pcdAXGOIxrivEMB", // 2
		"uLxbrt7OckT4IlaqhzvJfAsKD1WH5SoMnBXCUZwg3ydYRmVQNpPi80e6E92", // 3
		"9C2YW6rz8JdMjAZX4kUHRSPVvQDB1tEhulKGaxywO3Fmfc70iTsbNInL5go", // 4
		"plqFJ2csnTKzfIAZiVPaN3x7eMw64EU01u8kRyQbSd9YHhmXBgjGWOtvr5D", // 5
		"jCReQPULI9Fqb8AO2vacBGoEwM5shJ0f4DyKmH1g7Y6SitpkuXnWzNT3dZV", // 6
		"40c9ZDa2A73yrYdwBSf5sQ6VGH1bPtCKgmolxMLJORIznU8qjkXNWpvTeiF", // 7
		"CU7RFcLZP4m1yj9VsbNAEXvrpB3e2lJonKHODT5uYiS6tMQIahfzgGx80dq", // 8
		"Zw30VsyI6gpuYS1Rk5zbhFrTfdJtaWCGLxlAeOj4vHUiBoK2DcqnMQPN89m", // 9
		"2fHUwyE3eWKVXCjcRortnzQmJpu1hiZ78kbIB0MFgNDO5aLvlAPYd9qx4sS", // 10
		"hur7K9bx8DfXjtiz0ePGv1EnJdMNIcATOUZyBW62YS35VqwsLQmolkgFHaC", // 11
		"eyON5RbAi0cSBXWDzPwsmFtvLJqTK9pnCxMIEr12uQlkVhZY4fHd36a8UGg", // 12
	}
	idseqdec [seqlen][128]uint8
)

func init() {
	for i := range idseqdec {
		for j, c := range idseqchars[i] {
			idseqdec[i][c] = uint8(j)
		}
	}
}

func SetIDCharTable(table Table) {
	idseqcharsOnce.Do(func() {
		idseqchars = table
		idseqdec = [seqlen][128]uint8{}
		for i := range idseqdec {
			for j, c := range idseqchars[i] {
				idseqdec[i][c] = uint8(j)
			}
		}
	})
}

var (
	// random sequence of digit number, alphabet, and "_-./", length 64
	b64seqcharsOnce sync.Once
	b64seqchars     = Table{
		"ND_2u8SvyakLUlpn-TBJz/Pwf3qcobC6FsZMGeRx10riW9QVtO+IYHEhgmd5j4KA", // 0
		"NaEKXc_wdC7V9xbvWitO+LHqpfSIQ0FM3lr5yRAzJ8UuBmZ/4YToDnG6she1-Pgk", // 1
		"H98j5U0Q_+wouWinfICcLVOMSblET2z7hg/X-6BDqrkY3d4N1JxPFApavmyZteRK", // 2
		"g0mPYA1vUplVas46wMCHWXG8uLS3TBfehqcEk+5Z7QitxOIn2/RJjbKFy-dr_z9N", // 3
		"D247tKfMNPR/3SmedGJvAYuI6ryqLVXOBHQ0W-s_i5UTEb98op+CZkFnzgj1xchl", // 4
		"FYc7Ad1ukhb4nZ69SOoN2jWmD+RLw3MCVPyf-gi5rU/veQGxBTHXtEIp_zq0Ks8J", // 5
		"163wju5Gax_X2y-fgvLAFH+bCRqI8VlBPhNWEKpeU/romsMQYi90DOT4dJz7Ztkn", // 6
		"9b6PvB0MaDh1Rodl/V-c7q5EQIN8LiOsKjHkyZ23Tp_eugWCxAY+m4znwtGSXJrf", // 7
		"VnEztw1GBi5XFSxKbRM3uU4Nfqlm-Z0TcoPCvyLHaJ2OWQs/gpre7Ad_h9Y8kID+", // 8
		"4DzJfl3d_tUIrsFLx58iOCg0bEwqV+-KmoAXSNjHpePnBvRT/auM2y79WZkYchG6", // 9
		"612JmgEc0KRPxZLy5UpGI+qfh/Q4_-kwndiXvSr9tzlMNauWAFT3H8OobsB7CeVj", // 10
		"hJ+oLE9U5bfOHxGu/I12NCeZ3awiTryndQSWY4_kXjFsl87MPmgcKzqB-VpRA6t0", // 11
		"zfpIDYkdwbq+n/iysUt0u9OGKlcr1MZAxNm2eQ8W_hXS6VFRPL5HvTC-B73aEJj4", // 12
	}
	b64encodings [seqlen]*base64.Encoding
)

func init() {
	for i := range b64encodings {
		b64encodings[i] = base64.NewEncoding(b64seqchars[i]).WithPadding(base64.NoPadding)
	}
}

func SetBase64Table(table Table) {
	b64seqcharsOnce.Do(func() {
		b64seqchars = table
		for i := range b64encodings {
			b64encodings[i] = base64.NewEncoding(b64seqchars[i]).WithPadding(base64.NoPadding)
		}
	})
}

var (
	// random sequence of digit number, alphabet, and "_-./", length 62
	b62seqcharsOnce sync.Once
	b62seqchars     = Table{
		"3EM9CuAqyesI5LNUxV4dGBpgzjYJvOKalbwnrhQXTS2Zf0oiP6mD8c1HW7kFtR", // 0
		"jSDRJWplctIi9rezKwM06XkF2Nu8V4QTsfBavL7HhCAdgn5Ob1yxGZoEPqUmY3", // 1
		"2CHsW74ONAbZ8ixzMlhLyv1oaGeXtDVYgJnPkuF95cm03jBISrQU6fdpERTqwK", // 2
		"1MhNDCycOujzHokPKqE8eAYJ6bTrtdUVIGwsZa0FL4B2gxn95XpfRvmWiQ7l3S", // 3
		"cvIqPYGn1DWRC8HrOz26NajUtxLeTuogm35kyZsp74EwKA0dhlfJXVibF9BSMQ", // 4
		"cUNEIZ6CVpPkmLovrl3J4hADdnYiuS2KjqTB8zQbGF7yMeRfgswX9H5OWa10xt", // 5
		"Cj81QAgOeDhd62JzmPsXHnaBvxI5UpWTVFutlcNRK0Mb9Sof3rqywk4Y7LGiEZ", // 6
		"CylQHbRVzNgvUIKon8FThLEqc0f4taSs9eWiM2xdGrjZP3wOBDJp7uA16XYmk5", // 7
		"8lBWPsoTypqgO3iMZ5xdkez07UQGrhvFa1cwRXEt42JIjmfV9NCKALYbHDunS6", // 8
		"xeo63E1SjVNKr57bT0OBkZlhMs8dJt2YcmwQPyDnIivuXzRL9CHWq4AaGpfFUg", // 9
		"NUSTmP0yv8XZxeILfJhWH5rQ4ElF3AMcOGBdqaj9usiVboDnz2RwK1Yp67kCtg", // 10
		"jOb2cCT58ryS7A9pYfGD0gPaJWtelmNLUk3wvnsxoMIuFV1zKiQHEhRB46XdZq", // 11
		"5dwzkgCaFxVp2fyAlNbMR1PvjmX4q9OIUYcT0QHZKWEDnhL8tu3ioe7B6rJSsG", // 12
	}
	b62encodings [seqlen]*base62.Encoding
)

func init() {
	for i := range b62encodings {
		b62encodings[i] = base62.NewEncoding(b62seqchars[i])
	}
}

func SetBase62Table(table Table) {
	b62seqcharsOnce.Do(func() {
		b62seqchars = table
		for i := range b62encodings {
			b62encodings[i] = base62.NewEncoding(b62seqchars[i])
		}
	})
}
