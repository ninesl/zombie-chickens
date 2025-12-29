package zcgame

import (
	"fmt"
	"log"
)

func (f FarmItemType) String() string {
	if cliMode {
		switch f {
		case HayBale:
			return brightGrey + italic + "Hay Bale" + reset
		case Scarecrow:
			return brightGrey + italic + "Scarecrow" + reset
		case Shotgun:
			return brightGrey + italic + "Shotgun" + reset
		case Ammo:
			return brightGrey + italic + "Ammo" + reset + redStar()
		case BoobyTrap:
			return brightGrey + italic + "Booby Trap" + reset + redStar()
		case Shield:
			return brightGrey + italic + "Shield" + reset + redStar()
		case Flamethrower:
			return brightGrey + italic + "Flamethrower" + reset
		case Fuel:
			return brightGrey + italic + "Fuel" + reset
		case WOLR:
			return brightGrey + italic + "W.O.L.R" + reset + redStar()
		default:
			return fmt.Sprintf("FarmItemType ERROR %d", int(f))
		}
	}

	// Non-CLI mode
	switch f {
	case HayBale:
		return "Hay Bale"
	case Scarecrow:
		return "Scarecrow"
	case Shotgun:
		return "Shotgun"
	case Ammo:
		return "Ammo*"
	case BoobyTrap:
		return "Booby Trap*"
	case Shield:
		return "Shield*"
	case Flamethrower:
		return "Flamethrower"
	case Fuel:
		return "Fuel"
	case WOLR:
		return "W.O.L.R*"
	default:
		return fmt.Sprintf("FarmItemType ERROR %d", int(f))
	}
}

func (t Turn) String() string {
	if cliMode {
		switch t {
		case Morning:
			return brightBlue + italic + "Morning" + reset
		case Afternoon:
			return orange + italic + "Afternoon" + reset
		case Night:
			return brightPurple + italic + "Night" + reset
		case Day:
			return italic + "Day" + reset
		default:
			return fmt.Sprintf("Turn ERROR %d", int(t))
		}
	}

	// Non-CLI mode
	switch t {
	case Morning:
		return "Morning"
	case Afternoon:
		return "Afternoon"
	case Night:
		return "Night"
	case Day:
		return "Day"
	default:
		return fmt.Sprintf("Turn ERROR %d", int(t))
	}
}

func (zt ZombieTrait) String() string {
	if cliMode {
		switch zt {
		case Invisible:
			return purple + "Invisible" + reset
		case Flying:
			return brightBlue + "Flying" + reset
		case Climbing:
			return yellow + "Climbing" + reset
		case Bulletproof:
			return blue + "Bulletproof" + reset
		case Fireproof:
			return red + "Fireproof" + reset
		case Timid:
			return brightGreen + "Timid" + reset
		case Exploding:
			return orange + "Exploding" + reset
		default:
			return fmt.Sprintf("ZombieTrait ERROR %d", int(zt))
		}
	}

	// Non-CLI mode
	switch zt {
	case Invisible:
		return "Invisible"
	case Flying:
		return "Flying"
	case Climbing:
		return "Climbing"
	case Bulletproof:
		return "Bulletproof"
	case Fireproof:
		return "Fireproof"
	case Timid:
		return "Timid"
	case Exploding:
		return "Exploding"
	default:
		return fmt.Sprintf("ZombieTrait ERROR %d", int(zt))
	}
}

func intSliceChoices(s ...int) string {
	return fmt.Sprintf("%+v", s)
}

func (s StageInTurn) String() string {
	if cliMode {
		switch s {
		case OptionalDiscard:
			return bold + italic + "Discard a card to draw a card from the deck (optional)" + reset
		case Play2Cards:
			return bold + italic + "Play 2 cards to your farm" + reset
		case Draw2Cards:
			return bold + italic + "Draw 2 cards from the deck or the 2 face-up cards" + reset
		case Nighttime:
			return bold + italic + "Progress through the night..." + reset
		default:
			return fmt.Sprintf("StageInTurn ERROR %d", int(s))
		}
	}

	// Non-CLI mode
	switch s {
	case OptionalDiscard:
		return "Discard a card to draw a card from the deck (optional)"
	case Play2Cards:
		return "Play 2 cards to your farm"
	case Draw2Cards:
		return "Draw 2 cards from the deck or the 2 face-up cards"
	case Nighttime:
		return "Progress through the night..."
	default:
		return fmt.Sprintf("StageInTurn ERROR %d", int(s))
	}
}

func (g *GameState) String() string {
	result := fmt.Sprintf("%s %d\n%s\n---\n", g.Turn, g.NightNum, g.PublicDayCards)
	for i, player := range g.Players {
		isCurrentPlayer := i == g.CurrentPlayerIdx
		result += player.StringWithVisibility(isCurrentPlayer, g.Turn)
		if i < len(g.Players)-1 {
			result += "\n---\n"
		}
	}
	result += fmt.Sprintf("\n---\n%s\n", g.StageInTurn)
	return result
}

func (p Players) String() string {
	result := ""
	for i, player := range p {
		result += fmt.Sprintf("%s", player)
		if i < len(p)-1 {
			result += "\n---\n"
		}
	}
	return result
}

func (p *Player) String() string {
	return p.StringWithVisibility(true, Night)
}

func (p *Player) StringWithVisibility(isCurrentPlayer bool, turn Turn) string {
	nightCardsStr := p.Farm.NightCards.StringWithVisibility(isCurrentPlayer, turn)
	return fmt.Sprintf("%s : %dhp\n%s\n%s\n%s", p.Name, p.Lives, nightCardsStr, p.Farm, p.Hand.String())
}

// returns the first night card in a nicely formatted way
func (n NightCards) String() string {
	return n.StringWithVisibility(true, Night)
}

func (n NightCards) StringWithVisibility(isCurrentPlayer bool, turn Turn) string {
	// Only show night cards during Night
	if turn != Night {
		return ""
	}

	if !isCurrentPlayer {
		// Show total count for non-current players
		return fmt.Sprintf("NightCard x %d", len(n))
	}

	// For current player showing a card, count is remaining cards (len - 1)
	if len(n) == 0 {
		return "NightCard x 0"
	}

	remainingCount := len(n) - 1
	countStr := fmt.Sprintf("NightCard x %d", remainingCount)
	card := n[0]

	if card.IsZombie() {
		return fmt.Sprintf("%s\n%s", countStr, ZombieChickens[card.ZombieKey])
	} else if card.IsEvent() {
		return fmt.Sprintf("%s\n%s", countStr, card.Event)
	}

	log.Fatal("this should not happen")
	return ""
}

func (e Event) String() string {
	if cliMode {
		return bold + e.Name + reset + "\n| " + italic + e.Description + reset + " |"
	}
	return e.Name + "\n| " + e.Description + " |"
}

func (zt ZombieTraits) String() string {
	if len(zt) == 0 {
		log.Fatal("a zombie should always have traits")
	}

	// NOTE: Sorting commented out - order preserved as-is for frontend consistency
	// sort.Slice(zt, func(i, j int) bool {
	// 	return zt[i] < zt[j]
	// })

	result := "|"
	for _, trait := range zt {
		result += fmt.Sprintf(" %s |", trait)
	}
	return result
}

func (z ZombieChicken) String() string {
	if cliMode {
		return fmt.Sprintf("%s%s%s\n%s", bold, z.Name, reset, z.Traits)
	}
	return fmt.Sprintf("%s\n%s", z.Name, z.Traits)
}

func (s Stacks) String() string {
	return s.stringWithIndices(false)
}

// StringForDiscard returns a string representation with item indices for discarding.
func (s Stacks) StringForDiscard() string {
	return s.stringWithIndices(true)
}

// StringForNight returns a string representation with stack indices (1-based, not sorted).
func (s Stacks) StringForNight() string {
	result := ""
	for i, stack := range s {
		result += fmt.Sprintf("%d:%s", i+1, stack)
		if i < len(s)-1 {
			result += "\n"
		}
	}
	return result
}

// stringWithIndices returns a string representation of stacks.
// NOTE: Sorting commented out - order preserved as-is for frontend consistency
func (s Stacks) stringWithIndices(showIndices bool) string {
	// NOTE: Sorting commented out - order preserved as-is for frontend consistency
	// Sort each inner stack by FarmItemType
	// for i := range s {
	// 	sort.Slice(s[i], func(a, b int) bool {
	// 		return s[i][a] < s[i][b]
	// 	})
	// }

	// Sort stacks by first element
	// sort.Slice(s, func(i, j int) bool {
	// 	if len(s[i]) == 0 {
	// 		return false
	// 	}
	// 	if len(s[j]) == 0 {
	// 		return true
	// 	}
	// 	return s[i][0] < s[j][0]
	// })

	result := ""
	idx := 1
	for i, stack := range s {
		if showIndices {
			result += stack.stringWithIndices(&idx)
		} else {
			result += fmt.Sprintf("%s", stack)
		}
		if i < len(s)-1 {
			// avoids trailing whitespace
			result += "\n"
		}
	}
	return result
}

// TotalItems returns the total number of items across all stacks.
func (s Stacks) TotalItems() int {
	count := 0
	for _, stack := range s {
		count += len(stack)
	}
	return count
}

func (f *Farm) String() string {
	return fmt.Sprintf("Farm:\n%s", f.Stacks)
}

func (f *Farm) StringForDiscard() string {
	return fmt.Sprintf("Farm:\n%s", f.Stacks.StringForDiscard())
}

func (f *Farm) StringForNight() string {
	return fmt.Sprintf("Farm:\n%s", f.Stacks.StringForNight())
}

func (h HandItem) String() string {
	if h.FarmItemType == NUM_FARM_ITEMS {
		return ""
	}
	return fmt.Sprintf("%s", h.FarmItemType)
}

func (h *Hand) String() string {
	return h.stringWithIndices(true)
}

func (h *Hand) StringWithoutIndices() string {
	return h.stringWithIndices(false)
}

func (h *Hand) stringWithIndices(showIndices bool) string {
	// NOTE: Sorting commented out - order preserved as-is for frontend consistency
	// Hand is sorted explicitly before [3] and [4] assignments in game.go
	// h.Sort()

	result := "Hand: { "
	first := true
	idx := 1
	for _, card := range h {
		if card.FarmItemType == NUM_FARM_ITEMS {
			continue // skip blank slots
		}
		if !first {
			result += ", "
		}
		if showIndices {
			result += fmt.Sprintf("%d:%s", idx, card)
		} else {
			result += fmt.Sprintf("%s", card)
		}
		idx++
		first = false
	}
	result += " }"

	return result
}

func (s Stack) String() string {
	// NOTE: Sorting commented out - order preserved as-is for frontend consistency
	// sort.Slice(s, func(i, j int) bool {
	// 	return s[i] < s[j]
	// })

	result := "{ "
	for i, item := range s {
		result += fmt.Sprintf("%s", item)
		if i < len(s)-1 {
			result += ", "
		}
	}
	result += " }"
	return result
}

func (s Stack) stringWithIndices(idx *int) string {
	// NOTE: Sorting commented out - order preserved as-is for frontend consistency
	// sort.Slice(s, func(i, j int) bool {
	// 	return s[i] < s[j]
	// })

	result := "{ "
	for i, item := range s {
		result += fmt.Sprintf("%d: %s", *idx, item)
		*idx++
		if i < len(s)-1 {
			result += ", "
		}
	}
	result += " }"
	return result
}

func (g *GameState) StatsString() string {
	zombiesKilled := 0
	eventsPlayed := 0
	for _, card := range g.DiscardedNightCards {
		if card.IsZombie() {
			zombiesKilled++
		} else if card.IsEvent() {
			eventsPlayed++
		}
	}

	dayCardsDiscarded := 0
	for _, count := range g.DiscardedDayCards {
		dayCardsDiscarded += count
	}

	return fmt.Sprintf("Zombies Killed: %d | Events Played: %d | Day Cards Discarded: %d", zombiesKilled, eventsPlayed, dayCardsDiscarded)
}
