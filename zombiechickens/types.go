package zombiechickens

import (
	"errors"
	"fmt"
)

type Player struct {
	Farm        *Farm
	CardsInHand [5]FarmItemType
}

type GameState struct {
	Players             []*Player
	DayDeck             Stack
	DiscardedDayCards   map[FarmItemType]int
	NightDeck           []NightCard
	DiscardedNightCards []NightCard
}

type ZombieTrait uint8

const (
	Invisible   ZombieTrait = iota // OVERCOMES: Shotgun, Flamethrower
	Flying                         // OVERCOMES: Hay Wall, Booby Trap
	Climbing                       // OVERCOMES: Hay Wall
	Bulletproof                    // OVERCOMES: Shotgun
	Fireproof                      // OVERCOMES: Flamethrower
	Timid                          // Timid zombies are frightened by scarecrows
	Exploding                      // Expldoing zombies destroy the stack used to defeat it
	NUM_ZOMBIE_TRAITS
)

type ZombieChicken struct {
	Traits    []ZombieTrait
	NumInDeck int8
	Name      string
}

var (
	ZombieChickens = map[int]ZombieChicken{
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
)

type Event struct {
	Action func(*Farm, *GameState)
	Name   string
}

var (
	// Each event needs to be applied to each Players' *Farm
	NightCardEvents = []Event{
		{
			Name: "Tornado",
			Action: func(f *Farm, g *GameState) {
				// All players discards 3 cards from their farm
			},
		},
		{
			Name: "Lightning Storm",
			Action: func(f *Farm, g *GameState) {
				// All players discards 2 cards from their farm
			},
		},
		{
			Name: "Blood Moon",
			Action: func(f *Farm, g *GameState) {
				// Zombies are flocking tonight!
				// All players draw 3 more Night cards

				f.NightCards = append(f.NightCards, g.NextNightCard())
				f.NightCards = append(f.NightCards, g.NextNightCard())
				f.NightCards = append(f.NightCards, g.NextNightCard())
			},
		},
		{
			Name: "Winter Solstice",
			Action: func(f *Farm, g *GameState) {
				// It's gonna be a long night!
				// All players draw 2 more Night cards
				f.NightCards = append(f.NightCards, g.NextNightCard())
				f.NightCards = append(f.NightCards, g.NextNightCard())
			},
		},
		{
			Name: "Squirrel Stampede",
			Action: func(f *Farm, g *GameState) {
				// A squirrel stampede triggers all
				// Booby Traps! All players discard
				// any Booby Traps on their farm.
			},
		},
		{
			Name: "Heavy Rainfall",
			Action: func(f *Farm, g *GameState) {
				// Water rusts Flamethrowers! All
				// players discard any Flamethrowers
				// and Fuel on their farm.
			},
		},
		{
			Name: "Silent Night",
			Action: func(f *Farm, g *GameState) {
				// No more zombies tonight!
				// All players discard any
				// remaining Night cards.

				for _, card := range f.NightCards {
					g.DiscardNightCard(card)
				}

				f.NightCards = f.NightCards[:0]
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

func CreateNightCardDeck() []NightCard {
	var (
		deck = make([]NightCard, 0)
	)

	for zIndex, chickenType := range ZombieChickens {
		for range chickenType.NumInDeck {
			deck = append(deck, NightCard{
				ZombieKey: zIndex,
			})
		}
	}

	for _, event := range NightCardEvents {
		deck = append(deck, NightCard{
			Event:     event,
			ZombieKey: -1,
		})
	}

	return deck
}

// redundant with IsEvent
func (n NightCard) IsZombie() bool {
	return n.ZombieKey != -1
}

// redundant with IsZombie
func (n NightCard) IsEvent() bool {
	return n.ZombieKey == -1
}

type FarmItemType uint16

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
		return ""
	}
}

func (f FarmItemType) IsOneTimeUse() bool {
	switch f {
	case Ammo, WOLR, BoobyTrap, Shield:
		return true
	default:
		return false
	}
}

const (
	HayBale      FarmItemType = iota // 20 | Stack 3 Hay Bales to build a Hay Wall
	Scarecrow                        //  6 | Scares away Timid Zombies
	Shotgun                          // 14 | Combies with Ammo to blast a zombie
	Ammo                             // 24 | Combine with Shotgun to blast a zombie (1-Time-Use)
	BoobyTrap                        // 10 | Terminates 1 Zombie (1-Time-Use)
	Shield                           //  6 | Shields a stack from an Exploding Zombie (1-Time-Use)
	Flamethrower                     //  6 | Combine with Fuel to roast a zombie
	Fuel                             //  6 | Combine with Flamethrower to roast a zombie
	WOLR                             //  4 | Destroys and 1 zombie plus everything else on your farm (1-Time-Use)
	NUM_FARM_ITEMS
)

var (
	// needed to check if every farm item in stack exists uniquely (3 unique hay bales, 1 shotgun + any # ammo...)
	StackNeededForLiveLookup = map[FarmItemType]Stack{
		HayBale:      {HayBale, HayBale, HayBale},
		Scarecrow:    {Scarecrow},
		Shotgun:      {}, // explicitly blank, ammo is used bc of one time use
		Ammo:         {Shotgun, Ammo},
		BoobyTrap:    {BoobyTrap},
		Shield:       {Shield},
		Flamethrower: {Flamethrower, Fuel},
		Fuel:         {}, // flamethrower is used
		WOLR:         {WOLR},
	}
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
	FarmItemTraitWeakness = map[FarmItemType]ZombieTrait{
		HayBale:      Flying,
		Scarecrow:    NUM_ZOMBIE_TRAITS, //TODO:FIXME: special case??
		Shotgun:      Bulletproof,       //technically unneeded
		Ammo:         Bulletproof,
		BoobyTrap:    Flying,
		Shield:       NUM_ZOMBIE_TRAITS, //TODO:FIXME: special case??
		Flamethrower: Fireproof,
		Fuel:         Fireproof,         //same as shotgun
		WOLR:         NUM_ZOMBIE_TRAITS, //TODO:FIXME: special case
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

type Farm struct {
	Stacks     []Stack
	NightCards []NightCard // current night cards being attacked with
}

func (f *Farm) PlayCard(item FarmItemType) {
	if f.Stacks == nil {
		f.Stacks = make([]Stack, 0)
	}

	err := f.AddToStacks(item)
	if err != nil {
		if errors.Is(err, ErrNeedsPlayerInput) {
			// handle player input, add to input stack?
		} else {
			//TODO:FIXME:
			panic(err)
		}
	}

	err = f.AddToStacks(item)
}

func (f *Farm) makeStackWith(item FarmItemType) {
	var stack = make(Stack, 0, 1)
	stack = append(stack, item)
	f.Stacks = append(f.Stacks, stack)
}

var ErrNeedsPlayerInput = errors.New("needs player input")

//("needs player input")

func (f *Farm) addToStackIndex(item FarmItemType, stackIndex int) error {
	if stackIndex >= len(f.Stacks) {
		return fmt.Errorf("%d out of bounds, length of farm stacks is %d", stackIndex, f.Stacks)
	} else if stackIndex < 0 {
		return fmt.Errorf("%d must be a postive number", stackIndex)
	}
	return nil
}

func (f *Farm) HasItemInStacks(item FarmItemType) bool {
	for _, stack := range f.Stacks {
		for _, card := range stack {
			if card == item {
				return true
			}
		}
	}
	return false
}

func (f *Farm) AddToStacks(item FarmItemType) error {
	if len(f.Stacks) > 0 {
		f.makeStackWith(item)
		return nil
	}

	// hasItem := f.HasItemInStacks(item)

	// f.Stacks[stackIndex] = append(f.Stacks[stackIndex], item)

	return nil
}
