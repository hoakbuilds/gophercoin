package blockchain

import "strconv"

// IntToHex converts an int into a hexadecimal.
func IntToHex(n int64) []byte {
	return []byte(strconv.FormatInt(n, 16))
}
