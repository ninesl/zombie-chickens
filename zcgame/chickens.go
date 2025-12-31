package zcgame

// ZombieChickens contains all zombie types that can appear in the night deck.
// Each zombie has a unique combination of traits that determine which defenses
// are effective against it. The map key is used as ZombieKey in NightCard.
var ZombieChickens = map[int]ZombieChicken{
	1: {
		Name:      "Raider",
		NumInDeck: 2,
		Traits:    []ZombieTrait{Flying, Bulletproof},
	},
	2: {
		Name:      "Walker",
		NumInDeck: 4,
		Traits:    []ZombieTrait{Fireproof, Exploding},
	},
	3: {
		Name:      "Chomper",
		NumInDeck: 2,
		Traits:    []ZombieTrait{Bulletproof, Fireproof, Timid},
	},
	4: {
		Name:      "Crawler",
		NumInDeck: 2,
		Traits:    []ZombieTrait{Climbing, Bulletproof},
	},
	5: {
		Name:      "Climber",
		NumInDeck: 2,
		Traits:    []ZombieTrait{Climbing, Fireproof, Exploding},
	},
	6: {
		Name:      "Clucker",
		NumInDeck: 2,
		Traits:    []ZombieTrait{Exploding},
	},
	7: {
		Name:      "Kablooey",
		NumInDeck: 4,
		Traits:    []ZombieTrait{Flying, Exploding},
	},
	8: {
		Name:      "Biter",
		NumInDeck: 10,
		Traits:    []ZombieTrait{Flying, Fireproof},
	},
	9: {
		Name:      "Blaster",
		NumInDeck: 2,
		Traits:    []ZombieTrait{Flying, Timid, Exploding},
	},
	10: {
		Name:      "Boomer",
		NumInDeck: 6,
		Traits:    []ZombieTrait{Flying, Bulletproof, Exploding},
	},
	11: {
		Name:      "Stalker",
		NumInDeck: 4,
		Traits:    []ZombieTrait{Invisible, Exploding},
	},
	12: {
		Name:      "Thunder",
		NumInDeck: 2,
		Traits:    []ZombieTrait{Invisible, Flying, Timid, Exploding},
	},
	13: {
		Name:      "Floater",
		NumInDeck: 2,
		Traits:    []ZombieTrait{Invisible, Flying, Timid},
	},
	14: {
		Name:      "Toaster",
		NumInDeck: 2,
		Traits:    []ZombieTrait{Flying, Fireproof, Timid, Exploding},
	},
	15: {
		Name:      "Sneaker",
		NumInDeck: 2,
		Traits:    []ZombieTrait{Invisible, Climbing},
	},
	16: {
		Name:      "Creeper",
		NumInDeck: 6,
		Traits:    []ZombieTrait{Invisible},
	},
}

// FindStacksThatCanKill returns indices of all stacks that can defeat the given zombie.
// It checks each stack against the zombie's traits to determine effectiveness.
func (f *Farm) FindStacksThatCanKill(zc ZombieChicken) []int {
	result := []int{}
	zt := zc.Traits

	for i, stack := range f.Stacks {
		// Scarecrow - only works on Timid zombies
		if zt.HasTrait(Timid) && stack.HasItem(Scarecrow) {
			result = append(result, i)
			continue
		}

		// Hay Wall (3 HayBales) - beaten by Flying or Climbing
		if !zt.HasTrait(Flying) && !zt.HasTrait(Climbing) && countItemInStack(stack, HayBale) >= 3 {
			result = append(result, i)
			continue
		}

		// Shotgun + Ammo - beaten by Bulletproof or Invisible
		if !zt.HasTrait(Bulletproof) && !zt.HasTrait(Invisible) && stack.HasItem(Shotgun) && stack.HasItem(Ammo) {
			result = append(result, i)
			continue
		}

		// Flamethrower + Fuel - beaten by Fireproof or Invisible
		if !zt.HasTrait(Fireproof) && !zt.HasTrait(Invisible) && stack.HasItem(Flamethrower) && stack.HasItem(Fuel) {
			result = append(result, i)
			continue
		}

		// BoobyTrap - beaten by Flying
		if !zt.HasTrait(Flying) && stack.HasItem(BoobyTrap) {
			result = append(result, i)
			continue
		}

		// WOLR - kills anything
		if stack.HasItem(WOLR) {
			result = append(result, i)
			continue
		}
	}

	return result
}

// HasTrait returns true if the trait slice contains the specified trait.
func (zt ZombieTraits) HasTrait(trait ZombieTrait) bool {
	for _, t := range zt {
		if t == trait {
			return true
		}
	}
	return false
}

// FindStacksThatCanKillForFree returns indices of stacks that can defeat the zombie
// without consuming any items. This is used for automatic defense selection.
//
// A defense is "free" if:
//   - The zombie is not Exploding (which destroys the stack)
//   - The stack does not contain one-time-use items (BoobyTrap, WOLR, Ammo)
//
// Free defenses include: Scarecrow (vs Timid), Hay Wall, Flamethrower+Fuel.
func (f *Farm) FindStacksThatCanKillForFree(zc ZombieChicken) []int {
	// Exploding zombies are never free - they destroy the stack
	if zc.Traits.HasTrait(Exploding) {
		return []int{}
	}

	allStacks := f.FindStacksThatCanKill(zc)
	result := []int{}

	for _, idx := range allStacks {
		stack := f.Stacks[idx]

		// Skip one-time-use items: BoobyTrap, WOLR, Shotgun+Ammo
		if stack.HasItem(BoobyTrap) || stack.HasItem(WOLR) || stack.HasItem(Ammo) {
			continue
		}

		// Scarecrow, Hay Wall, Flamethrower+Fuel are free
		result = append(result, idx)
	}

	return result
}

// DescribeDefense returns a human-readable description of what defense a stack provides.
// This is used for display messages when a zombie is killed.
func (s Stack) DescribeDefense(zc ZombieChicken) string {
	// Check in order of specificity
	if s.HasItem(WOLR) {
		return "W.O.L.R."
	}
	if s.HasItem(Scarecrow) && zc.Traits.HasTrait(Timid) {
		return "Scarecrow"
	}
	if countItemInStack(s, HayBale) >= 3 {
		return "Hay Wall"
	}
	if s.HasItem(Shotgun) && s.HasItem(Ammo) {
		return "Shotgun"
	}
	if s.HasItem(Flamethrower) && s.HasItem(Fuel) {
		return "Flamethrower"
	}
	if s.HasItem(BoobyTrap) {
		return "Booby Trap"
	}
	return "defense" // fallback
}

// UseDefenseStack uses the defense at the given stack index to defeat a zombie.
// This handles all the side effects of using a defense:
//   - WOLR destroys the entire farm
//   - Exploding zombies destroy the stack (unless useShield is true)
//   - One-time-use items (Ammo, BoobyTrap) are discarded
//   - Shield is consumed if useShield is true
func (f *Farm) UseDefenseStack(stackIdx int, zc ZombieChicken, useShield bool, g *gameState) {
	stack := f.Stacks[stackIdx]

	// WOLR destroys everything on the farm - handle first since it overrides all other logic
	if stack.HasItem(WOLR) {
		for _, s := range f.Stacks {
			for _, item := range s {
				g.discardDayCard(item)
			}
		}
		f.Stacks = Stacks{}
		return
	}

	// If Exploding, destroy the entire stack (unless shield is used)
	if zc.Traits.HasTrait(Exploding) {
		if useShield {
			// Remove shield from whichever stack has it
			for i := range f.Stacks {
				if f.Stacks[i].HasItem(Shield) {
					f.Stacks[i].RemoveItem(Shield)
					g.discardDayCard(Shield)
					break
				}
			}
		} else {
			// Discard all items in the destroyed stack
			for _, item := range f.Stacks[stackIdx] {
				g.discardDayCard(item)
			}
			f.Stacks[stackIdx] = Stack{}
			f.clearStacks()
			return
		}
	}

	// Remove one-time-use items
	if stack.HasItem(Ammo) {
		f.Stacks[stackIdx].RemoveItem(Ammo)
		g.discardDayCard(Ammo)
	}
	if stack.HasItem(BoobyTrap) {
		f.Stacks[stackIdx].RemoveItem(BoobyTrap)
		g.discardDayCard(BoobyTrap)
	}

	f.clearStacks()
}
