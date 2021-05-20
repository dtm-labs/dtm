package main

import (
	"time"

	"github.com/yedf/dtm/dtmsvr"
	"github.com/yedf/dtm/examples"
)

type M = map[string]interface{}

func main() {
	dtmsvr.LoadConfig()
	go dtmsvr.StartSvr()
	go examples.StartSvr()
	time.Sleep(1000 * 1000 * 1000 * 1000)
}
