package crypto

import (
	"encoding/base32"
	"encoding/base64"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Option_Base64(t *testing.T) {
	for _, testkey := range testKeyList {
		ciphertext, err := CFBEncrypt(plaintext, testkey, Base64(base64.URLEncoding))
		assert.Nil(t, err)

		t.Log(string(ciphertext))
		base64Pattern := regexp.MustCompile(`^[A-Za-z0-9_\-]+=*$`)
		assert.Regexp(t, base64Pattern, string(ciphertext))

		decrypted, err := CFBDecrypt(ciphertext, testkey, Base64(base64.URLEncoding))
		assert.Nil(t, err)
		assert.Equal(t, plaintext, decrypted)
	}
}

func Test_Option_Base32(t *testing.T) {
	for _, testkey := range testKeyList {
		ciphertext, err := CBCEncrypt(plaintext, testkey,
			Base32(base32.HexEncoding.WithPadding(base32.NoPadding)))
		assert.Nil(t, err)

		t.Log(string(ciphertext))
		base32Pattern := regexp.MustCompile(`^[A-Za-z0-9]+$`)
		assert.Regexp(t, base32Pattern, string(ciphertext))

		decrypted, err := CBCDecrypt(ciphertext, testkey,
			Base32(base32.HexEncoding.WithPadding(base32.NoPadding)))
		assert.Nil(t, err)
		assert.Equal(t, plaintext, decrypted)
	}
}

func Test_Option_Base62(t *testing.T) {
	for _, testkey := range testKeyList {
		ciphertext, err := GCMEncrypt(plaintext, testkey, Base62(nil))
		assert.Nil(t, err)

		t.Log(string(ciphertext))
		base62Pattern := regexp.MustCompile(`^[A-Za-z0-9]+$`)
		assert.Regexp(t, base62Pattern, string(ciphertext))

		decrypted, err := GCMDecrypt(ciphertext, testkey, Base62(nil))
		assert.Nil(t, err)
		assert.Equal(t, plaintext, decrypted)
	}
}
