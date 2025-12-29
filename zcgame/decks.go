package zcgame

// pops the first element of g.DayDeck, refills and shuffles when g.DayDeck <= 1
func (g *GameState) nextDayCard() FarmItemType {
	if len(g.DayDeck) == 0 {
		g.refillDayCards()
	} else if len(g.DayDeck) == 1 {
		card := g.DayDeck[0]
		g.DayDeck = g.DayDeck[:0] // clear
		g.refillDayCards()
		return card
	}

	card := g.DayDeck[0]
	g.DayDeck = g.DayDeck[1:]

	return card
}

func (g *GameState) discardDayCard(item FarmItemType) {
	g.DiscardedDayCards[item]++
}

func (g *GameState) refillDayCards() {
	for farmItem, amount := range g.DiscardedDayCards {
		for range amount {
			g.DayDeck = append(g.DayDeck, farmItem)
		}
	}

	shuffle(g.DayDeck)

	g.DiscardedDayCards = map[FarmItemType]int{} // clear
}

func (g *GameState) nextNightCard() NightCard {
	if len(g.NightDeck) == 0 {
		g.NightDeck = g.DiscardedNightCards
		shuffle(g.NightDeck)
		g.DiscardedNightCards = make([]NightCard, 0)
	}
	nightCard := g.NightDeck[0]
	g.NightDeck = g.NightDeck[1:]

	return nightCard
}

func (g *GameState) discardNightCard(n NightCard) {
	if g.DiscardedNightCards == nil {
		g.DiscardedNightCards = make([]NightCard, 0)
	}

	g.DiscardedNightCards = append(g.DiscardedNightCards, n)
}
