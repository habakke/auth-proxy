package cookie

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestEncryptCookieValue(t *testing.T) {
	const key = "2345asdYDS!2012L"
	const payload = "This is a cookie payload"

	e1, err := EncryptCookieValue(key, payload)
	if err != nil {
		t.Errorf("failed to encrypt payload")
	}

	e2, err := EncryptCookieValue(key, payload)
	if err != nil {
		t.Errorf("failed to encrypt payload")
	}

	assert.NotEqual(t, e1, e2)
}

func TestEncryptAndDecryptCookieValue(t *testing.T) {
	const key = "2345asdYDS!2012L"
	const payload = "This is a cookie payload"

	e, err := EncryptCookieValue(key, payload)
	if err != nil {
		t.Errorf("failed to encrypt payload")
	}

	assert.NotEqual(t, payload, e)

	d, err := DecryptCookieValue(key, e)
	if err != nil {
		t.Errorf("failed to decrypt payload")
	}

	assert.Equal(t, payload, d)
}

func TestSignAndValidateCookieValue(t *testing.T) {
	const seed = "0123456789abcdefghijklmnopqrstuv"
	const payload = "This is a cookie payload"
	const cookieName = "test"
	value := SignCookieValue(seed, cookieName, payload, time.Now())
	assert.NotEmpty(t, value, "signed value is empty")

	c := &http.Cookie{Value: value, Name: cookieName}
	cookiePayload, _, ok := Validate(c, seed, time.Hour*24)
	assert.True(t, ok, "validation has failed")
	assert.Equal(t, payload, cookiePayload, "payload does not match expected value")
}

func TestEncodeAndDecodeAccessToken(t *testing.T) {
	const secret = "0123456789abcdefghijklmnopqrstuv"
	const token = "my access token"
	c, err := NewCipher([]byte(secret))
	assert.Equal(t, nil, err)

	encoded, err := c.Encrypt(token)
	assert.Equal(t, nil, err)

	decoded, err := c.Decrypt(encoded)
	assert.Equal(t, nil, err)

	assert.NotEqual(t, token, encoded)
	assert.Equal(t, token, decoded)
}

func TestEncodeAndDecodeAccessTokenB64(t *testing.T) {
	const secretBase64 = "A3Xbr6fu6Al0HkgrP1ztjb-mYiwmxgNPP-XbNsz1WBk="
	const token = "my access token"

	secret, _ := base64.URLEncoding.DecodeString(secretBase64)
	c, err := NewCipher([]byte(secret))
	assert.Equal(t, nil, err)

	encoded, err := c.Encrypt(token)
	assert.Equal(t, nil, err)

	decoded, err := c.Decrypt(encoded)
	assert.Equal(t, nil, err)

	assert.NotEqual(t, token, encoded)
	assert.Equal(t, token, decoded)
}
