package zcgame

import (
	"fmt"
	"math/rand"
)

// shuffle shuffles all elements of a slice in-place and also returns the slice.
func shuffle[T any](slice []T) []T {
	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
	return slice
}

type PlayerPlayChoices struct {
	// AutoloadShotgun:
	// - true: Puts Ammo on shotgun with least ammo (ties go to first index)
	//         Also places Shotgun on Ammo stack with most ammo
	// - false: Puts Ammo on unloaded shotgun if exactly 1, else ErrNeedsPlayerInput
	AutoloadShotgun bool

	// AutoBuildHayWall:
	// - true: Puts HayBale on incomplete wall with most HayBales (but <3)
	// - false: Puts HayBale on incomplete wall if exactly 1, else ErrNeedsPlayerInput
	AutoBuildHayWall bool
}

type Player struct {
	Name        string
	Lives       int
	Farm        *Farm
	Hand        Hand
	PlayChoices PlayerPlayChoices
}

type Turn uint8

const (
	Morning Turn = iota
	Afternoon
	Night
	Day // TODO: implement co-op mode
)

type StageInTurn uint8

const (
	OptionalDiscard StageInTurn = iota
	Play2Cards
	Draw2Cards
	Nighttime
)

type Players []*Player

type PublicDayCards [2]FarmItemType

func (p PublicDayCards) String() string {
	return fmt.Sprintf("%s", Stack(p[:]))
}

// DaySubStage tracks sub-stages within the day turn for resumption
type DaySubStage uint8

const (
	DaySubStageOptionalDiscard DaySubStage = iota
	DaySubStagePlay1
	DaySubStagePlay1Stack // Waiting for stack selection after PlayCard returns NeedsPlayerInput
	DaySubStagePlay2
	DaySubStagePlay2Stack // Waiting for stack selection after PlayCard returns NeedsPlayerInput
	DaySubStageDraw
)

// NightSubStage tracks sub-stages within the night turn for resumption
type NightSubStage uint8

const (
	NightSubStageProcessCards     NightSubStage = iota
	NightSubStageZombieAutoKilled               // Waiting for confirmation after auto-kill
	NightSubStageNoDefense                      // Waiting for confirmation when no defense
	NightSubStageEliminated                     // Waiting for confirmation after elimination
	NightSubStageChooseDefense                  // Waiting for player to choose defense stack
	NightSubStageChooseShield                   // Waiting for shield decision
	NightSubStageConfirmLifeLoss                // Waiting for confirmation after choosing -1
	NightSubStageEventConfirm                   // Waiting for event confirmation
	NightSubStageEventDiscard                   // Waiting for event discard selection
)

type GameState struct {
	Players             Players
	CurrentPlayerIdx    int
	StageInTurn         StageInTurn
	Turn                Turn
	DayDeck             Stack
	PublicDayCards      PublicDayCards
	DiscardedDayCards   map[FarmItemType]int
	NightDeck           NightCards
	DiscardedNightCards NightCards
	NightNum            int

	// State tracking for resumption after player input
	DaySubStage        DaySubStage
	NightSubStage      NightSubStage
	PendingCardItem    FarmItemType // Card being played that needs stack selection
	PendingStackChoice int          // Stack index chosen for card placement
	PlayerTurnIndex    int          // Tracks which player's turn during day loop (0 to len(Players)-1)

	// Night state tracking
	NightCardsDealt       bool           // Whether night cards have been dealt this night
	NightPlayerIndex      int            // Index in current night round
	NightPlayersToProcess int            // Number of players to process in current round
	NightAnyCardProcessed bool           // Whether any card was processed this round
	CurrentNightCard      *NightCard     // Current night card being processed
	CurrentZombie         *ZombieChicken // Current zombie being fought
	ChosenStackIdx        int            // Stack chosen for defense

	// Event discard state
	EventDiscardStartIdx  int // Starting player index (current player when event triggered)
	EventDiscardPlayerIdx int // Which player is discarding (offset from start, 0 to len(Players)-1)
	EventDiscardRemaining int // How many more cards to discard for current player
	EventDiscardTotal     int // Total cards to discard per player (for display)

	// Pending event confirmation (saved before action runs so it's not overwritten)
	PendingEventName string
	PendingEventDesc string
}

type ZombieTrait uint8

const (
	Invisible   ZombieTrait = iota // OVERCOMES: Shotgun, Flamethrower
	Flying                         // OVERCOMES: Hay Wall, Booby Trap
	Climbing                       // OVERCOMES: Hay Wall
	Bulletproof                    // OVERCOMES: Shotgun
	Fireproof                      // OVERCOMES: Flamethrower
	Timid                          // Timid zombies are frightened by scarecrows
	Exploding                      // Exploding zombies destroy the stack used to defeat it
	NUM_ZOMBIE_TRAITS
)

type ZombieTraits []ZombieTrait

type ZombieChicken struct {
	Traits    ZombieTraits
	NumInDeck int8
	Name      string
}

type Event struct {
	Action            func(*GameState) *PlayerInputNeeded
	Name, Description string
}

var (
	// Each event needs to be applied to each Players' *Farm
	// Tornado and Lightning Storm are first for debug testing
	NightCardEvents = []Event{
		{
			Name:        "Lightning Storm",
			Description: "All players discards 2 cards from their farm.",
			Action: func(g *GameState) *PlayerInputNeeded {
				return g.startEventDiscard(2)
			},
		},
		{
			Name:        "Tornado",
			Description: "All players discard 3 cards from their farm.",
			Action: func(g *GameState) *PlayerInputNeeded {
				return g.startEventDiscard(3)
			},
		},
		{
			Name:        "Blood Moon",
			Description: "Zombies are flocking tonight!\nAll players draw 3 more Night cards.",
			Action: func(g *GameState) *PlayerInputNeeded {
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
			Action: func(g *GameState) *PlayerInputNeeded {
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
			Action: func(g *GameState) *PlayerInputNeeded {
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
			Action: func(g *GameState) *PlayerInputNeeded {
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
			Action: func(g *GameState) *PlayerInputNeeded {
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
)

// A night card is given to each player each Night for the number of nights + 1
// that have passed this game.
//
// A night card without an event has an Event that is zeroed out.
// ZombieKey is set to -1 for Events.
type NightCard struct {
	Event     Event
	ZombieKey int
}

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

type FarmItemType uint16

const (
	HayBale      FarmItemType = iota // 20 | Stack 3 Hay Bales to build a Hay Wall
	Scarecrow                        //  6 | Scares away Timid Zombies
	Shotgun                          // 14 | Combines with Ammo to blast a zombie
	Ammo                             // 24 | Combine with Shotgun to blast a zombie (1-Time-Use)
	BoobyTrap                        // 10 | Terminates 1 Zombie (1-Time-Use)
	Shield                           //  6 | Shields a stack from an Exploding Zombie (1-Time-Use)
	Flamethrower                     //  6 | Combine with Fuel to roast a zombie
	Fuel                             //  6 | Combine with Flamethrower to roast a zombie
	WOLR                             //  4 | Destroys and 1 zombie plus everything else on your farm (1-Time-Use)
	NUM_FARM_ITEMS
)

func (f FarmItemType) IsOneTimeUse() bool {
	switch f {
	case Ammo, WOLR, BoobyTrap, Shield:
		return true
	default:
		return false
	}
}

var (
	// legal farm items key can be stacked with, if blank must be a new stack
	CanBeStackedWithLookup = map[FarmItemType]Stack{
		HayBale:      {HayBale},
		Scarecrow:    {},
		Shotgun:      {Ammo}, // explicitly blank, ammo is used bc of one time use
		Ammo:         {Shotgun, Ammo},
		BoobyTrap:    {},
		Shield:       {},
		Flamethrower: {Fuel},
		Fuel:         {Flamethrower},
		WOLR:         {},
	}

	DayCardAmounts = map[FarmItemType]int{
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
)

type Stack []FarmItemType
type Stacks []Stack

// Sort sorts items within each stack, removes empty stacks in-place,
// then sorts stacks by their first element.
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

// Sort sorts the stack items in-place by FarmItemType.
func (s Stack) Sort() {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// Sort sorts the hand in-place by visible (true first), then by FarmItemType.
// Called explicitly in game.go before accessing [3] and [4] slots.
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

func (s *Stack) RemoveItem(item FarmItemType) {
	for i, card := range *s {
		if card == item {
			*s = append((*s)[:i], (*s)[i+1:]...)
			return
		}
	}
}

type HandItem struct {
	FarmItemType FarmItemType
	Visible      bool // returns true if visible to other players. used for front end
}

type Hand [5]HandItem

type Farm struct {
	Stacks     Stacks
	NightCards NightCards // current night cards being attacked from
}

// RemoveItemByFlatIndex removes an item by its flat index across all stacks.
// Returns the removed item type, or NUM_FARM_ITEMS if index is out of bounds.
// This is called internally by the state machine with pre-validated indices from event discards.
func (f *Farm) RemoveItemByFlatIndex(flatIdx int, g *GameState) FarmItemType {
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
