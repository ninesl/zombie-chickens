package zcgame

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

// CLIMode controls whether ANSI escape codes are used in String() output
var CLIMode = true

// ANSI escape codes
const (
	Reset        = "\033[0m"
	Bold         = "\033[1m"
	Italic       = "\033[3m"
	Red          = "\033[31m"
	Green        = "\033[32m"
	Yellow       = "\033[33m"
	Blue         = "\033[34m"
	Purple       = "\033[35m"
	Orange       = "\033[38;5;208m"
	BrightGreen  = "\033[92m"
	BrightBlue   = "\033[94m"
	BrightPurple = "\033[95m"
	BrightGrey   = "\033[38;5;105m"
)

// RedStar for one-time-use items (CLI only)
func redStar() string {
	if CLIMode {
		return Red + "*" + Reset
	}
	return "*"
}

// PlayerColors for players 1-4
var PlayerColors = []string{Green, Yellow, BrightPurple, BrightBlue}

// PlayerColor returns the color for the given player index, or empty string if not CLI mode
func PlayerColor(idx int) string {
	if CLIMode {
		return PlayerColors[idx%len(PlayerColors)]
	}
	return ""
}

// ClearScreen clears the terminal screen
func ClearScreen() {
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
	cmd.Run()
}

// RefreshRender clears screen and prints game state
func RefreshRender(g *GameState) {
	ClearScreen()
	fmt.Printf("%s\n", g.StatsString())
	fmt.Printf("%s", g)
	if len(g.Players) > 0 {
		fmt.Printf("%s's %s\n", g.CurrentPlayer().Name, g.Turn)
	}
}

// RefreshRenderForDiscard renders the game state with farm item indices shown and hand indices hidden.
// Used during Lightning Storm and Tornado events.
func RefreshRenderForDiscard(g *GameState) {
	ClearScreen()
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

// RefreshRenderForNight renders the game state with farm stack indices shown and hand indices hidden.
// Used during zombie attacks at night.
func RefreshRenderForNight(g *GameState) {
	ClearScreen()
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
	// Render the appropriate game state
	switch inputNeeded.RenderType {
	case RenderNormal:
		RefreshRender(g)
	case RenderForDiscard:
		RefreshRenderForDiscard(g)
	case RenderForNight:
		RefreshRenderForNight(g)
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
			fmt.Printf("ERROR, retry input: %s\n", IntSliceChoices(inputNeeded.ValidChoices...))
			continue
		}

		for _, choice := range inputNeeded.ValidChoices {
			if input == choice {
				return input
			}
		}
		fmt.Printf("ERROR, retry input: %s\n", IntSliceChoices(inputNeeded.ValidChoices...))
	}
}
