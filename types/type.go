package types

import (
	"sync"
)

type Network struct {
	Name     string
	Url      string
	Chainid  int
	Gaslimit int
	Coin     int64
	Token    string

	Address string
	Mykey   string
	Nonce   uint64
	Mux     sync.Mutex
}

type CacheConfig struct {
	Conn      string
	DBNum     string
	Passworkd string
}
