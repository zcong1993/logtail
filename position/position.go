package position

type Position interface {
	Name() string
	GetAll() map[string]int
	Get(name string) (int, bool)
	Put(name string, p int) error
	Replace(new map[string]int) error
	Stop()
}
