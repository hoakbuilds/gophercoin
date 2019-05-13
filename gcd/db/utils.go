package db

import "strconv"

// IntToHex converst an int into a hexadecimal.
// This will be used in the function prepareData
// coded below
func IntToHex(n int64) []byte {
	return []byte(strconv.FormatInt(n, 16))
}

// ReverseBytes reverses a byte array
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
