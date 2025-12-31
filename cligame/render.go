package cligame

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/ninesl/zombie-chickens/zcgame"
)

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

// statsString returns a summary of game statistics
func statsString(v zcgame.GameView) string {
	discardedNight := v.DiscardedNightCards()
	zombiesKilled := 0
	eventsPlayed := 0
	for _, card := range discardedNight {
		if card.IsZombie() {
			zombiesKilled++
		} else if card.IsEvent() {
			eventsPlayed++
		}
	}

	discardedDay := v.DiscardedDayCards()
	dayCardsDiscarded := 0
	for _, count := range discardedDay {
		dayCardsDiscarded += count
	}

	return fmt.Sprintf("Zombies Killed: %d | Events Played: %d | Day Cards Discarded: %d", zombiesKilled, eventsPlayed, dayCardsDiscarded)
}

// gameString returns the CLI-formatted game state
func gameString(v zcgame.GameView) string {
	players := v.Players()
	currentPlayerIdx := v.CurrentPlayerIdx()
	turn := v.Turn()

	result := fmt.Sprintf("%s %d\n%s\n---\n", TurnString(turn), v.NightNum(), PublicDayCardsString(v.PublicDayCards()))
	for i, pv := range players {
		isCurrentPlayer := i == currentPlayerIdx
		result += PlayerString(pv.Name(), pv.Lives(), pv.NightCards(), pv.Stacks(), pv.Hand(), isCurrentPlayer, turn, i)
		if i < len(players)-1 {
			result += "\n---\n"
		}
	}
	result += fmt.Sprintf("\n---\n%s\n", StageInTurnString(v.StageInTurn()))
	return result
}

// RefreshRender clears the terminal and prints the current game state.
// This is the standard render used during day turns.
func RefreshRender(v zcgame.GameView) {
	// Sort all players' hands and stacks before rendering
	players := v.Players()
	for range players {
		// Note: Can't sort via view since it returns copies
		// The actual sorting happens in the game state
	}

	clearScreen()
	fmt.Printf("%s\n", statsString(v))
	fmt.Printf("%s", gameString(v))
	if v.PlayerCount() > 0 {
		fmt.Printf("%s's %s\n", ColorPlayerName(v.CurrentPlayer().Name(), v.CurrentPlayerIdx()), TurnString(v.Turn()))
	}
}

// refreshRenderForDiscard renders the game state with farm item indices shown and hand indices hidden.
// Used during Lightning Storm and Tornado events.
func refreshRenderForDiscard(v zcgame.GameView) {
	players := v.Players()
	currentPlayerIdx := v.CurrentPlayerIdx()
	turn := v.Turn()

	clearScreen()
	fmt.Printf("%s\n", statsString(v))
	fmt.Printf("%s %d\n%s\n---\n", TurnString(turn), v.NightNum(), PublicDayCardsString(v.PublicDayCards()))
	for i, pv := range players {
		isCurrentPlayer := i == currentPlayerIdx
		fmt.Printf("%s", PlayerStringForDiscard(pv.Name(), pv.Lives(), pv.NightCards(), pv.Stacks(), pv.Hand(), isCurrentPlayer, turn, i))
		if i < len(players)-1 {
			fmt.Printf("\n---\n")
		}
	}
	fmt.Printf("\n---\n%s\n", StageInTurnString(v.StageInTurn()))
	if v.PlayerCount() > 0 {
		fmt.Printf("%s's %s\n", ColorPlayerName(v.CurrentPlayer().Name(), v.CurrentPlayerIdx()), TurnString(turn))
	}
}

// refreshRenderForNight renders the game state with farm stack indices shown and hand indices hidden.
// Used during zombie attacks at night.
func refreshRenderForNight(v zcgame.GameView) {
	players := v.Players()
	currentPlayerIdx := v.CurrentPlayerIdx()
	turn := v.Turn()

	clearScreen()
	fmt.Printf("%s\n", statsString(v))
	fmt.Printf("%s %d\n%s\n---\n", TurnString(turn), v.NightNum(), PublicDayCardsString(v.PublicDayCards()))
	for i, pv := range players {
		isCurrentPlayer := i == currentPlayerIdx
		fmt.Printf("%s", PlayerStringForNight(pv.Name(), pv.Lives(), pv.NightCards(), pv.Stacks(), pv.Hand(), isCurrentPlayer, turn, i))
		if i < len(players)-1 {
			fmt.Printf("\n---\n")
		}
	}
	fmt.Printf("\n---\n%s\n", StageInTurnString(v.StageInTurn()))
	if v.PlayerCount() > 0 {
		fmt.Printf("%s's %s\n", ColorPlayerName(v.CurrentPlayer().Name(), v.CurrentPlayerIdx()), TurnString(turn))
	}
}
