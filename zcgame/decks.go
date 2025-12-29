package zcgame

// nextDayCard draws and returns the top card from the day deck.
// If the deck is empty or has only one card, it refills from the discard pile.
func (g *gameState) nextDayCard() FarmItemType {
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

// discardDayCard adds a card to the day discard pile.
func (g *gameState) discardDayCard(item FarmItemType) {
	g.DiscardedDayCards[item]++
}

// refillDayCards moves all discarded day cards back into the deck and shuffles.
func (g *gameState) refillDayCards() {
	for farmItem, amount := range g.DiscardedDayCards {
		for range amount {
			g.DayDeck = append(g.DayDeck, farmItem)
		}
	}

	shuffle(g.DayDeck)

	g.DiscardedDayCards = map[FarmItemType]int{} // clear
}

// nextNightCard draws and returns the top card from the night deck.
// If the deck is empty, it refills from the discard pile and shuffles.
func (g *gameState) nextNightCard() NightCard {
	if len(g.NightDeck) == 0 {
		g.NightDeck = g.DiscardedNightCards
		shuffle(g.NightDeck)
		g.DiscardedNightCards = make([]NightCard, 0)
	}
	nightCard := g.NightDeck[0]
	g.NightDeck = g.NightDeck[1:]

	return nightCard
}

// discardNightCard adds a night card to the discard pile.
func (g *gameState) discardNightCard(n NightCard) {
	if g.DiscardedNightCards == nil {
		g.DiscardedNightCards = make([]NightCard, 0)
	}

	g.DiscardedNightCards = append(g.DiscardedNightCards, n)
}
