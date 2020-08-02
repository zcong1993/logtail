package position

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

const positionFileMode = 0600

type Config struct {
	Filename     string
	SaveInterval time.Duration
}

type JSONFile struct {
	cfg      *Config
	position map[string]int64
	mu       sync.Mutex
	logger   log.Logger
	quit     chan struct{}
	done     chan struct{}
}

func NewJSONFile(cfg *Config, logger log.Logger) *JSONFile {
	f := &JSONFile{
		cfg:      cfg,
		position: make(map[string]int64, 0),
		logger:   log.With(logger, "component", "position", "type", "jsonFile"),
		quit:     make(chan struct{}),
		done:     make(chan struct{}),
	}

	go f.run()

	return f
}

func (f *JSONFile) Name() string {
	return "jsonFile"
}

func (f *JSONFile) GetAll() map[string]int64 {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.position
}

func (f *JSONFile) Get(name string) int64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.position[name]
	if !ok {
		return 0
	}

	return p
}

func (f *JSONFile) Put(name string, p int64) error {
	level.Debug(f.logger).Log("name", name, "p", p)
	f.mu.Lock()
	defer f.mu.Unlock()
	f.position[name] = p

	return nil
}

func (f *JSONFile) Replace(new map[string]int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.position = new

	return nil
}

func (f *JSONFile) Remove(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.position, name)

	return nil
}

func (f *JSONFile) run() {
	t := time.NewTicker(f.cfg.SaveInterval)
	defer func() {
		err := f.save()
		if err != nil {
			level.Error(f.logger).Log("msg", "save position error", "error", err)
		}
		f.done <- struct{}{}
	}()

	for {
		select {
		case <-f.quit:
			return
		case <-t.C:
			err := f.save()
			if err != nil {
				level.Error(f.logger).Log("msg", "save position error", "error", err)
			}
			f.cleanup()
		}
	}
}

func (f *JSONFile) save() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	content, err := json.Marshal(f.position)

	if err != nil {
		return err
	}
	// create new
	fname := f.cfg.Filename + "-new"
	err = ioutil.WriteFile(fname, content, os.FileMode(positionFileMode))
	if err != nil {
		return err
	}
	// rename
	return os.Rename(fname, f.cfg.Filename)
}

func (f *JSONFile) cleanup() {
	f.mu.Lock()
	defer f.mu.Unlock()
	for filename := range f.position {
		if _, err := os.Stat(filename); err != nil {
			if os.IsNotExist(err) {
				// File no longer exists.
				delete(f.position, filename)
			} else {
				// Can't determine if file exists or not, some other error.
				level.Warn(f.logger).Log("msg", "could not determine if log file "+
					"still exists while cleaning positions file", "error", err)
			}
		}
	}
}

func (f *JSONFile) Stop() {
	close(f.quit)
	<-f.done
}
