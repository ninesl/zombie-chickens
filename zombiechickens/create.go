package zombiechickens

// random order based on DayCards global variable configs
func createDayDeck() Stack {
	var deck = Stack{}
	for farmItem, amount := range DayCardAmounts {
		for range amount {
			deck = append(deck, farmItem)
		}
	}
	return deck
}

func createNightDeck() []NightCard {
	return nil
}

func CreateNewGame(numPlayers int) *GameState {
	var g = &GameState{
		DayDeck:   createDayDeck(),
		NightDeck: createNightDeck(),
	}

	var players = make([]*Player, 0, numPlayers)
	for range numPlayers {
		players = append(players, &Player{
			Farm:        &Farm{},
			CardsInHand: [5]FarmItemType{},
		})
	}
	g.Players = players

	return g
}
