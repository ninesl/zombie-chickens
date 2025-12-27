package zombiechickens

// random order based on DayCards global variable configs
func CreateDayDeck() Stack {
	var deck = Stack{}
	for farmItem, amount := range DayCardAmounts {
		for range amount {
			deck = append(deck, farmItem)
		}
	}

	//Shuffle

	return deck
}

func CreateNightDeck() []NightCard {
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

	//shuffle

	return deck
}

func CreateNewGame(numPlayers int) *GameState {
	var g = &GameState{
		DayDeck:   CreateDayDeck(),
		NightDeck: CreateNightDeck(),
	}

	var players = make([]*Player, 0, numPlayers)

	for range numPlayers {
		players = append(players, newPlayer())
	}
	g.Players = players

	return g
}

func newPlayer() *Player {
	return &Player{
		Farm: &Farm{
			Stacks:     []Stack{},
			NightCards: []NightCard{},
		},
		CardsInHand: [5]FarmItemType{},
	}
}
