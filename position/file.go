package position

import (
	"os"
	"sync"
)

type JSONFile struct {
	file  *os.File
	cache map[string]int
	mu    sync.Mutex
}

func (f *JSONFile) Name() string {
	return "jsonFile"
}

//func (f *JSONFile) GetAll() map[string]int {
//
//}
//
//func (f *JSONFile) Get(name string) (int, bool) {}
//
//func (f *JSONFile) Put(name string, p int) error {}
//
//func (f *JSONFile) Replace(new map[string]int) error {}
//
//func (f *JSONFile) io.Closer {}
