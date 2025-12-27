package zombiechickens

import "math/rand"

// Shuffle shuffles all elements of a slice in-place and also returns the slice.
func Shuffle[T any](slice []T) []T {
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
	Day // used for co-op
)

type Players []*Player

type GameState struct {
	Players             Players
	Turn                Turn
	DayDeck             Stack
	DiscardedDayCards   map[FarmItemType]int
	NightDeck           []NightCard
	DiscardedNightCards []NightCard
	NightNum            int
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

type ZombieTraits []ZombieTrait

type ZombieChicken struct {
	Traits    ZombieTraits
	NumInDeck int8
	Name      string
}

type Event struct {
	Action            func(*GameState)
	Name, Description string
}

var (
	// Each event needs to be applied to each Players' *Farm
	NightCardEvents = []Event{
		{
			Name: "Tornado",
			Action: func(g *GameState) {
				// All players discards 3 cards from their farm
			},
		},
		{
			Name: "Lightning Storm",
			Action: func(g *GameState) {
				// All players discards 2 cards from their farm
			},
		},
		{
			Name:        "Blood Moon",
			Description: "Zombies are flocking tonight!\nAll players draw 3 more Night cards.",
			Action: func(g *GameState) {
				for _, player := range g.Players {
					player.Farm.NightCards = append(player.Farm.NightCards, g.NextNightCard())
					player.Farm.NightCards = append(player.Farm.NightCards, g.NextNightCard())
					player.Farm.NightCards = append(player.Farm.NightCards, g.NextNightCard())
				}
			},
		},
		{
			Name:        "Winter Solstice",
			Description: "It's gonna be a long night! All players draw 2 more Night cards.",
			Action: func(g *GameState) {
				for _, player := range g.Players {
					player.Farm.NightCards = append(player.Farm.NightCards, g.NextNightCard())
					player.Farm.NightCards = append(player.Farm.NightCards, g.NextNightCard())
				}
			},
		},
		{
			Name:        "Squirrel Stampede",
			Description: "A squirrel stampede triggers all Booby Traps! All players discard any Booby Traps on their farm.",
			Action: func(g *GameState) {
				for i := range g.Players {
					if g.Players[i].Farm.HasItemInStacks(BoobyTrap) {
						// could have better/more performant logic here. need to benchmark
						for j := range g.Players[i].Farm.Stacks {
							g.Players[i].Farm.Stacks[j].RemoveItem(BoobyTrap)
						}
					}

					g.Players[i].Farm.clearStacks()
				}
			},
		},
		{
			Name:        "Heavy Rainfall",
			Description: "Water rusts Flamethrowers! All players discard any Flamethrowers and Fuel on their farm.",
			Action: func(g *GameState) {
				for i := range g.Players {
					if g.Players[i].Farm.HasItemInStacks(Fuel) {
						// could have better/more performant logic here. need to benchmark
						for j := range g.Players[i].Farm.Stacks {
							g.Players[i].Farm.Stacks[j].RemoveItem(BoobyTrap)
						}
					}

					g.Players[i].Farm.clearStacks()
				}
			},
		},
		{
			Name:        "Silent Night",
			Description: "No more zombies tonight! All players discard any remaining Night cards.",
			Action: func(g *GameState) {
				for _, player := range g.Players {
					for _, card := range player.Farm.NightCards {
						g.DiscardNightCard(card)
					}
					player.Farm.NightCards = player.Farm.NightCards[:0] // clear

				}
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

// redundant with IsEvent
func (n NightCard) IsZombie() bool {
	return n.ZombieKey != -1
}

// redundant with IsZombie
func (n NightCard) IsEvent() bool {
	return n.ZombieKey == -1
}

type FarmItemType uint16

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

func (f FarmItemType) IsOneTimeUse() bool {
	switch f {
	case Ammo, WOLR, BoobyTrap, Shield:
		return true
	default:
		return false
	}
}

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
type Stacks []Stack

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
