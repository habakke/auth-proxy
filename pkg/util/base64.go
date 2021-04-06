package util

import "encoding/base64"

func Base64Encode(message []byte) string {
	b := make([]byte, base64.StdEncoding.EncodedLen(len(message)))
	base64.StdEncoding.Encode(b, message)
	return string(b)
}

func Base64Decode(message string) (b []byte, err error) {
	var l int
	b = make([]byte, base64.StdEncoding.DecodedLen(len(message)))
	l, err = base64.StdEncoding.Decode(b, []byte(message))
	if err != nil {
		return
	}
	return b[:l], nil
}
