package position

import "sync"

type Mem struct {
	name  string
	store map[string]int64
	mu    sync.Mutex
}

func (m *Mem) Name() string {
	return "mem"
}

func (m *Mem) GetAll() map[string]int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return cloneStore(m.store)
}

func (m *Mem) Get(name string) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, ok := m.store[name]
	if !ok {
		return 0
	}
	return i
}

func (m *Mem) Put(name string, p int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[name] = p
	return nil
}

func (m *Mem) Replace(new map[string]int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store = cloneStore(new)
	return nil
}

func (m *Mem) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.store, name)
	return nil
}

func (m *Mem) Stop() {
	return
}

func cloneStore(mp map[string]int64) map[string]int64 {
	res := make(map[string]int64, len(mp))
	for k, v := range mp {
		res[k] = v
	}
	return res
}
