// +build ignore

package main

import (
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	mathrand "math/rand"
	"strings"
)

const (
	letters       = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	digits        = "0123456789"
	b64StdChars   = "+/"
	b64URLChars   = "-_"
	notFirstChars = b64StdChars + b64URLChars
)

func main() {
	typ := flag.String("type", "", "id / base64 / base62")
	urlsafe := flag.Bool("urlsafe", false, "use url-safe encoding for base64")
	flag.Parse()

	var result []string
	switch *typ {
	default:
		log.Fatal("missing type argument")
	case "id":
		result = GenerateIDCharTable()
	case "base64":
		result = GenerateBase64Table(*urlsafe)
	case "base62":
		result = GenerateBase62Table()
	}

	for i, row := range result {
		fmt.Printf("\"%s\", // %d\n", row, i)
	}
}

func GenerateIDCharTable() []string {
	var charTable = []byte(letters + digits)
	return makeRandomCharsTable(13, 59, charTable)
}

func GenerateBase62Table() []string {
	var charTable = []byte(letters + digits)
	return makeRandomCharsTable(13, 62, charTable)
}

func GenerateBase64Table(urlsafe bool) []string {
	var charTable = []byte(letters + digits + b64URLChars)
	if !urlsafe {
		charTable = append(charTable, b64StdChars...)
	}
	return makeRandomCharsTable(13, 64, charTable)
}

func makeRandomCharsTable(size, length int, charTable []byte) []string {
	if len(charTable) < length {
		panic("invalid charTable length not enough")
	}

	seed := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, seed); err != nil {
		panic(err)
	}
	rand_ := mathrand.New(mathrand.NewSource(int64(binary.LittleEndian.Uint64(seed))))

	result := make([]string, 0, size)
	for i := 0; i < size; i++ {
		for {
			rand_.Shuffle(len(charTable), func(i, j int) {
				charTable[i], charTable[j] = charTable[j], charTable[i]
			})
			if !strings.ContainsAny(string(charTable[:2]), notFirstChars) {
				break
			}
		}
		chars := string(charTable[:length])
		result = append(result, chars)
	}
	return result
}
