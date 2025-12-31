package main

import (
	"flag"

	"github.com/ninesl/zombie-chickens/cligame"
	"github.com/ninesl/zombie-chickens/webapp"
	"github.com/ninesl/zombie-chickens/zcgame"
)

func main() {
	web := flag.Bool("web", false, "run web server instead of CLI game")
	debug := flag.Bool("debug", false, "put events on top of night deck for testing")
	flag.Parse()

	zcgame.DebugMode = *debug

	if *web {
		webapp.RunServer()
	} else {
		cligame.RunGame()
	}
}
