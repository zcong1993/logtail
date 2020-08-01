package tailer

import (
	"os"
	"time"

	"github.com/go-kit/kit/log/level"

	"github.com/go-kit/kit/log"
	"github.com/nxadm/tail"
	"github.com/zcong1993/logtail/handler"
	"github.com/zcong1993/logtail/position"
)

type Config struct {
	filename             string
	handler              handler.Handler
	position             position.Position
	positionSyncInterval time.Duration
}

type Tailer struct {
	cfg    *Config
	tail   *tail.Tail
	logger log.Logger

	quit chan struct{}
	done chan struct{}
}

func NewTailer(cfg *Config, logger log.Logger) (*Tailer, error) {
	filename := cfg.filename
	position := cfg.position
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	pos := position.Get(filename)
	if fi.Size() < pos {
		position.Remove(filename)
	}

	tail, err := tail.TailFile(filename, tail.Config{
		Follow: true,
		Poll:   true,
		ReOpen: true,
		Location: &tail.SeekInfo{
			Offset: pos,
			Whence: 0,
		},
	})
	if err != nil {
		return nil, err
	}

	logger = log.With(logger, "component", "tailer")
	t := &Tailer{
		cfg:    cfg,
		logger: logger,
		tail:   tail,
		quit:   make(chan struct{}),
		done:   make(chan struct{}),
	}

	go t.run()
	return t, nil
}

func (t *Tailer) run() {
	ticker := time.NewTicker(t.cfg.positionSyncInterval)
	defer func() {
		t.savePosition()
		close(t.done)
	}()

	for {
		select {
		case <-t.quit:
			return
		case <-ticker.C:
			t.savePosition()
		case l, ok := <-t.tail.Lines:
			if !ok {
				return
			}
			if l.Err != nil {
				level.Error(t.logger).Log("msg", "error reading line", "path", t.cfg.filename, "error", l.Err)
			}
			if err := t.cfg.handler(t.cfg.filename, l.Time, l.Text); err != nil {
				level.Error(t.logger).Log("msg", "error handling line", "path", t.cfg.filename, "error", err)
			}
		}
	}
}

func (t *Tailer) savePosition() error {
	pos, err := t.tail.Tell()
	if err != nil {
		return err
	}
	return t.cfg.position.Put(t.cfg.filename, pos)
}

func (t *Tailer) Stop() {
	close(t.quit)
	<-t.done
}
