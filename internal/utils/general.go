package utils

import "encoding/base64"

func BytesToBase64String(data []byte) string {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(dst, data)
	return string(dst)
}
