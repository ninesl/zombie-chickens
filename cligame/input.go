package cligame

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/ninesl/zombie-chickens/zcgame"
)

// intSliceChoices formats a slice of integers for display as valid choices.
func intSliceChoices(s ...int) string {
	return fmt.Sprintf("%+v", s)
}

// GatherInput displays the game state and prompt, then reads player input from stdin.
// It validates input against ValidChoices and re-prompts on invalid input.
//
// The render type from inputNeeded determines how the game state is displayed:
//   - RenderNormal: Standard view with hand indices
//   - RenderForDiscard: Farm item indices shown for event discards
//   - RenderForNight: Stack indices shown for defense selection
//   - RenderNone: No render, just the prompt
func GatherInput(v zcgame.GameView, inputNeeded *zcgame.PlayerInputNeeded) int {
	// Render the appropriate game state
	switch inputNeeded.RenderType {
	case zcgame.RenderNormal:
		RefreshRender(v)
	case zcgame.RenderForDiscard:
		refreshRenderForDiscard(v)
	case zcgame.RenderForNight:
		refreshRenderForNight(v)
	case zcgame.RenderNone:
		// No render
	}

	fmt.Printf("%s: ", inputNeeded.Message)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		if !scanner.Scan() {
			continue
		}
		text := scanner.Text()
		input, err := strconv.Atoi(text)
		if err != nil {
			fmt.Printf("ERROR, retry input: %s\n", intSliceChoices(inputNeeded.ValidChoices...))
			continue
		}

		for _, choice := range inputNeeded.ValidChoices {
			if input == choice {
				return input
			}
		}
		fmt.Printf("ERROR, retry input: %s\n", intSliceChoices(inputNeeded.ValidChoices...))
	}
}
