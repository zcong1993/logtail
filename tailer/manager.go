package tailer

import (
	"os"
	"path/filepath"
	"time"

	"github.com/bmatcuk/doublestar"
	"github.com/fsnotify/fsnotify"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/zcong1993/logtail/handler"
	"github.com/zcong1993/logtail/position"
)

type ManagerConfig struct {
	Path         string
	Handler      handler.Handler
	Position     position.Position
	SyncInterval time.Duration
}

type Manager struct {
	cfg    *ManagerConfig
	logger log.Logger

	watcher *fsnotify.Watcher
	watches map[string]struct{}
	tails   map[string]*Tailer

	quit chan struct{}
	done chan struct{}
}

func NewManager(cfg *ManagerConfig, logger log.Logger) (*Manager, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	m := &Manager{
		cfg:     cfg,
		logger:  logger,
		watcher: watcher,
		watches: make(map[string]struct{}),
		tails:   make(map[string]*Tailer),
		quit:    make(chan struct{}),
		done:    make(chan struct{}),
	}

	err = m.sync()
	if err != nil {
		level.Error(m.logger).Log("msg", "error running sync function", "error", err)
	}
	go m.run()
	return m, nil
}

func (m *Manager) Stop() {
	close(m.quit)
	<-m.done
}

func (m *Manager) run() {
	t := time.NewTicker(m.cfg.SyncInterval)
	defer func() {
		err := m.watcher.Close()
		if err != nil {
			level.Error(m.logger).Log("msg", "closing watcher", "error", err)
		}
		tailersMp := make(map[string]struct{}, len(m.tails))
		for k := range m.tails {
			tailersMp[k] = struct{}{}
		}
		m.stopTailing(tailersMp)
	}()

	for {
		select {
		case <-m.quit:
			return
		case <-t.C:
			err := m.sync()
			if err != nil {
				level.Error(m.logger).Log("msg", "error running sync function", "error", err)
			}
		case event := <-m.watcher.Events:
			switch event.Op {
			case fsnotify.Create:
				matched, err := doublestar.Match(m.cfg.Path, event.Name)
				if err != nil {
					level.Error(m.logger).Log("msg", "failed to match file", "error", err, "filename", event.Name)
					continue
				}
				if !matched {
					level.Debug(m.logger).Log("msg", "new file does not match glob", "filename", event.Name)
					continue
				}
				m.startTail([]string{event.Name})
			default:
				// No-op we only care about Create events
			}
		case err := <-m.watcher.Errors:
			level.Error(m.logger).Log("msg", "error from fswatch", "error", err)
		}
	}
}

func (m *Manager) sync() error {
	matches, err := doublestar.Glob(m.cfg.Path)
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		level.Debug(m.logger).Log("msg", "no files matched requested path, nothing will be tailed", "path", m.cfg.Path)
	}

	for i := 0; i < len(matches); i++ {
		if !filepath.IsAbs(matches[i]) {
			path, err := filepath.Abs(matches[i])
			if err != nil {
				return err
			}
			matches[i] = path
		}
	}

	// Get the current unique set of dirs to watch.
	dirs := map[string]struct{}{}
	for _, p := range matches {
		dirs[filepath.Dir(p)] = struct{}{}
	}

	// Add any directories which are not already being watched.
	toStartWatching := missing(m.watches, dirs)
	m.startWatching(toStartWatching)

	// Remove any directories which no longer need watching.
	toStopWatching := missing(dirs, m.watches)
	m.stopWatching(toStopWatching)

	m.watches = dirs

	// start tail
	m.startTail(matches)
	// stop tail
	matchesMp := make(map[string]struct{}, len(matches))
	for _, m := range matches {
		matchesMp[m] = struct{}{}
	}

	tailersMp := make(map[string]struct{}, len(m.tails))
	for k := range m.tails {
		tailersMp[k] = struct{}{}
	}

	stops := missing(matchesMp, tailersMp)
	m.stopTailing(stops)
	return nil
}

func (m *Manager) startTail(paths []string) {
	for _, p := range paths {
		if _, ok := m.tails[p]; ok {
			continue
		}
		fi, err := os.Stat(p)
		if err != nil {
			level.Error(m.logger).Log("msg", "failed to tail file, stat failed", "error", err, "filename", p)
			continue
		}
		if fi.IsDir() {
			level.Error(m.logger).Log("msg", "failed to tail file", "error", "file is a directory", "filename", p)
			continue
		}
		level.Debug(m.logger).Log("msg", "tailing new file", "filename", p)
		tailer, err := NewTailer(&Config{
			filename:             p,
			handler:              m.cfg.Handler,
			position:             m.cfg.Position,
			positionSyncInterval: m.cfg.SyncInterval,
		}, m.logger)
		if err != nil {
			level.Error(m.logger).Log("msg", "failed to start tailer", "error", err, "filename", p)
			continue
		}
		m.tails[p] = tailer
	}
}

func (m *Manager) stopTailing(ps map[string]struct{}) {
	for p := range ps {
		if tailer, ok := m.tails[p]; ok {
			err := tailer.Stop()
			if err != nil {
				level.Error(m.logger).Log("msg", "stop tailer", "error", err)
			}
			delete(m.tails, p)
		}
	}
}

func (m *Manager) startWatching(dirs map[string]struct{}) {
	for dir := range dirs {
		if _, ok := m.watches[dir]; ok {
			continue
		}
		level.Debug(m.logger).Log("msg", "watching new directory", "directory", dir)
		if err := m.watcher.Add(dir); err != nil {
			level.Error(m.logger).Log("msg", "error adding directory to watcher", "error", err)
		}
	}
}

func (m *Manager) stopWatching(dirs map[string]struct{}) {
	for dir := range dirs {
		if _, ok := m.watches[dir]; !ok {
			continue
		}
		level.Debug(m.logger).Log("msg", "removing directory from watcher", "directory", dir)
		if err := m.watcher.Remove(dir); err != nil {
			level.Error(m.logger).Log("msg", "failed to remove directory from watcher", "error", err)
		}
	}
}

func missing(as map[string]struct{}, bs map[string]struct{}) map[string]struct{} {
	c := map[string]struct{}{}
	for a := range bs {
		if _, ok := as[a]; !ok {
			c[a] = struct{}{}
		}
	}
	return c
}
