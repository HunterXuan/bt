package convert

import "strconv"

// ParseStringToUint64 convert string to uint64
func ParseStringToUint64(s string) uint64 {
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}

	return val
}
