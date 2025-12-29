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

	// Game loop with input handling
	for {
		// Try to advance the game
		gameOver, inputNeeded := game.ContinueDay()

		if inputNeeded != nil {
			// CLI mode: gather input and continue
			if !zcgame.CLIMode {
				log.Fatal("Player input needed but not in CLI mode: ", inputNeeded.Message)
			}

			input := zcgame.GatherCLIInput(game, inputNeeded)

			// Provide input and continue
			for {
				gameOver, inputNeeded = game.ContinueAfterInput(input)
				if inputNeeded == nil {
					break
				}
				input = zcgame.GatherCLIInput(game, inputNeeded)
			}
		}

		if gameOver {
			// Day completed successfully, show state and continue
			zcgame.RefreshRender(game)
			continue
		}

		// Game over - all players eliminated
		fmt.Println("GAME OVER - All players have been eliminated!")
		break
	}
}
