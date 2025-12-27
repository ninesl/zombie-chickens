package zombiechickens

func (g *GameState) NextDayCard() FarmItemType {
	if len(g.DayDeck) == 0 {
		g.RefillDayCards()
	}

	// could have out of bounds checks. not going to handle this bc unlikely with game logic
	card := g.DayDeck[0]
	g.DayDeck = g.DayDeck[1:]

	return card
}

func (g *GameState) RefillDayCards() {
	for farmItem, amount := range g.DiscardedDayCards {
		for range amount {
			g.DayDeck = append(g.DayDeck, farmItem)
		}
	}

	//TODO: shuffle DayDeck

	// clear
	g.DiscardedDayCards = map[FarmItemType]int{}
}

func (g *GameState) NextNightCard() NightCard {
	if len(g.NightDeck) == 0 {
		g.NightDeck = g.DiscardedNightCards
		// shuffle g.NightDeck
		g.DiscardedNightCards = make([]NightCard, 0)
	}
	nightCard := g.NightDeck[0]
	g.NightDeck = g.NightDeck[1:]

	return nightCard
}

func (g *GameState) DiscardNightCard(n NightCard) {
	if g.DiscardedNightCards == nil {
		g.DiscardedNightCards = make([]NightCard, 0)
	}

	g.DiscardedNightCards = append(g.DiscardedNightCards, n)
}
