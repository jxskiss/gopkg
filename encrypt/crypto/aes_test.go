package crypto

import (
	"bytes"
	"regexp"
	"testing"
)

var (
	testkey   = []byte("_test_test_test_")
	plainText = []byte("hello 世界")
)

func Test_GCM(t *testing.T) {
	cipherText, err := GCMEncrypt(plainText, testkey)
	if err != nil {
		t.Errorf("GCM failed encrypt: %v", err)
	}
	decrypted, err := GCMDecrypt(cipherText, testkey)
	if err != nil {
		t.Errorf("GCM failed decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plainText) {
		t.Errorf("GCM got invalid decrypted result")
	}
}

func Test_GCM_NewKey(t *testing.T) {
	cipherText, key, additional, err := GCMEncryptNewKey(plainText)
	if err != nil {
		t.Errorf("GCM_NewKey failed encrypt: %v", err)
	}
	decrypted, err := GCMDecrypt(cipherText, key, AdditionalData(additional))
	if err != nil {
		t.Errorf("GCM_NewKey failed decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plainText) {
		t.Errorf("GCM_NewKey got invalid decrypted result")
	}
}

func Test_CBC(t *testing.T) {
	cipherText, err := CBCEncrypt(plainText, testkey)
	if err != nil {
		t.Errorf("CBC failed encrypt: %v", err)
	}
	decrypted, err := CBCDecrypt(cipherText, testkey)
	if err != nil {
		t.Errorf("CBC failed decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plainText) {
		t.Errorf("CBC got invalid decrypted result")
	}
}

func Test_CFB(t *testing.T) {
	cipherText, err := CFBEncrypt(plainText, testkey)
	if err != nil {
		t.Errorf("CFB failed encrypt: %v", err)
	}
	decrypted, err := CFBDecrypt(cipherText, testkey)
	if err != nil {
		t.Errorf("CFB failed decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plainText) {
		t.Errorf("CFB got invalid decrypted result")
	}
}

func Test_Option_Base64(t *testing.T) {
	cipherText, err := CFBEncrypt(plainText, testkey, Base64)
	if err != nil {
		t.Errorf("Option_Base64 failed encrypt: %v", err)
	}
	t.Log(string(cipherText))
	base64Pattern := regexp.MustCompile(`^[A-Za-z0-9_\-]+=*$`)
	if !base64Pattern.Match(cipherText) {
		t.Errorf("Option_Base64 got invalid character")
	}
}

func Test_KeyPadding(t *testing.T) {
	for i := 0; i < 17; i++ {
		key := make([]byte, i)
		key = KeyPadding(key)
		if len(key) != 16 {
			t.Errorf("KeyPadding got invalid key length")
		}
	}
	for i := 17; i < 25; i++ {
		key := make([]byte, i)
		key = KeyPadding(key)
		if len(key) != 24 {
			t.Errorf("KeyPadding got invalid key length")
		}
	}
	for i := 25; i < 50; i++ {
		key := make([]byte, i)
		key = KeyPadding(key)
		if len(key) != 32 {
			t.Errorf("KeyPadding got invalid key length")
		}
	}
}
