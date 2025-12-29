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

// SetCLIMode sets whether CLI mode is enabled.
// When true, ANSI escape codes are used for colored output.
// When false (default), plain text output is used for API/frontend consumption.
// This should be called once at startup before creating any games.
func SetCLIMode(enabled bool) {
	cliMode = enabled
}

// IsCLIMode returns whether CLI mode is currently enabled.
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

// RefreshRender clears screen and prints game state
func RefreshRender(g *GameState) {

	for _, p := range g.Players {
		p.Hand.Sort()
		for i := range p.Farm.Stacks {
			p.Farm.Stacks[i].Sort()
		}
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
func refreshRenderForDiscard(g *GameState) {
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
func refreshRenderForNight(g *GameState) {
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

// GatherCLIInput displays the prompt and gathers input from the user via CLI.
// This should only be called in CLI mode.
func GatherCLIInput(g *GameState, inputNeeded *PlayerInputNeeded) int {
	if !cliMode {
		log.Fatalf("tried to get CLI input but cliMode is %v", cliMode)
	}

	// Render the appropriate game state
	switch inputNeeded.RenderType {
	case RenderNormal:
		RefreshRender(g)
	case RenderForDiscard:
		refreshRenderForDiscard(g)
	case RenderForNight:
		refreshRenderForNight(g)
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
