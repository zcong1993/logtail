package handler

import (
	"fmt"
	"time"
)

var StdHandler = HandlerFunc(func(filename string, time time.Time, text string) error {
	fmt.Printf("filename=%s time=%s text=%s\n", filename, time, text)
	return nil
})
