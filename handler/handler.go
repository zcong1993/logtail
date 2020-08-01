package handler

import "time"

type Handler interface {
	Handle(filename string, time time.Time, text string) error
}

type HandlerFunc func(filename string, time time.Time, text string) error

func (f HandlerFunc) Handle(filename string, time time.Time, text string) error {
	return f(filename, time, text)
}
