package zcgame

import "fmt"

// createDayDeck creates a shuffled deck of day cards based on DayCardAmounts.
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

// createNightDeck creates a shuffled deck of night cards containing all zombies and events.
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

// StartingLivesLookup maps player count to starting lives per player.
// Fewer players get more lives to balance difficulty.
var StartingLivesLookup = map[int]int{
	1: 5,
	2: 5,
	3: 4,
	4: 4,
}

// dealPublicDayCards draws two cards from the day deck to be the public cards.
func (g *gameState) dealPublicDayCards() {
	g.PublicDayCards = [2]FarmItemType{g.nextDayCard(), g.nextDayCard()}
}

// CreateNewGame initializes a new game with the given player names.
// Returns a GameView for interacting with the game, or an error if the
// player count is invalid (must be 1-4 players).
//
// The game is initialized with:
//   - Shuffled day and night decks
//   - Each player with starting lives (based on player count) and 5 cards
//   - Two public day cards face-up for drawing
//   - Turn set to Morning, ready for the first player's turn
//
// After creation, call ContinueDay() on the returned GameView to begin the game.
func CreateNewGame(playerNames ...string) (GameView, error) {
	if len(playerNames) == 0 {
		return GameView{}, fmt.Errorf("must provide at least 1 player")
	} else if len(playerNames) > 4 {
		return GameView{}, fmt.Errorf("must provide max 4 player names")
	}

	var g = &gameState{
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
		return GameView{}, err
	}

	return NewGameView(g), nil
}

// DebugEventsOnTop reorders the night deck for testing purposes.
// Places Blood Moon first, Winter Solstice second, then other events, then zombies.
// This allows predictable testing of event handling.
func (g *gameState) DebugEventsOnTop() {
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

// createPlayer creates a new player with starting lives and a 5-card hand.
// The player's name may be colored with ANSI codes if CLI mode is enabled.
func createPlayer(g *gameState, name string, numPlayers int, playerIdx int) *Player {
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
