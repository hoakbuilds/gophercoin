package blockchain

import "math"

// Unexported constants
const (
	blocksBucket        = "blockchain"
	targetBits          = 1
	utxoBucket          = "utxo"
	bucketExtension     = ".db"
	genesisCoinbaseData = "May 7 2019, 10:00pm, The Times	Jürgen Klopp makes Liverpool believe they can do the impossible		Matt Dickinson, Chief Sports Writer"
)

var (
	maxNonce = math.MaxInt64
)
