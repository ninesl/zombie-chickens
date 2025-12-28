package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ninesl/zombie-chickens/zcgame"
)

func main() {
	debug := flag.Bool("debug", false, "put events in top of night deck for testing")
	flag.Parse()

	names := flag.Args()
	if len(names) < 1 {
		log.Fatal("usage: go run . [-debug] name1 [name2 ...]")
	}

	game, err := zcgame.CreateNewGame(names...)
	if err != nil {
		log.Fatal(err)
	}

	if *debug {
		game.DebugEventsOnTop()
	}

	for range 7 {
		zcgame.RefreshRender(game)
		if !game.DoDay() {
			zcgame.ClearScreen()
			fmt.Printf("%s\n", game.StatsString())
			fmt.Println("GAME OVER - All players have been eliminated!")
			break
		}
	}
}
