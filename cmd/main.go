package main

import (
	"fmt"

	"github.com/nxadm/tail"
)

func main() {
	t, err := tail.TailFile("./log.log", tail.Config{
		ReOpen: true,
		Poll:   true,
		Follow: true,
	})

	if err != nil {
		panic(err)
	}

	for l := range t.Lines {
		fmt.Println(l.Text)
	}
}
