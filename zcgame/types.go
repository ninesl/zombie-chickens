// Package zcgame implements the core game logic for Zombie Chickens, a card game
// where players defend their farms from zombie chicken attacks.
//
// The package provides a state machine-based game engine that supports both CLI
// and API usage patterns through interruptible execution. When player input is
// required, the engine returns a PlayerInputNeeded signal containing valid choices,
// allowing the caller to gather input and resume execution.
//
// Key types:
//   - gameState: The central game state containing all players, decks, and turn tracking
//   - Player: A player with lives, a hand of cards, and a farm to defend
//   - Farm: Contains defensive stacks and pending night cards (zombie attacks)
//   - Stack/Stacks: Collections of FarmItemType cards that form defenses
//   - NightCard: Either a zombie attack or an event that affects all players
//
// Game flow:
//  1. Create a new game with CreateNewGame(playerNames...)
//  2. Call ContinueDay() to advance the game state
//  3. When PlayerInputNeeded is returned, gather input and call ContinueAfterInput(choice)
//  4. Repeat until the game ends (all players eliminated)
package zcgame

import (
	"fmt"
	"math/rand"
)

// shuffle shuffles all elements of a slice in-place and returns the slice.
// Uses the Fisher-Yates algorithm via rand.Shuffle.
func shuffle[T any](slice []T) []T {
	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
	return slice
}

// PlayerPlayChoices configures automatic card placement behavior during day turns.
// When enabled, these options reduce the number of prompts by automatically
// choosing optimal placements for certain card types.
type PlayerPlayChoices struct {
	// AutoloadShotgun controls automatic Shotgun/Ammo pairing:
	//   - true: Ammo is placed on the shotgun with least ammo (ties go to first index).
	//           Shotgun is placed on the ammo stack with most ammo.
	//   - false: Ammo is placed on an unloaded shotgun only if exactly one exists,
	//            otherwise prompts for player input.
	AutoloadShotgun bool

	// AutoBuildHayWall controls automatic Hay Wall construction:
	//   - true: HayBale is placed on the incomplete wall with most HayBales (but <3).
	//   - false: HayBale is placed on an incomplete wall only if exactly one exists,
	//            otherwise prompts for player input.
	AutoBuildHayWall bool
}

// Player represents a participant in the game with their current state.
// Each player has lives (health), a hand of cards to play, and a farm to defend.
type Player struct {
	Name        string            // Display name (may include ANSI codes in CLI mode)
	Lives       int               // Remaining lives; eliminated when <= 0
	Farm        *Farm             // The player's farm with defensive stacks
	Hand        Hand              // Current hand of 5 cards
	PlayChoices PlayerPlayChoices // Automatic placement preferences
}

// Turn represents the current phase of a game day.
// A full day consists of Morning, Afternoon, and Night phases.
type Turn uint8

const (
	Morning   Turn = iota // First day phase: all players take turns
	Afternoon             // Second day phase: all players take turns again
	Night                 // Night phase: zombie attacks are resolved
	Day                   // Reserved for future co-op mode
)

// StageInTurn represents the current stage within a player's day turn.
// Each player progresses through these stages during Morning and Afternoon.
type StageInTurn uint8

const (
	OptionalDiscard StageInTurn = iota // Player may discard one card to draw a replacement
	Play2Cards                         // Player must play 2 cards from hand to farm
	Draw2Cards                         // Player draws 2 cards (public or from deck)
	Nighttime                          // Night phase processing (not a player turn stage)
)

// Players is a slice of Player pointers representing all active players in the game.
type Players []*Player

// PublicDayCards represents the two face-up cards available for drawing.
// Players can choose to draw these instead of drawing from the deck.
type PublicDayCards [2]FarmItemType

func (p PublicDayCards) String() string {
	return fmt.Sprintf("%s", Stack(p[:]))
}

// DaySubStage tracks the current sub-stage within a player's day turn.
// This enables the state machine to pause for player input and resume at the
// correct point. The game saves this value when returning PlayerInputNeeded
// and uses it to continue execution after input is provided.
type DaySubStage uint8

const (
	DaySubStageOptionalDiscard DaySubStage = iota // Waiting for optional discard decision
	DaySubStagePlay1                              // Waiting for first card selection
	DaySubStagePlay1Stack                         // Waiting for stack selection for first card
	DaySubStagePlay2                              // Waiting for second card selection
	DaySubStagePlay2Stack                         // Waiting for stack selection for second card
	DaySubStageDraw                               // Waiting for draw choice (public vs deck)
)

// NightSubStage tracks the current sub-stage within the night phase.
// Similar to DaySubStage, this enables interruptible execution during
// zombie attacks and event processing.
type NightSubStage uint8

const (
	NightSubStageProcessCards     NightSubStage = iota // Processing night cards normally
	NightSubStageZombieAutoKilled                      // Zombie was auto-killed; awaiting confirmation
	NightSubStageNoDefense                             // No defense available; awaiting life loss confirmation
	NightSubStageEliminated                            // Player eliminated; awaiting confirmation
	NightSubStageChooseDefense                         // Player must choose which stack to use
	NightSubStageChooseShield                          // Player must decide whether to use Shield
	NightSubStageConfirmLifeLoss                       // Player chose to lose life; awaiting confirmation
	NightSubStageEventConfirm                          // Event triggered; awaiting confirmation
	NightSubStageEventDiscard                          // Event requires discards; awaiting selection
)

// gameState is the central data structure containing all game state.
// It tracks players, decks, turn progression, and state machine position.
//
// The state machine fields (DaySubStage, NightSubStage, etc.) enable
// interruptible execution. When player input is needed, the current
// position is saved so execution can resume after input is provided.
//
// This type is unexported. Use CreateNewGame to get a GameView, which
// provides controlled access to the game state.
type gameState struct {
	// Core game state
	Players             Players              // Active players (eliminated players are removed)
	CurrentPlayerIdx    int                  // Index of the player whose turn it is
	StageInTurn         StageInTurn          // Current stage within a player's turn
	Turn                Turn                 // Current phase of the day (Morning/Afternoon/Night)
	DayDeck             Stack                // Draw pile for day cards
	PublicDayCards      PublicDayCards       // Two face-up cards available for drawing
	DiscardedDayCards   map[FarmItemType]int // Discard pile counts by card type
	NightDeck           NightCards           // Draw pile for night cards (zombies and events)
	DiscardedNightCards NightCards           // Discarded night cards
	NightNum            int                  // Current night number (increases each night)

	// Day turn state machine fields
	DaySubStage        DaySubStage   // Current sub-stage within day turn
	NightSubStage      NightSubStage // Current sub-stage within night phase
	PendingCardItem    FarmItemType  // Card awaiting stack selection
	PendingStackChoice int           // Stack index chosen for card placement
	PlayerTurnIndex    int           // Which player's turn in current phase (0 to len(Players)-1)

	// Night phase state machine fields
	NightCardsDealt       bool           // Whether night cards have been dealt this night
	NightPlayerIndex      int            // Current player index in night round
	NightPlayersToProcess int            // Total players to process in current round
	NightAnyCardProcessed bool           // Whether any card was processed this round
	CurrentNightCard      *NightCard     // Night card currently being resolved
	CurrentZombie         *ZombieChicken // Zombie currently attacking
	ChosenStackIdx        int            // Stack index chosen for defense

	// Event discard state (for Lightning Storm, Tornado)
	EventDiscardStartIdx  int // Player index when event was triggered
	EventDiscardPlayerIdx int // Current player offset from start (0 to len(Players)-1)
	EventDiscardRemaining int // Cards remaining to discard for current player
	EventDiscardTotal     int // Total cards each player must discard

	// Pending event display state
	PendingEventName string // Event name saved for confirmation display
	PendingEventDesc string // Event description saved for confirmation display
}

// ZombieTrait represents a special ability that a zombie chicken can have.
// Traits determine which defenses are effective against a zombie.
type ZombieTrait uint8

const (
	Invisible         ZombieTrait = iota // Overcomes: Shotgun, Flamethrower (cannot be seen to target)
	Flying                               // Overcomes: Hay Wall, Booby Trap (flies over ground defenses)
	Climbing                             // Overcomes: Hay Wall (climbs over walls)
	Bulletproof                          // Overcomes: Shotgun (bullets have no effect)
	Fireproof                            // Overcomes: Flamethrower (immune to fire)
	Timid                                // Weakness: Frightened by Scarecrows
	Exploding                            // Destroys the stack used to defeat it (unless Shield is used)
	NUM_ZOMBIE_TRAITS                    // Sentinel value for bounds checking
)

// ZombieTraits is a slice of traits belonging to a zombie chicken.
type ZombieTraits []ZombieTrait

// ZombieChicken represents a type of zombie that can attack players.
// Each zombie type has a unique combination of traits that determine
// which defenses are effective against it.
type ZombieChicken struct {
	Traits    ZombieTraits // Special abilities this zombie has
	NumInDeck int8         // How many of this zombie type exist in the night deck
	Name      string       // Display name for this zombie type
}

// Event represents a special night card that affects all players.
// Unlike zombie cards, events trigger global effects such as forcing
// discards or adding extra night cards.
type Event struct {
	Action      func(*gameState) *PlayerInputNeeded // Effect to apply; may return input request
	Name        string                              // Display name for the event
	Description string                              // Description of the event's effect
}

// NightCardEvents contains all event cards that can appear in the night deck.
// Events affect all players simultaneously and are processed when drawn.
// Lightning Storm and Tornado are listed first for debug testing convenience.
var NightCardEvents = []Event{
	{
		Name:        "Lightning Storm",
		Description: "All players discards 2 cards from their farm.",
		Action: func(g *gameState) *PlayerInputNeeded {
			return g.startEventDiscard(2)
		},
	},
	{
		Name:        "Tornado",
		Description: "All players discard 3 cards from their farm.",
		Action: func(g *gameState) *PlayerInputNeeded {
			return g.startEventDiscard(3)
		},
	},
	{
		Name:        "Blood Moon",
		Description: "Zombies are flocking tonight!\nAll players draw 3 more Night cards.",
		Action: func(g *gameState) *PlayerInputNeeded {
			for i := range g.Players {
				idx := (g.CurrentPlayerIdx + i) % len(g.Players)
				for range 3 {
					g.Players[idx].Farm.NightCards = append(g.Players[idx].Farm.NightCards, g.nextNightCard())
				}
			}
			return nil
		},
	},
	{
		Name:        "Winter Solstice",
		Description: "It's gonna be a long night! All players draw 2 more Night cards.",
		Action: func(g *gameState) *PlayerInputNeeded {
			for i := range g.Players {
				idx := (g.CurrentPlayerIdx + i) % len(g.Players)
				g.Players[idx].Farm.NightCards = append(g.Players[idx].Farm.NightCards, g.nextNightCard())
				g.Players[idx].Farm.NightCards = append(g.Players[idx].Farm.NightCards, g.nextNightCard())
			}
			return nil
		},
	},
	{
		Name:        "Squirrel Stampede",
		Description: "A squirrel stampede triggers all Booby Traps! All players discard any Booby Traps on their farm.",
		Action: func(g *gameState) *PlayerInputNeeded {
			for i := range g.Players {
				idx := (g.CurrentPlayerIdx + i) % len(g.Players)
				if g.Players[idx].Farm.HasItemInStacks(BoobyTrap) {
					for j := range g.Players[idx].Farm.Stacks {
						if g.Players[idx].Farm.Stacks[j].HasItem(BoobyTrap) {
							g.Players[idx].Farm.Stacks[j].RemoveItem(BoobyTrap)
							g.discardDayCard(BoobyTrap)
						}
					}
				}
				g.Players[idx].Farm.clearStacks()
			}
			return nil
		},
	},
	{
		Name:        "Heavy Rainfall",
		Description: "Water rusts Flamethrowers! All players discard any Flamethrowers and Fuel on their farm.",
		Action: func(g *gameState) *PlayerInputNeeded {
			for i := range g.Players {
				idx := (g.CurrentPlayerIdx + i) % len(g.Players)
				for j := range g.Players[idx].Farm.Stacks {
					if g.Players[idx].Farm.Stacks[j].HasItem(Flamethrower) {
						g.Players[idx].Farm.Stacks[j].RemoveItem(Flamethrower)
						g.discardDayCard(Flamethrower)
					}
					if g.Players[idx].Farm.Stacks[j].HasItem(Fuel) {
						g.Players[idx].Farm.Stacks[j].RemoveItem(Fuel)
						g.discardDayCard(Fuel)
					}
				}
				g.Players[idx].Farm.clearStacks()
			}
			return nil
		},
	},
	{
		Name:        "Silent Night",
		Description: "No more zombies tonight! All players discard any remaining Night cards.",
		Action: func(g *gameState) *PlayerInputNeeded {
			for i := range g.Players {
				idx := (g.CurrentPlayerIdx + i) % len(g.Players)
				for _, card := range g.Players[idx].Farm.NightCards {
					g.discardNightCard(card)
				}
				g.Players[idx].Farm.NightCards = g.Players[idx].Farm.NightCards[:0] // clear
			}
			return nil
		},
	},
}

// NightCard represents a card drawn during the night phase.
// Each night card is either a zombie attack or an event.
// Players receive NightNum cards each night, which are resolved one at a time.
//
// For zombie cards, ZombieKey indexes into ZombieChickens map.
// For event cards, ZombieKey is -1 and Event contains the effect.
type NightCard struct {
	Event     Event // Event data (only valid when ZombieKey == -1)
	ZombieKey int   // Index into ZombieChickens map, or -1 for events
}

// NightCards is a slice of NightCard, used for decks and discard piles.
type NightCards []NightCard

// IsZombie returns true if this night card contains a zombie (not an event).
// Both IsZombie and IsEvent exist for semantic clarity in different contexts -
// use whichever reads more naturally in your code.
func (n NightCard) IsZombie() bool {
	return n.ZombieKey != -1
}

// IsEvent returns true if this night card contains an event (not a zombie).
// Both IsZombie and IsEvent exist for semantic clarity in different contexts -
// use whichever reads more naturally in your code.
func (n NightCard) IsEvent() bool {
	return n.ZombieKey == -1
}

// FarmItemType represents a type of card that can be played to a player's farm.
// Cards are combined into stacks to form defenses against zombie attacks.
type FarmItemType uint16

const (
	HayBale        FarmItemType = iota // 20 in deck | Stack 3 to build a Hay Wall
	Scarecrow                          //  6 in deck | Scares away Timid zombies
	Shotgun                            // 14 in deck | Combines with Ammo to blast a zombie
	Ammo                               // 24 in deck | Combine with Shotgun (one-time-use)
	BoobyTrap                          // 10 in deck | Terminates 1 zombie (one-time-use)
	Shield                             //  6 in deck | Protects stack from Exploding zombie (one-time-use)
	Flamethrower                       //  6 in deck | Combine with Fuel to roast a zombie
	Fuel                               //  6 in deck | Combine with Flamethrower to roast a zombie
	WOLR                               //  4 in deck | Destroys 1 zombie plus entire farm (one-time-use)
	NUM_FARM_ITEMS                     // Sentinel value for bounds checking and empty hand slots
)

// IsOneTimeUse returns true if this item type is consumed when used.
// One-time-use items are discarded after defeating a zombie.
func (f FarmItemType) IsOneTimeUse() bool {
	switch f {
	case Ammo, WOLR, BoobyTrap, Shield:
		return true
	default:
		return false
	}
}

// CanBeStackedWithLookup defines which item types can be combined in a stack.
// An empty slice means the item must always start a new stack.
var CanBeStackedWithLookup = map[FarmItemType]Stack{
	HayBale:      {HayBale},       // Hay Bales stack together (max 3 for a wall)
	Scarecrow:    {},              // Must be alone
	Shotgun:      {Ammo},          // Can be placed on Ammo stacks
	Ammo:         {Shotgun, Ammo}, // Can be added to Shotgun or other Ammo
	BoobyTrap:    {},              // Must be alone
	Shield:       {},              // Must be alone
	Flamethrower: {Fuel},          // Can be placed on Fuel
	Fuel:         {Flamethrower},  // Can be placed on Flamethrower
	WOLR:         {},              // Must be alone
}

// DayCardAmounts defines how many of each card type exist in the day deck.
var DayCardAmounts = map[FarmItemType]int{
	HayBale:      20,
	Scarecrow:    6,
	Shotgun:      14,
	Ammo:         24,
	BoobyTrap:    10,
	Shield:       6,
	Flamethrower: 6,
	Fuel:         6,
	WOLR:         4,
}

// Stack is a collection of FarmItemType cards that form a single defense.
// Valid stacks follow specific rules (e.g., Shotgun+Ammo, 3 HayBales for a wall).
// Use Farm.PlayCard to add cards to stacks with automatic rule enforcement.
type Stack []FarmItemType

// Stacks is a collection of Stack, representing all defenses on a player's farm.
type Stacks []Stack

// Sort sorts items within each stack by FarmItemType, removes empty stacks in-place,
// then sorts the stacks themselves by their first element.
//
// This method modifies the Stacks slice in-place. After sorting:
//   - Each stack's items are ordered by FarmItemType value (ascending)
//   - Empty stacks are removed from the slice
//   - Stacks are ordered by their first element's FarmItemType value (ascending)
//
// Example:
//
//	stacks := Stacks{{Ammo, Shotgun}, {}, {HayBale, HayBale}}
//	stacks.Sort()
//	// Result: {{HayBale, HayBale}, {Shotgun, Ammo}}
func (s *Stacks) Sort() {
	// Sort items within each stack
	for i := range *s {
		(*s)[i].Sort()
	}

	// Remove empty stacks in-place
	writeIdx := 0
	for i := range *s {
		if len((*s)[i]) > 0 {
			(*s)[writeIdx] = (*s)[i]
			writeIdx++
		}
	}
	*s = (*s)[:writeIdx]

	// Sort stacks by first element (bubble sort)
	for i := 0; i < len(*s); i++ {
		for j := i + 1; j < len(*s); j++ {
			if (*s)[i][0] > (*s)[j][0] {
				(*s)[i], (*s)[j] = (*s)[j], (*s)[i]
			}
		}
	}
}

// Sort sorts the stack's items in-place by FarmItemType value (ascending).
// Uses bubble sort for simplicity since stacks are typically small (1-3 items).
//
// Example:
//
//	stack := Stack{Ammo, Shotgun, Ammo}
//	stack.Sort()
//	// Result: {Shotgun, Ammo, Ammo}
func (s Stack) Sort() {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// Sort sorts the hand in-place, moving empty slots to the end and ordering
// remaining cards by FarmItemType value (ascending).
//
// Empty slots (FarmItemType == NUM_FARM_ITEMS) are always placed at the end.
// This ensures that hand indices 0-2 (or 0-4 when full) contain valid cards,
// which is important for the game loop that accesses slots [3] and [4] after drawing.
//
// Example:
//
//	hand := Hand{{Ammo, false}, {NUM_FARM_ITEMS, false}, {HayBale, false}, ...}
//	hand.Sort()
//	// Result: {{HayBale, false}, {Ammo, false}, {NUM_FARM_ITEMS, false}, ...}
func (h *Hand) Sort() {
	// Move NUM_FARM_ITEMS (blank slots) to the end, then sort by FarmItemType
	for i := 0; i < len(h); i++ {
		for j := i + 1; j < len(h); j++ {
			// Blank slots go to end
			if h[i].FarmItemType == NUM_FARM_ITEMS && h[j].FarmItemType != NUM_FARM_ITEMS {
				h[i], h[j] = h[j], h[i]
			} else if h[i].FarmItemType != NUM_FARM_ITEMS && h[j].FarmItemType != NUM_FARM_ITEMS {
				// Both non-blank: sort by FarmItemType
				if h[i].FarmItemType > h[j].FarmItemType {
					h[i], h[j] = h[j], h[i]
				}
			}
		}
	}
}

// RemoveItem removes the first occurrence of item from the stack.
// If the item is not found, the stack is unchanged.
func (s *Stack) RemoveItem(item FarmItemType) {
	for i, card := range *s {
		if card == item {
			*s = append((*s)[:i], (*s)[i+1:]...)
			return
		}
	}
}

// HandItem represents a single card in a player's hand.
type HandItem struct {
	FarmItemType FarmItemType // The card type, or NUM_FARM_ITEMS for empty slot
	Visible      bool         // Whether this card is visible to other players (for frontend)
}

// Hand is a fixed-size array of 5 HandItem slots representing a player's hand.
// Empty slots have FarmItemType set to NUM_FARM_ITEMS.
type Hand [5]HandItem

// Farm represents a player's defensive area containing stacks of cards
// and pending night cards (zombie attacks) to resolve.
type Farm struct {
	Stacks     Stacks     // Defensive stacks built from day cards
	NightCards NightCards // Pending night cards to resolve this night
}

// RemoveItemByFlatIndex removes an item by its flat index across all stacks.
// Returns the removed item type, or NUM_FARM_ITEMS if index is out of bounds.
// This is called internally by the state machine with pre-validated indices from event discards.
func (f *Farm) RemoveItemByFlatIndex(flatIdx int, g *gameState) FarmItemType {
	idx := 0
	for i := range f.Stacks {
		for j := range f.Stacks[i] {
			if idx == flatIdx {
				item := f.Stacks[i][j]
				f.Stacks[i] = append(f.Stacks[i][:j], f.Stacks[i][j+1:]...)
				f.clearStacks()
				g.discardDayCard(item)
				return item
			}
			idx++
		}
	}
	return NUM_FARM_ITEMS
}
