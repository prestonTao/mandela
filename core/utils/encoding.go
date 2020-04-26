package utils

import (
	"encoding/base64"
)

func EnBase64(bs []byte) string {
	return base64.StdEncoding.EncodeToString(bs)
}
func DenBase64(str string) (bs []byte, err error) {
	bs, err = base64.StdEncoding.DecodeString(str)
	return
}
