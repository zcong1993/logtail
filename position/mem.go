package position

import "sync"

type Mem struct {
	name  string
	store map[string]int
	mu    sync.Mutex
}

func (m *Mem) Name() string {
	return "mem"
}

func (m *Mem) GetAll() map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return cloneStore(m.store)
}

func (m *Mem) Get(name string) (int, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, ok := m.store[name]
	return i, ok
}

func (m *Mem) Put(name string, p int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[name] = p
	return nil
}

func (m *Mem) Replace(new map[string]int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store = cloneStore(new)
	return nil
}

func (m *Mem) Close() error {
	m.store = nil
	return nil
}

func cloneStore(mp map[string]int) map[string]int {
	res := make(map[string]int, len(mp))
	for k, v := range mp {
		res[k] = v
	}
	return res
}
