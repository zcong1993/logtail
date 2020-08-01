package position

type Position interface {
	Name() string
	GetAll() map[string]int64
	Get(name string) int64
	Put(name string, p int64) error
	Replace(new map[string]int64) error
	Remove(name string) error
	Stop()
}
