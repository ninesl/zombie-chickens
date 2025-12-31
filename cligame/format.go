// Package cligame implements the CLI interface for the zombie-chickens game.
// This package handles all terminal-specific functionality including ANSI colors,
// screen rendering, and input gathering.
package cligame

import (
	"fmt"

	"github.com/ninesl/zombie-chickens/zcgame"
)

// ANSI escape codes for terminal coloring
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

// PlayerColors for players 1-4
var PlayerColors = []string{Green, Yellow, BrightPurple, BrightBlue}

// RedStar returns a red asterisk for one-time-use items
func RedStar() string {
	return Red + "*" + Reset
}

// PlayerColor returns the color for the given player index
func PlayerColor(idx int) string {
	return PlayerColors[idx%len(PlayerColors)]
}

// ColorPlayerName applies color to a player name based on their index
func ColorPlayerName(name string, idx int) string {
	return PlayerColor(idx) + name + Reset
}

// FarmItemString returns the CLI-formatted string for a FarmItemType with ANSI colors
func FarmItemString(f zcgame.FarmItemType) string {
	switch f {
	case zcgame.HayBale:
		return BrightGrey + Italic + "Hay Bale" + Reset
	case zcgame.Scarecrow:
		return BrightGrey + Italic + "Scarecrow" + Reset
	case zcgame.Shotgun:
		return BrightGrey + Italic + "Shotgun" + Reset
	case zcgame.Ammo:
		return BrightGrey + Italic + "Ammo" + Reset + RedStar()
	case zcgame.BoobyTrap:
		return BrightGrey + Italic + "Booby Trap" + Reset + RedStar()
	case zcgame.Shield:
		return BrightGrey + Italic + "Shield" + Reset + RedStar()
	case zcgame.Flamethrower:
		return BrightGrey + Italic + "Flamethrower" + Reset
	case zcgame.Fuel:
		return BrightGrey + Italic + "Fuel" + Reset
	case zcgame.WOLR:
		return BrightGrey + Italic + "W.O.L.R" + Reset + RedStar()
	default:
		return fmt.Sprintf("FarmItemType ERROR %d", int(f))
	}
}

// TurnString returns the CLI-formatted string for a Turn with ANSI colors
func TurnString(t zcgame.Turn) string {
	switch t {
	case zcgame.Morning:
		return BrightBlue + Italic + "Morning" + Reset
	case zcgame.Afternoon:
		return Orange + Italic + "Afternoon" + Reset
	case zcgame.Night:
		return BrightPurple + Italic + "Night" + Reset
	case zcgame.Day:
		return Italic + "Day" + Reset
	default:
		return fmt.Sprintf("Turn ERROR %d", int(t))
	}
}

// ZombieTraitString returns the CLI-formatted string for a ZombieTrait with ANSI colors
func ZombieTraitString(zt zcgame.ZombieTrait) string {
	switch zt {
	case zcgame.Invisible:
		return Purple + "Invisible" + Reset
	case zcgame.Flying:
		return BrightBlue + "Flying" + Reset
	case zcgame.Climbing:
		return Yellow + "Climbing" + Reset
	case zcgame.Bulletproof:
		return Blue + "Bulletproof" + Reset
	case zcgame.Fireproof:
		return Red + "Fireproof" + Reset
	case zcgame.Timid:
		return BrightGreen + "Timid" + Reset
	case zcgame.Exploding:
		return Orange + "Exploding" + Reset
	default:
		return fmt.Sprintf("ZombieTrait ERROR %d", int(zt))
	}
}

// StageInTurnString returns the CLI-formatted string for a StageInTurn with ANSI colors
func StageInTurnString(s zcgame.StageInTurn) string {
	switch s {
	case zcgame.OptionalDiscard:
		return Bold + Italic + "Discard a card to draw a card from the deck (optional)" + Reset
	case zcgame.Play2Cards:
		return Bold + Italic + "Play 2 cards to your farm" + Reset
	case zcgame.Draw2Cards:
		return Bold + Italic + "Draw 2 cards from the deck or the 2 face-up cards" + Reset
	case zcgame.Nighttime:
		return Bold + Italic + "Progress through the night..." + Reset
	default:
		return fmt.Sprintf("StageInTurn ERROR %d", int(s))
	}
}

// EventString returns the CLI-formatted string for an Event with ANSI colors
func EventString(name, description string) string {
	return Bold + name + Reset + "\n| " + Italic + description + Reset + " |"
}

// ZombieTraitsString returns the CLI-formatted string for ZombieTraits with ANSI colors
func ZombieTraitsString(traits zcgame.ZombieTraits) string {
	result := "|"
	for _, trait := range traits {
		result += fmt.Sprintf(" %s |", ZombieTraitString(trait))
	}
	return result
}

// ZombieChickenString returns the CLI-formatted string for a ZombieChicken with ANSI colors
func ZombieChickenString(name string, traits zcgame.ZombieTraits) string {
	return fmt.Sprintf("%s%s%s\n%s", Bold, name, Reset, ZombieTraitsString(traits))
}

// StackString returns the CLI-formatted string for a Stack
func StackString(s zcgame.Stack) string {
	result := "{ "
	for i, item := range s {
		result += FarmItemString(item)
		if i < len(s)-1 {
			result += ", "
		}
	}
	result += " }"
	return result
}

// StackStringWithIndices returns the CLI-formatted string for a Stack with numbered indices
func StackStringWithIndices(s zcgame.Stack, idx *int) string {
	result := "{ "
	for i, item := range s {
		result += fmt.Sprintf("%d: %s", *idx, FarmItemString(item))
		*idx++
		if i < len(s)-1 {
			result += ", "
		}
	}
	result += " }"
	return result
}

// StacksString returns the CLI-formatted string for Stacks
func StacksString(stacks zcgame.Stacks) string {
	result := ""
	for i, stack := range stacks {
		result += StackString(stack)
		if i < len(stacks)-1 {
			result += "\n"
		}
	}
	return result
}

// StacksStringForDiscard returns the CLI-formatted string for Stacks with item indices
func StacksStringForDiscard(stacks zcgame.Stacks) string {
	result := ""
	idx := 1
	for i, stack := range stacks {
		result += StackStringWithIndices(stack, &idx)
		if i < len(stacks)-1 {
			result += "\n"
		}
	}
	return result
}

// StacksStringForNight returns the CLI-formatted string for Stacks with stack indices (1-based)
func StacksStringForNight(stacks zcgame.Stacks) string {
	result := ""
	for i, stack := range stacks {
		result += fmt.Sprintf("%d:%s", i+1, StackString(stack))
		if i < len(stacks)-1 {
			result += "\n"
		}
	}
	return result
}

// FarmString returns the CLI-formatted string for a Farm
func FarmString(stacks zcgame.Stacks) string {
	return fmt.Sprintf("Farm:\n%s", StacksString(stacks))
}

// FarmStringForDiscard returns the CLI-formatted string for a Farm with item indices
func FarmStringForDiscard(stacks zcgame.Stacks) string {
	return fmt.Sprintf("Farm:\n%s", StacksStringForDiscard(stacks))
}

// FarmStringForNight returns the CLI-formatted string for a Farm with stack indices
func FarmStringForNight(stacks zcgame.Stacks) string {
	return fmt.Sprintf("Farm:\n%s", StacksStringForNight(stacks))
}

// HandItemString returns the CLI-formatted string for a HandItem
func HandItemString(h zcgame.HandItem) string {
	if h.FarmItemType == zcgame.NUM_FARM_ITEMS {
		return ""
	}
	return FarmItemString(h.FarmItemType)
}

// HandString returns the CLI-formatted string for a Hand with indices
func HandString(h zcgame.Hand) string {
	return handStringWithIndices(h, true)
}

// HandStringWithoutIndices returns the CLI-formatted string for a Hand without indices
func HandStringWithoutIndices(h zcgame.Hand) string {
	return handStringWithIndices(h, false)
}

func handStringWithIndices(h zcgame.Hand, showIndices bool) string {
	result := "Hand: { "
	first := true
	idx := 1
	for _, card := range h {
		if card.FarmItemType == zcgame.NUM_FARM_ITEMS {
			continue
		}
		if !first {
			result += ", "
		}
		if showIndices {
			result += fmt.Sprintf("%d:%s", idx, FarmItemString(card.FarmItemType))
		} else {
			result += FarmItemString(card.FarmItemType)
		}
		idx++
		first = false
	}
	result += " }"
	return result
}

// PublicDayCardsString returns the CLI-formatted string for PublicDayCards
func PublicDayCardsString(cards zcgame.PublicDayCards) string {
	return StackString(zcgame.Stack(cards[:]))
}

// NightCardsString returns the CLI-formatted string for NightCards with visibility control
func NightCardsString(cards zcgame.NightCards, isCurrentPlayer bool, turn zcgame.Turn) string {
	// Only show night cards during Night
	if turn != zcgame.Night {
		return ""
	}

	if !isCurrentPlayer {
		return fmt.Sprintf("NightCard x %d", len(cards))
	}

	if len(cards) == 0 {
		return "NightCard x 0"
	}

	remainingCount := len(cards) - 1
	countStr := fmt.Sprintf("NightCard x %d", remainingCount)
	card := cards[0]

	if card.IsZombie() {
		zc := zcgame.ZombieChickens[card.ZombieKey]
		return fmt.Sprintf("%s\n%s", countStr, ZombieChickenString(zc.Name, zc.Traits))
	} else if card.IsEvent() {
		return fmt.Sprintf("%s\n%s", countStr, EventString(card.Event.Name, card.Event.Description))
	}

	return ""
}

// PlayerString returns the CLI-formatted string for a Player
func PlayerString(name string, lives int, nightCards zcgame.NightCards, stacks zcgame.Stacks, hand zcgame.Hand, isCurrentPlayer bool, turn zcgame.Turn, playerIdx int) string {
	nightCardsStr := NightCardsString(nightCards, isCurrentPlayer, turn)
	coloredName := ColorPlayerName(name, playerIdx)
	return fmt.Sprintf("%s : %dhp\n%s\n%s\n%s", coloredName, lives, nightCardsStr, FarmString(stacks), HandString(hand))
}

// PlayerStringForDiscard returns the CLI-formatted string for a Player during discard events
func PlayerStringForDiscard(name string, lives int, nightCards zcgame.NightCards, stacks zcgame.Stacks, hand zcgame.Hand, isCurrentPlayer bool, turn zcgame.Turn, playerIdx int) string {
	nightCardsStr := NightCardsString(nightCards, isCurrentPlayer, turn)
	coloredName := ColorPlayerName(name, playerIdx)
	return fmt.Sprintf("%s : %dhp\n%s\n%s\n%s", coloredName, lives, nightCardsStr, FarmStringForDiscard(stacks), HandStringWithoutIndices(hand))
}

// PlayerStringForNight returns the CLI-formatted string for a Player during night phase
func PlayerStringForNight(name string, lives int, nightCards zcgame.NightCards, stacks zcgame.Stacks, hand zcgame.Hand, isCurrentPlayer bool, turn zcgame.Turn, playerIdx int) string {
	nightCardsStr := NightCardsString(nightCards, isCurrentPlayer, turn)
	coloredName := ColorPlayerName(name, playerIdx)
	return fmt.Sprintf("%s : %dhp\n%s\n%s\n%s", coloredName, lives, nightCardsStr, FarmStringForNight(stacks), HandStringWithoutIndices(hand))
}
