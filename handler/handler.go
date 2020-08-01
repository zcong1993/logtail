package handler

import "time"

type Handler = func(filename string, time time.Time, text string) error
