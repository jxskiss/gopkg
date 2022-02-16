package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testKeyList = [][]byte{
		[]byte("_test_test_test_"),                 // 16 bytes
		[]byte("_test_test_test_12345678"),         // 24 bytes
		[]byte("_test_test_test_1234567890abcdef"), // 32 bytes
	}
	plaintext = []byte("hello 世界")
)

func Test_GCM(t *testing.T) {
	for _, testkey := range testKeyList {
		ciphertext, err := GCMEncrypt(plaintext, testkey)
		assert.Nil(t, err)

		decrypted, err := GCMDecrypt(ciphertext, testkey)
		assert.Nil(t, err)
		assert.Equal(t, plaintext, decrypted)
	}
}

func Test_GCM_EmptyKey(t *testing.T) {
	emptyKey := []byte("")
	ciphertext, err := GCMEncrypt(plaintext, emptyKey)
	assert.Nil(t, err)

	decrypted, err := GCMDecrypt(ciphertext, emptyKey)
	assert.Nil(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func Test_GCM_NewKey(t *testing.T) {
	ciphertext, key, additional, err := GCMEncryptNewKey(plaintext)
	assert.Nil(t, err)

	decrypted, err := GCMDecrypt(ciphertext, key, AdditionalData(additional))
	assert.Nil(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func Test_CBC(t *testing.T) {
	for _, testkey := range testKeyList {
		ciphertext, err := CBCEncrypt(plaintext, testkey)
		assert.Nil(t, err)

		decrypted, err := CBCDecrypt(ciphertext, testkey)
		assert.Nil(t, err)
		assert.Equal(t, plaintext, decrypted)
	}
}

func Test_CFB(t *testing.T) {
	for _, testkey := range testKeyList {
		ciphertext, err := CFBEncrypt(plaintext, testkey)
		assert.Nil(t, err)

		decrypted, err := CFBDecrypt(ciphertext, testkey)
		assert.Nil(t, err)
		assert.Equal(t, plaintext, decrypted)
	}
}

func Test_KeyPadding(t *testing.T) {
	for i := 0; i < 17; i++ {
		key := make([]byte, i)
		key = KeyPadding(key)
		assert.Len(t, key, 16)
	}
	for i := 17; i < 25; i++ {
		key := make([]byte, i)
		key = KeyPadding(key)
		assert.Len(t, key, 24)
	}
	for i := 25; i < 50; i++ {
		key := make([]byte, i)
		key = KeyPadding(key)
		assert.Len(t, key, 32)
	}
}
