package zcgame

import "fmt"

// random order based on DayCards global variable configs
func createDayDeck() Stack {
	var deck = Stack{}
	for farmItem, amount := range DayCardAmounts {
		for range amount {
			deck = append(deck, farmItem)
		}
	}

	shuffle(deck)

	return deck
}

func createNightDeck() []NightCard {
	var deck = make([]NightCard, 0)
	for zKey, zombie := range ZombieChickens {
		for range zombie.NumInDeck {
			deck = append(deck, NightCard{
				ZombieKey: zKey,
			})
		}
	}

	for _, event := range NightCardEvents {
		deck = append(deck, NightCard{
			Event:     event,
			ZombieKey: -1,
		})
	}

	shuffle(deck)

	return deck
}

var (
	StartingLivesLookup = map[int]int{
		1: 5,
		2: 5,
		3: 4,
		4: 4,
	}
)

func (g *GameState) dealPublicDayCards() {
	g.PublicDayCards = [2]FarmItemType{g.nextDayCard(), g.nextDayCard()}
}

func CreateNewGame(playerNames ...string) (*GameState, error) {
	if len(playerNames) == 0 {
		return nil, fmt.Errorf("must provide at least 1 player")
	} else if len(playerNames) > 4 {
		return nil, fmt.Errorf("must provide max 4 player names")
	}

	var g = &GameState{
		DayDeck:             createDayDeck(),
		NightDeck:           createNightDeck(),
		Turn:                Morning,
		StageInTurn:         OptionalDiscard,
		CurrentPlayerIdx:    0, //redundant
		NightNum:            1,
		DiscardedDayCards:   make(map[FarmItemType]int),
		DiscardedNightCards: NightCards{},
	}

	g.dealPublicDayCards()
	g.Players = make([]*Player, 0, len(playerNames))
	for i, name := range playerNames {
		g.Players = append(g.Players, createPlayer(g, name, len(playerNames), i))
	}
	if err := g.assertNewGame(); err != nil {
		return nil, err
	}

	return g, nil
}

// DebugEventsOnTop puts Blood Moon and Winter Solstice as the first 2 cards,
// then other events, then zombies.
func (g *GameState) DebugEventsOnTop() {
	// Find Blood Moon and Winter Solstice specifically
	var bloodMoon, winterSolstice NightCard
	otherEvents := make([]NightCard, 0)
	zombies := make([]NightCard, 0)

	for _, card := range g.NightDeck {
		if card.IsEvent() {
			if card.Event.Name == "Blood Moon" {
				bloodMoon = card
			} else if card.Event.Name == "Winter Solstice" {
				winterSolstice = card
			} else {
				otherEvents = append(otherEvents, card)
			}
		} else {
			zombies = append(zombies, card)
		}
	}

	// Blood Moon first, Winter Solstice second, then other events, then zombies
	g.NightDeck = append([]NightCard{bloodMoon, winterSolstice}, otherEvents...)
	g.NightDeck = append(g.NightDeck, zombies...)
}

// creates new player with a fresh 5 card hand
func createPlayer(g *GameState, name string, numPlayers int, playerIdx int) *Player {
	// Apply color to player name if CLI mode
	displayName := name
	if cliMode {
		displayName = playerColors[playerIdx%len(playerColors)] + name + reset
	}

	return &Player{
		Name:  displayName,
		Lives: StartingLivesLookup[numPlayers],
		Farm: &Farm{
			Stacks:     Stacks{},
			NightCards: NightCards{},
		},
		Hand: Hand{
			HandItem{FarmItemType: g.nextDayCard()},
			HandItem{FarmItemType: g.nextDayCard()},
			HandItem{FarmItemType: g.nextDayCard()},
			HandItem{FarmItemType: g.nextDayCard()},
			HandItem{FarmItemType: g.nextDayCard()},
		},
		PlayChoices: PlayerPlayChoices{
			AutoloadShotgun:  true,
			AutoBuildHayWall: true,
		},
	}
}
