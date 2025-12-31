package cligame

import (
	"flag"
	"fmt"
	"log"

	"github.com/ninesl/zombie-chickens/zcgame"
)

func RunGame() {
	names := flag.Args()
	if len(names) < 1 {
		log.Fatal("usage: go run . [-debug] name1 [name2 ...]")
	}

	game, err := zcgame.CreateNewGame(names...)
	if err != nil {
		log.Fatal(err)
	}

	// Game loop
	for {
		// Try to advance the game
		gameContinues, inputNeeded := game.ContinueDay()

		if inputNeeded != nil {
			// Gather input and continue
			input := GatherInput(game, inputNeeded)

			// Provide input and continue
			for {
				gameContinues, inputNeeded = game.ContinueAfterInput(input)
				if inputNeeded == nil {
					break
				}
				input = GatherInput(game, inputNeeded)
			}
		}
		RefreshRender(game)

		if gameContinues {
			// Day completed successfully, show state and continue
			continue
		}

		// Game over - all players eliminated
		fmt.Println("GAME OVER - All players have been eliminated!")
		break
	}
}
