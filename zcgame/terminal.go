package zcgame

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

// cliMode controls whether ANSI escape codes are used in String() output.
// Default is false (API mode). Use SetCLIMode(true) before creating a game for CLI usage.
var cliMode = false

// SetCLIMode enables or disables CLI mode for the package.
// When enabled, String() methods include ANSI escape codes for colored terminal output.
// When disabled (default), plain text is output for API/frontend consumption.
//
// This should be called once at startup before creating any games.
// Changing this value mid-game may cause inconsistent output formatting.
func SetCLIMode(enabled bool) {
	cliMode = enabled
}

// IsCLIMode returns true if CLI mode is currently enabled.
func IsCLIMode() bool {
	return cliMode
}

// ANSI escape codes (unexported - internal use only)
const (
	reset        = "\033[0m"
	bold         = "\033[1m"
	italic       = "\033[3m"
	red          = "\033[31m"
	green        = "\033[32m"
	yellow       = "\033[33m"
	blue         = "\033[34m"
	purple       = "\033[35m"
	orange       = "\033[38;5;208m"
	brightGreen  = "\033[92m"
	brightBlue   = "\033[94m"
	brightPurple = "\033[95m"
	brightGrey   = "\033[38;5;105m"
)

// redStar returns a red asterisk for one-time-use items (CLI only)
func redStar() string {
	if cliMode {
		return red + "*" + reset
	}
	return "*"
}

// playerColors for players 1-4
var playerColors = []string{green, yellow, brightPurple, brightBlue}

// playerColor returns the color for the given player index, or empty string if not CLI mode
func playerColor(idx int) string {
	if cliMode {
		return playerColors[idx%len(playerColors)]
	}
	return ""
}

// clearScreen clears the terminal screen
func clearScreen() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.Command("clear")
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		// Fallback: print ANSI escape code
		fmt.Print("\033[H\033[2J")
		return
	}

	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatalf("crashed when clearing screen: %v", err)
	}
}

// RefreshRender clears the terminal and prints the current game state.
// It sorts all players' hands and stacks before rendering for consistent display.
// This is the standard render used during day turns.
func RefreshRender(v GameView) {
	g := v.game

	for _, p := range g.Players {
		p.Hand.Sort()
		p.Farm.Stacks.Sort()
	}

	clearScreen()
	fmt.Printf("%s\n", g.StatsString())
	fmt.Printf("%s", g)
	if len(g.Players) > 0 {
		fmt.Printf("%s's %s\n", g.CurrentPlayer().Name, g.Turn)
	}
}

// refreshRenderForDiscard renders the game state with farm item indices shown and hand indices hidden.
// Used during Lightning Storm and Tornado events.
func refreshRenderForDiscard(v GameView) {
	g := v.game
	clearScreen()
	fmt.Printf("%s\n", g.StatsString())
	fmt.Printf("%s %d\n%s\n---\n", g.Turn, g.NightNum, g.PublicDayCards)
	for i, player := range g.Players {
		isCurrentPlayer := i == g.CurrentPlayerIdx
		nightCardsStr := player.Farm.NightCards.StringWithVisibility(isCurrentPlayer, g.Turn)
		fmt.Printf("%s : %dhp\n%s\n%s\n%s", player.Name, player.Lives, nightCardsStr, player.Farm.StringForDiscard(), player.Hand.StringWithoutIndices())
		if i < len(g.Players)-1 {
			fmt.Printf("\n---\n")
		}
	}
	fmt.Printf("\n---\n%s\n", g.StageInTurn)
	if len(g.Players) > 0 {
		fmt.Printf("%s's %s\n", g.CurrentPlayer().Name, g.Turn)
	}
}

// refreshRenderForNight renders the game state with farm stack indices shown and hand indices hidden.
// Used during zombie attacks at night.
func refreshRenderForNight(v GameView) {
	g := v.game
	clearScreen()
	fmt.Printf("%s\n", g.StatsString())
	fmt.Printf("%s %d\n%s\n---\n", g.Turn, g.NightNum, g.PublicDayCards)
	for i, player := range g.Players {
		isCurrentPlayer := i == g.CurrentPlayerIdx
		nightCardsStr := player.Farm.NightCards.StringWithVisibility(isCurrentPlayer, g.Turn)
		fmt.Printf("%s : %dhp\n%s\n%s\n%s", player.Name, player.Lives, nightCardsStr, player.Farm.StringForNight(), player.Hand.StringWithoutIndices())
		if i < len(g.Players)-1 {
			fmt.Printf("\n---\n")
		}
	}
	fmt.Printf("\n---\n%s\n", g.StageInTurn)
	if len(g.Players) > 0 {
		fmt.Printf("%s's %s\n", g.CurrentPlayer().Name, g.Turn)
	}
}

// GatherCLIInput displays the game state and prompt, then reads player input from stdin.
// It validates input against ValidChoices and re-prompts on invalid input.
//
// This function should only be called when CLI mode is enabled.
// It will log.Fatal if called when cliMode is false.
//
// The render type from inputNeeded determines how the game state is displayed:
//   - RenderNormal: Standard view with hand indices
//   - RenderForDiscard: Farm item indices shown for event discards
//   - RenderForNight: Stack indices shown for defense selection
//   - RenderNone: No render, just the prompt
func GatherCLIInput(v GameView, inputNeeded *PlayerInputNeeded) int {
	if !cliMode {
		log.Fatalf("tried to get CLI input but cliMode is %v", cliMode)
	}

	// Render the appropriate game state
	switch inputNeeded.RenderType {
	case RenderNormal:
		RefreshRender(v)
	case RenderForDiscard:
		refreshRenderForDiscard(v)
	case RenderForNight:
		refreshRenderForNight(v)
	case RenderNone:
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
