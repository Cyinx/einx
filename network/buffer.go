package network

import (
	"sync"
)

const BUFFER_BASE_SIZE = 1024

var buffer_pool *sync.Pool = &sync.Pool{New: func() interface{} { return make([]byte, BUFFER_BASE_SIZE) }}
