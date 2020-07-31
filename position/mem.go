package position

import "sync"

type Mem struct {
	name  string
	store map[string]int
	mu    sync.Mutex
}
