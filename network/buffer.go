package network

import (
	"sync"
)

const BUFFER_BASE_SIZE = 128

var buffer_pool *sync.Pool = &sync.Pool{New: func() interface{} { return make([]byte, 0, BUFFER_BASE_SIZE) }}
