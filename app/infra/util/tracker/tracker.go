package tracker

import (
	"encoding/hex"
)

// RestoreToHexString 将从 URL 获取到的字符串还原为 Hex 字符串
func RestoreToHexString(str string) string {
	s := hex.EncodeToString([]byte(str))
	return s
}

// RestoreToByteString 将 Hex 字符串再还原为 20-byte 字符串
func RestoreToByteString(str string) string {
	byteSlice, err := hex.DecodeString(str)
	if err != nil {
		return str
	}

	return string(byteSlice)
}
