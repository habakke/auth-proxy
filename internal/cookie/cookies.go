package cookie

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const CSRFCookieName = "csrf_state"

func MakeCookie(name string, value string) *http.Cookie {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   86400,
	}

	return c
}

func MakeInvalidationCookie(name string) *http.Cookie {
	c := &http.Cookie{
		Name:   name,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	return c
}

func MakeCSRFCookie(nonce string) *http.Cookie {
	return MakeCookie(CSRFCookieName, nonce)
}

// cookies are stored in a 3 part (value + timestamp + signature) to enforce that the values are as originally set.
// additionally, the 'value' is encrypted so it's opaque to the browser

// Validate ensures a cookie is properly signed
func Validate(cookie *http.Cookie, seed string, expiration time.Duration) (value string, t time.Time, ok bool) {
	// value, timestamp, sig
	parts := strings.Split(cookie.Value, "|")
	if len(parts) != 3 {
		return
	}
	sig := cookieSignature(seed, cookie.Name, parts[0], parts[1])
	if checkHmac(parts[2], sig) {
		ts, err := strconv.Atoi(parts[1])
		if err != nil {
			return
		}
		// The expiration timestamp set when the cookie was created
		// isn't sent back by the browser. Hence, we check whether the
		// creation timestamp stored in the cookie falls within the
		// window defined by (Now()-expiration, Now()].
		t = time.Unix(int64(ts), 0)
		if t.After(time.Now().Add(expiration*-1)) && t.Before(time.Now().Add(time.Minute*5)) {
			// it's a valid cookie. now get the contents
			rawValue, err := base64.URLEncoding.DecodeString(parts[0])
			if err == nil {
				value = string(rawValue)
				ok = true
				return
			}
		}
	}
	return
}

func Copy(req *http.Request, res http.ResponseWriter, cookieName string) error {
	c, err := req.Cookie(cookieName)
	if err != nil {
		return fmt.Errorf("request does not have a cookie named %s", cookieName)
	}

	http.SetCookie(res, c)
	return nil
}

// SignCookieValue returns a cookie that is signed and can later be checked with Validate
func SignCookieValue(seed string, cookieName string, value string, now time.Time) string {
	encodedValue := base64.URLEncoding.EncodeToString([]byte(value))
	timeStr := fmt.Sprintf("%d", now.Unix())
	sig := cookieSignature(seed, cookieName, encodedValue, timeStr)
	cookieVal := fmt.Sprintf("%s|%s|%s", encodedValue, timeStr, sig)
	return cookieVal
}

// EncryptCookieValue returns an encrypted cookie payload
func EncryptCookieValue(key string, value string) (encoded string, err error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", fmt.Errorf("invalid encryption key length %d", len(key))
	}
	c, err := NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	encoded, err = c.Encrypt(value)
	if err != nil {
		return "", err
	}
	return encoded, nil
}

// DecryptCookieValue returns a decrypted cookie payload
func DecryptCookieValue(key string, encryptedValue string) (value string, err error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", fmt.Errorf("invalid decryption key length %d", len(key))
	}
	c, err := NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	value, err = c.Decrypt(encryptedValue)
	if err != nil {
		return "", err
	}
	return value, nil
}

func cookieSignature(args ...string) string {
	h := hmac.New(sha256.New, []byte(args[0]))
	for _, arg := range args[1:] {
		if _, err := h.Write([]byte(arg)); err != nil {
			log.Fatal().AnErr("err", err).Msg("failed to write cookie sha256 signature")
		}
	}
	var b []byte
	b = h.Sum(b)
	return base64.URLEncoding.EncodeToString(b)
}

func checkHmac(input, expected string) bool {
	inputMAC, err1 := base64.URLEncoding.DecodeString(input)
	if err1 == nil {
		expectedMAC, err2 := base64.URLEncoding.DecodeString(expected)
		if err2 == nil {
			return hmac.Equal(inputMAC, expectedMAC)
		}
	}
	return false
}

// Cipher provides methods to encrypt and decrypt cookie values
type Cipher struct {
	cipher.Block
}

// NewCipher returns a new aes Cipher for encrypting cookie values
func NewCipher(secret []byte) (*Cipher, error) {
	c, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}
	return &Cipher{Block: c}, err
}

// Encrypt a value for use in a cookie
func (c *Cipher) Encrypt(value string) (string, error) {
	if len(value) > 64*1024*1024 {
		return "", errors.New("value is to large")
	}
	ciphertext := make([]byte, aes.BlockSize+len(value))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("failed to create initialization vector %s", err)
	}

	stream := cipher.NewCFBEncrypter(c.Block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(value))
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt a value from a cookie to it's original string
func (c *Cipher) Decrypt(s string) (string, error) {
	encrypted, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt cookie value %s", err)
	}

	if len(encrypted) < aes.BlockSize {
		return "", fmt.Errorf("encrypted cookie value should be "+
			"at least %d bytes, but is only %d bytes",
			aes.BlockSize, len(encrypted))
	}

	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(c.Block, iv)
	stream.XORKeyStream(encrypted, encrypted)

	return string(encrypted), nil
}
