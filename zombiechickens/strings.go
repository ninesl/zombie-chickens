package zombiechickens

import (
	"fmt"
	"sort"
)

func (f FarmItemType) String() string {
	switch f {
	case HayBale:
		return "Hay Bale"
	case Scarecrow:
		return "Scarecrow"
	case Shotgun:
		return "Shotgun"
	case Ammo:
		return "Ammo"
	case BoobyTrap:
		return "Booby Trap"
	case Shield:
		return "Shield"
	case Flamethrower:
		return "Flamethrower"
	case Fuel:
		return "Fuel"
	case WOLR:
		return "W.O.L.R"
	default:
		return fmt.Sprintf("FarmItemType ERROR %d", int(f))
	}
}

func (t Turn) String() string {
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

func (g *GameState) String() string {
	// var sb strings.Builder

	// sb.WriteString(g.Turn.String() + "\n")

	return fmt.Sprintf(`%s %d\n%s`, g.Turn, g.NightNum, g.Players)

	// what player 1 is fighting?
}

func (p Players) String() string {
	result := ""
	for _, player := range p {
		result += fmt.Sprintf("%s\n", player)
	}
	return result
}

func (p Player) String() string {
	return fmt.Sprintf("%s : %dhp\n%s\n%s\n%s", p.Name, p.Lives, p.Farm.NightCards, p.Farm, p.Hand)
}

// returns the first night card in a nicely formatted way
func (n NightCards) String() string {
	if len(n) == 0 {
		return ""
	}

	card := n[0]

	if card.IsZombie() {
		return fmt.Sprintf("%s", ZombieChickens[card.ZombieKey])
	} else if card.IsEvent() {
		return fmt.Sprintf("%s", card.Event)
	}

	panic("this should not happen")
}

func (e Event) String() string {
	return e.Name + "\n| " + e.Description + " |"
}

func (zt ZombieTraits) String() string {
	if len(zt) == 0 {
		panic("zombie should always have traits")
	}

	// Sort by ZombieTrait value
	sort.Slice(zt, func(i, j int) bool {
		return zt[i] < zt[j]
	})

	result := "|"
	for _, trait := range zt {
		result += fmt.Sprintf(" %s |", trait)
	}
	return result
}

func (z ZombieChicken) String() string {
	return fmt.Sprintf("%s\n%s", z.Name, z.Traits)
}

func (s Stacks) String() string {
	// Sort each inner stack by FarmItemType
	for i := range s {
		sort.Slice(s[i], func(a, b int) bool {
			return s[i][a] < s[i][b]
		})
	}

	// Sort stacks by first element
	sort.Slice(s, func(i, j int) bool {
		if len(s[i]) == 0 {
			return false
		}
		if len(s[j]) == 0 {
			return true
		}
		return s[i][0] < s[j][0]
	})

	result := ""
	for i, stack := range s {
		result += fmt.Sprintf("%s", stack)
		if i < len(s)-1 {
			// avoids trailing whitespace
			result += "\n"
		}
	}
	return result
}

func (f *Farm) String() string {
	return fmt.Sprintf("Farm:\n%s", f.Stacks)
}

func (h HandItem) String() string {
	return fmt.Sprintf("%s", h.FarmItemType)
}

func (h Hand) String() string {
	// Sort by visible (true first), then by FarmItemType
	sorted := h
	sort.Slice(sorted[:], func(i, j int) bool {
		if sorted[i].Visible != sorted[j].Visible {
			return sorted[i].Visible // true comes before false
		}
		return sorted[i].FarmItemType < sorted[j].FarmItemType
	})

	result := "Hand: "
	for i, card := range sorted {
		result += fmt.Sprintf("%s", card)
		if i < len(sorted)-1 {
			// avoids trailing comma
			result += ", "
		}
	}

	return result
}

func (s Stack) String() string {
	// Sort by FarmItemType
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})

	result := "{ "
	for _, item := range s {
		result += fmt.Sprintf("%s ", item)
	}
	result += "}"
	return result
}
