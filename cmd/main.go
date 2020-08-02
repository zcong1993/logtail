package main

import (
	"os"

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
	var (
		positionFile = kingpin.Flag("position.file", "Position file name.").Default("position.json").String()
		interval     = kingpin.Flag("interval", "All ticker interval.").Default("10s").Duration()
		path         = kingpin.Flag("path", "Watch log path, support double star.").Required().String()
	)

	flag.AddFlags(kingpin.CommandLine, &logConfig)
	kingpin.CommandLine.GetFlag("help").Short('h')
	kingpin.Parse()

	logger := log.New(&logConfig)
	level.Info(logger).Log("msg", "Starting logtailer")

	po := position.NewJSONFile(&position.Config{
		Filename:     *positionFile,
		SaveInterval: *interval,
	}, logger)

	_, err := tailer.NewManager(&tailer.ManagerConfig{
		Path:         *path,
		Handler:      handler.StdHandler,
		Position:     po,
		SyncInterval: *interval,
	}, logger)

	if err != nil {
		level.Error(logger).Log("msg", "create manage error", "error", err)
		os.Exit(1)
	}

	c := make(chan struct{})
	<-c
}
