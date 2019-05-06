package main

import "strconv"

// IntToHex converst an int into a hexadecimal.
// This will be used in the function prepareData
// coded below
func IntToHex(n int64) []byte {
	return []byte(strconv.FormatInt(n, 16))
}
