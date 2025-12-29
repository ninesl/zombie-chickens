package zcgame

// countItemInStack returns how many times item appears in stack
func countItemInStack(stack Stack, item FarmItemType) int {
	count := 0
	for _, card := range stack {
		if card == item {
			count++
		}
	}
	return count
}

// findStacks returns indices of all stacks containing ALL of the specified items
// Example: findStacks(Stack{Shotgun}) finds all stacks with at least 1 Shotgun
// Example: findStacks(Stack{Shotgun, Ammo}) finds all stacks with both Shotgun AND Ammo
func (f *Farm) findStacks(items Stack) []int {
	result := []int{}
	for i, stack := range f.Stacks {
		hasAll := true
		for _, item := range items {
			if !stack.HasItem(item) {
				hasAll = false
				break
			}
		}
		if hasAll {
			result = append(result, i)
		}
	}
	return result
}

func (f *Farm) PlayCard(item FarmItemType, choices PlayerPlayChoices) *PlayCardResult {
	if f.Stacks == nil {
		f.Stacks = make([]Stack, 0)
	}

	// Empty farm - always create new stack
	if len(f.Stacks) == 0 {
		f.makeStackWith(item)
		return nil
	}

	switch item {
	case Scarecrow, BoobyTrap, Shield, WOLR:
		// Simple items - always new stack
		f.makeStackWith(item)
	case Flamethrower:
		// Find fuel-only stacks (has Fuel, no Flamethrower)
		fuelStacks := f.findStacks(Stack{Fuel})
		fuelOnlyStacks := []int{}
		for _, idx := range fuelStacks {
			if !f.Stacks[idx].HasItem(Flamethrower) {
				fuelOnlyStacks = append(fuelOnlyStacks, idx)
			}
		}
		if len(fuelOnlyStacks) > 0 {
			// Add to first unpaired Fuel
			f.Stacks[fuelOnlyStacks[0]] = append(f.Stacks[fuelOnlyStacks[0]], Flamethrower)
		} else {
			f.makeStackWith(Flamethrower)
		}
	case Fuel:
		// Find flamethrower stacks without Fuel
		flamethrowerStacks := f.findStacks(Stack{Flamethrower})
		unfueledStacks := []int{}
		for _, idx := range flamethrowerStacks {
			if !f.Stacks[idx].HasItem(Fuel) {
				unfueledStacks = append(unfueledStacks, idx)
			}
		}
		if len(unfueledStacks) > 0 {
			// Add to first unpaired Flamethrower
			f.Stacks[unfueledStacks[0]] = append(f.Stacks[unfueledStacks[0]], Fuel)
		} else {
			f.makeStackWith(Fuel)
		}
	case Shotgun:
		// Find ammo-only stacks (has Ammo, no Shotgun)
		ammoStacks := f.findStacks(Stack{Ammo})
		ammoOnlyStacks := []int{}
		for _, idx := range ammoStacks {
			if !f.Stacks[idx].HasItem(Shotgun) {
				ammoOnlyStacks = append(ammoOnlyStacks, idx)
			}
		}
		if len(ammoOnlyStacks) > 0 {
			if choices.AutoloadShotgun {
				// Find the ammo stack with MOST ammo (same stack always if choice was always on)
				maxAmmoCount := 0
				maxAmmoIdx := ammoOnlyStacks[0]
				for _, idx := range ammoOnlyStacks {
					count := countItemInStack(f.Stacks[idx], Ammo)
					if count > maxAmmoCount {
						maxAmmoCount = count
						maxAmmoIdx = idx
					}
				}
				f.Stacks[maxAmmoIdx] = append(f.Stacks[maxAmmoIdx], Shotgun)
			} else {
				return &PlayCardResult{
					ValidStacks: ammoStacks,
					Message:     "choose to load shotgun with ammo or start new stack",
				}
			}
		} else {
			f.makeStackWith(Shotgun)
		}
	case Ammo:
		shotgunStacks := f.findStacks(Stack{Shotgun})
		if len(shotgunStacks) == 0 {
			// No shotguns - stack ammo in pairs of 2
			ammoStacks := f.findStacks(Stack{Ammo})
			ammoOnlyStacks := []int{}
			for _, idx := range ammoStacks {
				if !f.Stacks[idx].HasItem(Shotgun) {
					ammoOnlyStacks = append(ammoOnlyStacks, idx)
				}
			}

			// Find stacks with exactly 1 ammo (incomplete pairs)
			incompleteStacks := []int{}
			for _, idx := range ammoOnlyStacks {
				count := countItemInStack(f.Stacks[idx], Ammo)
				if count == 1 {
					incompleteStacks = append(incompleteStacks, idx)
				}
			}

			if len(incompleteStacks) > 0 {
				// Add to first incomplete stack (lowest index)
				f.Stacks[incompleteStacks[0]] = append(f.Stacks[incompleteStacks[0]], Ammo)
			} else {
				// All stacks have 2 ammo or no ammo stacks exist - create new stack
				f.makeStackWith(Ammo)
			}
		} else if choices.AutoloadShotgun {
			// Find shotgun with LEAST ammo (ties go to first index)
			minAmmoCount := -1
			minAmmoIdx := shotgunStacks[0]
			for _, idx := range shotgunStacks {
				count := countItemInStack(f.Stacks[idx], Ammo)
				if minAmmoCount == -1 || count < minAmmoCount {
					minAmmoCount = count
					minAmmoIdx = idx
				} else if count == minAmmoCount && idx < minAmmoIdx {
					minAmmoIdx = idx // tie-breaker: first index
				}
			}
			f.Stacks[minAmmoIdx] = append(f.Stacks[minAmmoIdx], Ammo)
		} else if len(shotgunStacks) == 1 {
			numAmmo := countItemInStack(f.Stacks[shotgunStacks[0]], Ammo)
			if numAmmo == 0 { // TODO: could be annoying as this is auto behavior without setting on
				f.Stacks[shotgunStacks[0]] = append(f.Stacks[shotgunStacks[0]], Ammo)
			} else {
				return &PlayCardResult{
					ValidStacks: shotgunStacks,
					Message:     "choose to load shotgun or start new ammo stack",
				}
			}
		} else {
			return &PlayCardResult{
				ValidStacks: shotgunStacks,
				Message:     "choose which shotgun to load with ammo",
			}
		}
	case HayBale:
		hayBaleStacks := f.findStacks(Stack{HayBale})
		// Filter for incomplete walls (<3 HayBales)
		incompleteWalls := []int{}
		for _, idx := range hayBaleStacks {
			count := countItemInStack(f.Stacks[idx], HayBale)
			if count < 3 {
				incompleteWalls = append(incompleteWalls, idx)
			}
		}
		if len(incompleteWalls) == 0 {
			// No incomplete walls - create new stack
			f.makeStackWith(HayBale)
		} else if choices.AutoBuildHayWall {
			// Find wall with MOST HayBales (but still <3)
			maxCount := 0
			maxIdx := incompleteWalls[0]
			for _, idx := range incompleteWalls {
				count := countItemInStack(f.Stacks[idx], HayBale)
				if count > maxCount {
					maxCount = count
					maxIdx = idx
				} else if count == maxCount && idx < maxIdx {
					maxIdx = idx // tie-breaker: first index
				}
			}
			f.Stacks[maxIdx] = append(f.Stacks[maxIdx], HayBale)
		} else if len(incompleteWalls) == 1 {
			numHay := countItemInStack(f.Stacks[incompleteWalls[0]], HayBale)
			if numHay == 1 { // TODO: could be annoying as this is auto behavior without setting on
				f.Stacks[incompleteWalls[0]] = append(f.Stacks[incompleteWalls[0]], HayBale)
			} else {
				return &PlayCardResult{
					ValidStacks: incompleteWalls,
					Message:     "choose to complete wall or start new one",
				}
			}
		} else {
			return &PlayCardResult{
				ValidStacks: incompleteWalls,
				Message:     "choose which hay wall to build",
			}
		}
	}

	f.clearStacks()
	return nil
}

func (f *Farm) makeStackWith(item FarmItemType) {
	var stack = make(Stack, 0, 1)
	stack = append(stack, item)
	f.Stacks = append(f.Stacks, stack)
}

// PlayCardResult is returned when PlayCard needs player input to decide
// where to place a card. A nil result means the card was played successfully.
type PlayCardResult struct {
	ValidStacks []int  // valid stack indices the player can choose from
	Message     string // human-readable prompt
}

// InputContext specifies what type of input is being requested
type InputContext uint8

const (
	InputContextPlayCard     InputContext = iota // Choosing where to play a card (stack selection)
	InputContextDiscard                          // Optional discard phase
	InputContextPlay                             // Play card from hand phase
	InputContextDraw                             // Draw cards phase (public vs deck)
	InputContextDefense                          // Night: choose defense stack
	InputContextShield                           // Night: use shield for exploding zombie?
	InputContextConfirm                          // Press 0 to continue
	InputContextEventDiscard                     // Discard cards during event (Lightning Storm, Tornado)
)

// RenderType specifies which render function to use
type RenderType uint8

const (
	RenderNormal RenderType = iota
	RenderForDiscard
	RenderForNight
	RenderNone
)

// PlayerInputNeeded signals that the game state machine requires player input to continue.
// It implements the error interface for historical compatibility with early error-based flow control,
// but it is not an error - it's a normal control flow signal in the state machine pattern.
// The game loop checks for this type to pause execution and gather input from CLI or API.
//
// In CLI mode, this triggers input gathering from stdin.
// In API mode, this should be returned to the caller to request input.
//
// Example usage pattern:
//
//	gameOver, inputNeeded := game.ContinueDay()
//	if inputNeeded != nil {
//	    input := gatherInput(inputNeeded)
//	    gameOver, inputNeeded = game.ContinueAfterInput(input)
//	}
type PlayerInputNeeded struct {
	Context      InputContext
	RenderType   RenderType
	Message      string // human-readable prompt
	ValidChoices []int  // valid input values

	// Optional context for specific input types
	Item        FarmItemType // for InputContextPlayCard - which item needs placement
	ValidStacks []int        // for InputContextPlayCard/InputContextDefense - valid stack indices
}

// Error implements the error interface. PlayerInputNeeded is used as a signal type
// rather than an actual error, allowing type assertions in the game loop.
func (e *PlayerInputNeeded) Error() string {
	return "needs player input: " + e.Message
}

// addToStackIndex adds item to f.Stacks[stackIndex]. Only called internally by the state machine
// with pre-validated indices from ValidChoices, so invalid indices indicate a bug and are ignored.
func (f *Farm) addToStackIndex(item FarmItemType, stackIndex int) {
	if stackIndex < 0 || stackIndex >= len(f.Stacks) {
		return
	}
	f.Stacks[stackIndex] = append(f.Stacks[stackIndex], item)
}

func (s Stack) HasItem(item FarmItemType) bool {
	for _, card := range s {
		if card == item {
			return true
		}
	}
	return false
}

// HasItemInStacks returns true if any stack on the farm contains the specified item.
func (f *Farm) HasItemInStacks(item FarmItemType) bool {
	for _, stack := range f.Stacks {
		if stack.HasItem(item) {
			return true
		}
	}
	return false
}

// removes all Stacks from f.Stacks where len(f.Stacks[i]) == 0
func (f *Farm) clearStacks() {
	for i := len(f.Stacks) - 1; i >= 0; i-- {
		if len(f.Stacks[i]) == 0 {
			f.Stacks = append(f.Stacks[:i], f.Stacks[i+1:]...)
		}
	}
}
