package zcgame

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

func (g *GameState) CurrentPlayer() *Player {
	return g.Players[g.CurrentPlayerIdx]
}

func (g *GameState) NextPlayer() {
	g.CurrentPlayerIdx++
	if g.CurrentPlayerIdx >= len(g.Players) {
		g.CurrentPlayerIdx = 0
	}
}

// EliminatePlayer discards all of a player's cards and removes them from the game.
// Returns true if the player was eliminated.
func (g *GameState) EliminatePlayer(player *Player) {
	// Find the player's index
	playerIdx := -1
	for i, p := range g.Players {
		if p == player {
			playerIdx = i
			break
		}
	}
	if playerIdx == -1 {
		return
	}

	// Discard all farm cards
	for _, stack := range player.Farm.Stacks {
		for _, item := range stack {
			g.DiscardDayCard(item)
		}
	}
	player.Farm.Stacks = Stacks{}

	// Discard all hand cards
	for _, handItem := range player.Hand {
		if handItem.FarmItemType != NUM_FARM_ITEMS {
			g.DiscardDayCard(handItem.FarmItemType)
		}
	}
	player.Hand = Hand{}

	// Discard remaining night cards
	for _, nightCard := range player.Farm.NightCards {
		g.DiscardNightCard(nightCard)
	}
	player.Farm.NightCards = NightCards{}

	// Remove player from the list
	g.Players = append(g.Players[:playerIdx], g.Players[playerIdx+1:]...)

	// Adjust CurrentPlayerIdx if needed
	if g.CurrentPlayerIdx > playerIdx {
		g.CurrentPlayerIdx--
	} else if g.CurrentPlayerIdx >= len(g.Players) {
		g.CurrentPlayerIdx = 0
	}
}
func (g *GameState) GatherPlayerInput(msg string, choices ...int) int {
	return g.gatherPlayerInputWithRender(RefreshRender, msg, choices...)
}

func (g *GameState) GatherPlayerInputForDiscard(msg string, choices ...int) int {
	return g.gatherPlayerInputWithRender(RefreshRenderForDiscard, msg, choices...)
}

func (g *GameState) GatherPlayerInputForNight(msg string, choices ...int) int {
	return g.gatherPlayerInputWithRender(RefreshRenderForNight, msg, choices...)
}

func (g *GameState) gatherPlayerInputWithRender(renderFunc func(*GameState), msg string, choices ...int) int {
	renderFunc(g)
	fmt.Printf("%s: ", msg)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		if !scanner.Scan() {
			continue
		}
		text := scanner.Text()
		input, err := strconv.Atoi(text)
		if err != nil {
			fmt.Printf("ERROR, retry input: %s\n", IntSliceChoices(choices...))
			continue
		}

		for _, choice := range choices {
			if input == choice {
				return input
			}
		}
		fmt.Printf("ERROR, retry input: %s\n", IntSliceChoices(choices...))
	}
}

// TODO: this is for the CLI variant
func (g *GameState) DoPlayerDayTurn() {
	// Sort hand at the start so display matches indices
	g.CurrentPlayer().Hand.Sort()

	// Stage 1: Optional Discard
	g.StageInTurn = OptionalDiscard
	choices := []int{1, 2, 3, 4, 5, 0}
	input := g.GatherPlayerInput("1-5 to discard, 0 to skip", choices...)
	if input != 0 {
		// Discard selected card and draw new one
		g.DiscardDayCard(g.CurrentPlayer().Hand[input-1].FarmItemType)
		g.CurrentPlayer().Hand[input-1] = HandItem{FarmItemType: g.NextDayCard()}
		g.CurrentPlayer().Hand.Sort()
	}

	// Stage 2: Play 2 Cards
	g.StageInTurn = Play2Cards
	// First card - choices 1-5
	choices = []int{1, 2, 3, 4, 5}
	input = g.GatherPlayerInput("1-5 in your hand to play", choices...)
	g.CurrentPlayer().Farm.PlayCard(g.CurrentPlayer().Hand[input-1].FarmItemType, g.CurrentPlayer().PlayChoices)
	g.CurrentPlayer().Hand[input-1] = HandItem{FarmItemType: NUM_FARM_ITEMS} // blank
	g.CurrentPlayer().Hand.Sort()

	// Second card - choices 1-4 (hand is sorted, blank is at end)
	choices = []int{1, 2, 3, 4}
	input = g.GatherPlayerInput("1-4 in your hand to play", choices...)
	g.CurrentPlayer().Farm.PlayCard(g.CurrentPlayer().Hand[input-1].FarmItemType, g.CurrentPlayer().PlayChoices)
	g.CurrentPlayer().Hand[input-1] = HandItem{FarmItemType: NUM_FARM_ITEMS} // blank
	g.CurrentPlayer().Hand.Sort()

	// Stage 3: Draw 2 Cards - empty slots are at indices 3 and 4 after sorting
	g.StageInTurn = Draw2Cards
	choices = []int{1, 2}
	input = g.GatherPlayerInput(fmt.Sprintf("1 for public cards (%s, %s), 2 for deck", g.PublicDayCards[0], g.PublicDayCards[1]), choices...)
	if input == 1 {
		g.CurrentPlayer().Hand[3] = HandItem{FarmItemType: g.PublicDayCards[0]}
		g.CurrentPlayer().Hand[4] = HandItem{FarmItemType: g.PublicDayCards[1]}
		g.DealPublicDayCards()
	} else {
		g.CurrentPlayer().Hand[3] = HandItem{FarmItemType: g.NextDayCard()}
		g.CurrentPlayer().Hand[4] = HandItem{FarmItemType: g.NextDayCard()}
	}
}

func (g *GameState) DoNightTurn() {
	g.StageInTurn = Nighttime

	// Deal night cards to each player
	for _, player := range g.Players {
		for range g.NightNum {
			player.Farm.NightCards = append(player.Farm.NightCards, g.NextNightCard())
		}
	}

	// Process night cards: all players' [0], then all players' [0], etc.
	// Always process the first card and remove it after processing.
	// We keep looping until no player has any night cards left.
	for {
		if len(g.Players) == 0 {
			break
		}

		anyCardProcessed := false
		playersToProcess := len(g.Players)

		for i := 0; i < playersToProcess && len(g.Players) > 0; i++ {
			player := g.CurrentPlayer()

			if len(player.Farm.NightCards) > 0 {
				anyCardProcessed = true
				nightCard := player.Farm.NightCards[0]

				if nightCard.IsEvent() {
					RefreshRender(g)
					nightCard.Event.Action(g)
					g.GatherPlayerInput("press 0 to continue", 0)
					g.DiscardNightCard(nightCard)
					// Remove the processed event card (if not already cleared by event like Silent Night)
					if len(player.Farm.NightCards) > 0 {
						player.Farm.NightCards = player.Farm.NightCards[1:]
					}
				} else {
					// It's a zombie
					zc := ZombieChickens[nightCard.ZombieKey]
					zombieKilled := g.handleZombieAttack(player, zc)
					if zombieKilled {
						g.DiscardNightCard(nightCard)
					}

					// Check if player still exists (wasn't eliminated)
					// EliminatePlayer already discards all night cards, so only remove if player is still alive
					playerStillExists := false
					for _, p := range g.Players {
						if p == player {
							playerStillExists = true
							break
						}
					}

					if playerStillExists && len(player.Farm.NightCards) > 0 {
						player.Farm.NightCards = player.Farm.NightCards[1:]
					}
				}
			}

			// Only advance to next player if there are still players
			if len(g.Players) > 0 {
				g.NextPlayer()
			}
		}

		// Stop when no player had any cards
		if !anyCardProcessed {
			break
		}
	}
}

// handleZombieAttack processes a zombie attack and returns true if the zombie was killed.
func (g *GameState) handleZombieAttack(player *Player, zc ZombieChicken) bool {
	// Check for free kill first
	freeStacks := player.Farm.FindStacksThatCanKillForFree(zc)
	if len(freeStacks) > 0 {
		player.Farm.UseDefenseStack(freeStacks[0], zc, false, g)
		g.GatherPlayerInputForNight(fmt.Sprintf("%s: zombie auto-killed, press 0 to continue", player.Name), 0)
		return true
	}

	// Check for any available defense
	allStacks := player.Farm.FindStacksThatCanKill(zc)
	if len(allStacks) == 0 {
		g.GatherPlayerInputForNight(fmt.Sprintf("%s: no defense, will lose a life, press 0 to continue", player.Name), 0)
		player.Lives--
		if player.Lives <= 0 {
			g.GatherPlayerInputForNight(fmt.Sprintf("%s has been eliminated! Press 0 to continue", player.Name), 0)
			g.EliminatePlayer(player)
		}
		return false
	}

	// Convert to 1-based indices for display
	allStacks1Based := make([]int, len(allStacks))
	for i, idx := range allStacks {
		allStacks1Based[i] = idx + 1
	}

	// Player has options - let them choose
	choices := append(allStacks1Based, -1)
	msg := fmt.Sprintf("%s: choose stack to use or -1 to take life (stacks: %s)", player.Name, IntSliceChoices(allStacks1Based...))
	input := g.GatherPlayerInputForNight(msg, choices...)

	if input == -1 {
		g.GatherPlayerInputForNight(fmt.Sprintf("%s: will lose a life, press 0 to continue", player.Name), 0)
		player.Lives--
		if player.Lives <= 0 {
			g.GatherPlayerInputForNight(fmt.Sprintf("%s has been eliminated! Press 0 to continue", player.Name), 0)
			g.EliminatePlayer(player)
		}
		return false
	}

	// Convert back to 0-based for internal use
	stackIdx := input - 1

	// Check if exploding and player has shield
	useShield := false
	if zc.Traits.HasTrait(Exploding) && player.Farm.HasItemInStacks(Shield) {
		msg := fmt.Sprintf("%s: use shield to save stack from exploding zombie? (1=yes, 0=no)", player.Name)
		useShield = g.GatherPlayerInputForNight(msg, 0, 1) == 1
	}

	player.Farm.UseDefenseStack(stackIdx, zc, useShield, g)
	return true
}

// DoDay runs a full day cycle. Returns true if the game should continue, false if game over.
func (g *GameState) DoDay() bool {
	if len(g.Players) == 0 {
		return false
	}

	// Morning: each player does all 3 day stages
	g.Turn = Morning
	for range len(g.Players) {
		g.DoPlayerDayTurn()
		g.NextPlayer()
	}
	RefreshRender(g)

	// Afternoon: each player does all 3 day stages
	// TODO: check for Day/co-op skip afternoon
	g.Turn = Afternoon
	for range len(g.Players) {
		g.DoPlayerDayTurn()
		g.NextPlayer()
	}
	RefreshRender(g)

	// Night: each player fights zombies
	g.Turn = Night
	g.DoNightTurn()

	if len(g.Players) == 0 {
		return false
	}

	RefreshRender(g)
	g.NightNum++
	return true
}

func (g *GameState) HasLivingPlayers() bool {
	for _, player := range g.Players {
		if player.Lives > 0 {
			return true
		}
	}
	return false
}

// discardCardsFromAllFarms prompts each player to discard n cards from their farm.
func (g *GameState) discardCardsFromAllFarms(n int) {
	for _, player := range g.Players {
		totalItems := player.Farm.Stacks.TotalItems()

		// Skip players with no farm items
		if totalItems == 0 {
			continue
		}

		// If n >= total items, discard everything
		if n >= totalItems {
			for _, stack := range player.Farm.Stacks {
				for _, item := range stack {
					g.DiscardDayCard(item)
				}
			}
			player.Farm.Stacks = Stacks{}
			continue
		}

		for i := 0; i < n; i++ {
			totalItems := player.Farm.Stacks.TotalItems()
			if totalItems == 0 {
				break
			}

			choices := make([]int, totalItems)
			for j := range choices {
				choices[j] = j + 1
			}

			input := g.GatherPlayerInputForDiscard(
				fmt.Sprintf("%s: choose card to discard (%d/%d)", player.Name, i+1, n),
				choices...,
			)

			player.Farm.RemoveItemByFlatIndex(input-1, g)
		}
	}
}
