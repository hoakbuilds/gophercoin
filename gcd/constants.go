package gcd

import "math"

var (
	maxNonce = math.MaxInt64
)

const (
	targetBits          = 1
	blocksBucket        = "blockchain"
	utxoBucket          = "utxo"
	bucketExtension     = ".db"
	genesisCoinbaseData = "May 7 2019, 10:00pm, The Times	Jürgen Klopp makes Liverpool believe they can do the impossible		Matt Dickinson, Chief Sports Writer"
)
