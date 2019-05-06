package main

import "math"

var (
	maxNonce = math.MaxInt64
)

const (
	targetBits   = 24
	dbFile       = "blockchain.db"
	blocksBucket = "blocks"
)
