package main

import (
	"time"

	"github.com/zcong1993/logtail/handler"
	"github.com/zcong1993/logtail/tailer"

	"github.com/go-kit/kit/log/level"
	"github.com/zcong1993/logtail/position"
	"github.com/zcong1993/x/log"
	"github.com/zcong1993/x/log/flag"
	"gopkg.in/alecthomas/kingpin.v2"
)

var logConfig = log.Config{}

func main() {
	flag.AddFlags(kingpin.CommandLine, &logConfig)
	kingpin.CommandLine.GetFlag("help").Short('h')
	kingpin.Parse()

	logger := log.New(&logConfig)
	level.Info(logger).Log("msg", "Starting logtailer")

	interval := time.Second * 5
	po := position.NewJSONFile(&position.Config{
		Filename:       "position.json",
		Saveint64erval: interval,
	}, logger)

	_, err := tailer.NewManager(&tailer.ManagerConfig{
		Path:         "./logs/*.log",
		Handler:      handler.StdHandler,
		Position:     po,
		SyncInterval: interval,
	}, logger)

	if err != nil {
		panic(err)
	}

	c := make(chan struct{})
	<-c
}
