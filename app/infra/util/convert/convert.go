package convert

// ParseBoolToUint8 将布尔值转换为uint8
func ParseBoolToUint8(b bool) uint8 {
	if b {
		return 1
	}

	return 0
}

// ParseUint8ToBool 将uint8转换为布尔值
func ParseUint8ToBool(u uint8) bool {
	if u > 0 {
		return true
	}

	return false
}
