package zcgame

import (
	"fmt"
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

// DoPlayerDayTurn processes one player's day turn.
// Returns nil when the turn is complete, or *PlayerInputNeeded when input is required.
// Call ProvideInput with the player's choice, then call DoPlayerDayTurn again to continue.
func (g *GameState) DoPlayerDayTurn() *PlayerInputNeeded {
	player := g.CurrentPlayer()

	for {
		switch g.DaySubStage {
		case DaySubStageOptionalDiscard:
			player.Hand.Sort()
			g.StageInTurn = OptionalDiscard
			return &PlayerInputNeeded{
				Context:      InputContextDiscard,
				RenderType:   RenderNormal,
				Message:      "1-5 to discard, 0 to skip",
				ValidChoices: []int{1, 2, 3, 4, 5, 0},
			}

		case DaySubStagePlay1:
			g.StageInTurn = Play2Cards
			return &PlayerInputNeeded{
				Context:      InputContextPlay,
				RenderType:   RenderNormal,
				Message:      "1-5 in your hand to play",
				ValidChoices: []int{1, 2, 3, 4, 5},
			}

		case DaySubStagePlay1Stack:
			// Player needs to choose which stack to place the pending card
			return g.createStackSelectionInput()

		case DaySubStagePlay2:
			return &PlayerInputNeeded{
				Context:      InputContextPlay,
				RenderType:   RenderNormal,
				Message:      "1-4 in your hand to play",
				ValidChoices: []int{1, 2, 3, 4},
			}

		case DaySubStagePlay2Stack:
			// Player needs to choose which stack to place the pending card
			return g.createStackSelectionInput()

		case DaySubStageDraw:
			g.StageInTurn = Draw2Cards
			return &PlayerInputNeeded{
				Context:      InputContextDraw,
				RenderType:   RenderNormal,
				Message:      fmt.Sprintf("1 for public cards (%s, %s), 2 for deck", g.PublicDayCards[0], g.PublicDayCards[1]),
				ValidChoices: []int{1, 2},
			}

		default:
			// Turn complete - reset for next player
			g.DaySubStage = DaySubStageOptionalDiscard
			return nil
		}
	}
}

// createStackSelectionInput creates the input request for stack selection during card play
func (g *GameState) createStackSelectionInput() *PlayerInputNeeded {
	player := g.CurrentPlayer()

	// Get valid stacks from PlayCard
	result := player.Farm.PlayCard(g.PendingCardItem, player.PlayChoices)
	if result == nil {
		// Card was auto-played, no input needed
		return nil
	}

	// Convert 0-based stack indices to 1-based for display, add 0 for "new stack"
	choices := make([]int, len(result.ValidStacks)+1)
	choices[0] = 0 // new stack option
	for i, idx := range result.ValidStacks {
		choices[i+1] = idx + 1
	}

	return &PlayerInputNeeded{
		Context:      InputContextPlayCard,
		RenderType:   RenderNormal,
		Message:      fmt.Sprintf("%s: %s (0 for new stack, or choose stack)", g.PendingCardItem, result.Message),
		ValidChoices: choices,
		Item:         g.PendingCardItem,
		ValidStacks:  result.ValidStacks,
	}
}

// ProvideInput provides the player's input and continues game execution.
// Returns nil when the current operation is complete, or *PlayerInputNeeded if more input is needed.
func (g *GameState) ProvideInput(choice int) *PlayerInputNeeded {
	switch g.Turn {
	case Morning, Afternoon:
		return g.provideDayInput(choice)
	case Night:
		return g.provideNightInput(choice)
	}
	return nil
}

// provideDayInput handles input during day turns
func (g *GameState) provideDayInput(choice int) *PlayerInputNeeded {
	player := g.CurrentPlayer()

	switch g.DaySubStage {
	case DaySubStageOptionalDiscard:
		if choice != 0 {
			g.DiscardDayCard(player.Hand[choice-1].FarmItemType)
			player.Hand[choice-1] = HandItem{FarmItemType: g.NextDayCard()}
			player.Hand.Sort()
		}
		g.DaySubStage = DaySubStagePlay1
		return g.DoPlayerDayTurn()

	case DaySubStagePlay1:
		g.PendingCardItem = player.Hand[choice-1].FarmItemType
		result := player.Farm.PlayCard(g.PendingCardItem, player.PlayChoices)
		if result != nil {
			// Need stack selection
			g.DaySubStage = DaySubStagePlay1Stack
			return g.DoPlayerDayTurn()
		}
		// Card played successfully
		player.Hand[choice-1] = HandItem{FarmItemType: NUM_FARM_ITEMS}
		player.Hand.Sort()
		g.DaySubStage = DaySubStagePlay2
		return g.DoPlayerDayTurn()

	case DaySubStagePlay1Stack:
		if choice == 0 {
			// New stack
			player.Farm.makeStackWith(g.PendingCardItem)
		} else {
			// Add to existing stack
			player.Farm.addToStackIndex(g.PendingCardItem, choice-1)
		}
		// Find and blank the played card from hand
		for i := range player.Hand {
			if player.Hand[i].FarmItemType == g.PendingCardItem {
				player.Hand[i] = HandItem{FarmItemType: NUM_FARM_ITEMS}
				break
			}
		}
		player.Hand.Sort()
		g.DaySubStage = DaySubStagePlay2
		return g.DoPlayerDayTurn()

	case DaySubStagePlay2:
		g.PendingCardItem = player.Hand[choice-1].FarmItemType
		result := player.Farm.PlayCard(g.PendingCardItem, player.PlayChoices)
		if result != nil {
			// Need stack selection
			g.DaySubStage = DaySubStagePlay2Stack
			return g.DoPlayerDayTurn()
		}
		// Card played successfully
		player.Hand[choice-1] = HandItem{FarmItemType: NUM_FARM_ITEMS}
		player.Hand.Sort()
		g.DaySubStage = DaySubStageDraw
		return g.DoPlayerDayTurn()

	case DaySubStagePlay2Stack:
		if choice == 0 { // New stack
			player.Farm.makeStackWith(g.PendingCardItem)
		} else { // Add to existing stack
			player.Farm.addToStackIndex(g.PendingCardItem, choice-1)
		}
		// Find and blank the played card from hand
		for i := range player.Hand {
			if player.Hand[i].FarmItemType == g.PendingCardItem {
				player.Hand[i] = HandItem{FarmItemType: NUM_FARM_ITEMS}
				break
			}
		}
		player.Hand.Sort()
		g.DaySubStage = DaySubStageDraw
		return g.DoPlayerDayTurn()

	case DaySubStageDraw:
		if choice == 1 {
			player.Hand[3] = HandItem{FarmItemType: g.PublicDayCards[0]}
			player.Hand[4] = HandItem{FarmItemType: g.PublicDayCards[1]}
			g.DealPublicDayCards()
		} else {
			player.Hand[3] = HandItem{FarmItemType: g.NextDayCard()}
			player.Hand[4] = HandItem{FarmItemType: g.NextDayCard()}
		}
		// Turn complete - reset substage and advance player
		g.DaySubStage = DaySubStageOptionalDiscard
		g.PlayerTurnIndex++
		g.NextPlayer()
		return nil
	}

	return nil
}

// DoNightTurn processes the night phase.
// Returns nil when complete, or *PlayerInputNeeded when input is required.
func (g *GameState) DoNightTurn() *PlayerInputNeeded {
	g.StageInTurn = Nighttime

	// Deal night cards once at the start of night
	if !g.NightCardsDealt {
		for _, player := range g.Players {
			for range g.NightNum {
				player.Farm.NightCards = append(player.Farm.NightCards, g.NextNightCard())
			}
		}
		g.NightCardsDealt = true
		g.NightPlayerIndex = 0
		g.NightPlayersToProcess = len(g.Players)
		g.NightAnyCardProcessed = false
	}

	return g.processNightCards()
}

// processNightCards continues processing night cards
func (g *GameState) processNightCards() *PlayerInputNeeded {
	for {
		if len(g.Players) == 0 {
			g.resetNightState()
			return nil
		}

		// Check if we're resuming from a sub-stage
		switch g.NightSubStage {
		case NightSubStageZombieAutoKilled:
			return &PlayerInputNeeded{
				Context:      InputContextConfirm,
				RenderType:   RenderForNight,
				Message:      fmt.Sprintf("%s: zombie auto-killed, press 0 to continue", g.CurrentPlayer().Name),
				ValidChoices: []int{0},
			}

		case NightSubStageNoDefense:
			return &PlayerInputNeeded{
				Context:      InputContextConfirm,
				RenderType:   RenderForNight,
				Message:      fmt.Sprintf("%s: no defense, will lose a life, press 0 to continue", g.CurrentPlayer().Name),
				ValidChoices: []int{0},
			}

		case NightSubStageEliminated:
			return &PlayerInputNeeded{
				Context:      InputContextConfirm,
				RenderType:   RenderForNight,
				Message:      fmt.Sprintf("%s has been eliminated! Press 0 to continue", g.CurrentPlayer().Name),
				ValidChoices: []int{0},
			}

		case NightSubStageChooseDefense:
			return g.createDefenseChoiceInput()

		case NightSubStageChooseShield:
			return &PlayerInputNeeded{
				Context:      InputContextShield,
				RenderType:   RenderForNight,
				Message:      fmt.Sprintf("%s: use shield to save stack from exploding zombie? (1=yes, 0=no)", g.CurrentPlayer().Name),
				ValidChoices: []int{1, 0},
			}

		case NightSubStageConfirmLifeLoss:
			return &PlayerInputNeeded{
				Context:      InputContextConfirm,
				RenderType:   RenderForNight,
				Message:      fmt.Sprintf("%s: will lose a life, press 0 to continue", g.CurrentPlayer().Name),
				ValidChoices: []int{0},
			}

		case NightSubStageEventConfirm:
			// Use saved event info (not CurrentNightCard, which may have changed)
			return &PlayerInputNeeded{
				Context:      InputContextConfirm,
				RenderType:   RenderForNight,
				Message:      fmt.Sprintf("%s: %s Press 0 to continue", g.PendingEventName, g.PendingEventDesc),
				ValidChoices: []int{0},
			}

		case NightSubStageEventDiscard:
			return g.createEventDiscardInput()
		}

		// Normal processing
		for g.NightPlayerIndex < g.NightPlayersToProcess && len(g.Players) > 0 {
			player := g.CurrentPlayer()

			if len(player.Farm.NightCards) > 0 {
				g.NightAnyCardProcessed = true
				nightCard := player.Farm.NightCards[0]
				g.CurrentNightCard = &nightCard

				if nightCard.IsEvent() {
					// Process event
					return g.processEventCard(nightCard)
				} else {
					// Process zombie
					return g.processZombieCard(nightCard)
				}
			}

			// Move to next player
			g.NightPlayerIndex++
			if len(g.Players) > 0 {
				g.NextPlayer()
			}
		}

		// Finished one round of all players
		if !g.NightAnyCardProcessed {
			// No more cards to process
			g.resetNightState()
			return nil
		}

		// Start next round
		g.NightPlayerIndex = 0
		g.NightPlayersToProcess = len(g.Players)
		g.NightAnyCardProcessed = false
	}
}

// processEventCard handles an event card
// startEventDiscard initializes the event discard state and returns input request if needed.
// Called from event Action functions for Lightning Storm and Tornado.
func (g *GameState) startEventDiscard(n int) *PlayerInputNeeded {
	g.EventDiscardTotal = n
	g.EventDiscardRemaining = n
	g.EventDiscardStartIdx = g.CurrentPlayerIdx // Start from player who drew the event
	g.EventDiscardPlayerIdx = 0                 // Offset from start (0 = current player)
	g.NightSubStage = NightSubStageEventDiscard
	return g.createEventDiscardInput()
}

func (g *GameState) processEventCard(nightCard NightCard) *PlayerInputNeeded {
	// Save event info for confirmation display
	g.PendingEventName = nightCard.Event.Name
	g.PendingEventDesc = nightCard.Event.Description

	// Show confirmation FIRST, before running the action
	g.NightSubStage = NightSubStageEventConfirm
	return g.processNightCards()
}

// createEventDiscardInput creates input request for event-based discards
func (g *GameState) createEventDiscardInput() *PlayerInputNeeded {
	// Find the player who needs to discard
	// Players are processed starting from EventDiscardStartIdx and cycling through
	for g.EventDiscardPlayerIdx < len(g.Players) {
		// Calculate actual player index by cycling from start
		actualIdx := (g.EventDiscardStartIdx + g.EventDiscardPlayerIdx) % len(g.Players)
		player := g.Players[actualIdx]
		totalItems := player.Farm.Stacks.TotalItems()

		if totalItems == 0 {
			g.EventDiscardPlayerIdx++
			continue
		}

		// If remaining >= total, discard all and move on
		if g.EventDiscardRemaining >= totalItems {
			for _, stack := range player.Farm.Stacks {
				for _, item := range stack {
					g.DiscardDayCard(item)
				}
			}
			player.Farm.Stacks = Stacks{}
			g.EventDiscardPlayerIdx++
			g.EventDiscardRemaining = g.EventDiscardTotal // Reset for next player
			continue
		}

		// Need player to choose
		choices := make([]int, totalItems)
		for j := range choices {
			choices[j] = j + 1
		}

		discardNum := g.EventDiscardTotal - g.EventDiscardRemaining + 1
		return &PlayerInputNeeded{
			Context:      InputContextEventDiscard,
			RenderType:   RenderForDiscard,
			Message:      fmt.Sprintf("%s: choose card to discard (%d/%d)", player.Name, discardNum, g.EventDiscardTotal),
			ValidChoices: choices,
		}
	}

	// All players done discarding
	// NOW discard and remove the event card
	player := g.CurrentPlayer()
	if len(player.Farm.NightCards) > 0 {
		nightCard := player.Farm.NightCards[0]
		g.DiscardNightCard(nightCard)
		player.Farm.NightCards = player.Farm.NightCards[1:]
	}

	// Move to next player in night round
	g.NightPlayerIndex++
	if len(g.Players) > 0 {
		g.NextPlayer()
	}
	g.NightSubStage = NightSubStageProcessCards
	return g.processNightCards()
}

// processZombieCard handles a zombie attack
func (g *GameState) processZombieCard(nightCard NightCard) *PlayerInputNeeded {
	player := g.CurrentPlayer()
	zc := ZombieChickens[nightCard.ZombieKey]
	g.CurrentZombie = &zc

	// Check for free kill first
	freeStacks := player.Farm.FindStacksThatCanKillForFree(zc)
	if len(freeStacks) > 0 {
		player.Farm.UseDefenseStack(freeStacks[0], zc, false, g)
		// Don't remove night card yet - keep it visible for confirmation display
		g.NightSubStage = NightSubStageZombieAutoKilled
		return g.processNightCards()
	}

	// Check for any available defense
	allStacks := player.Farm.FindStacksThatCanKill(zc)
	if len(allStacks) == 0 {
		g.NightSubStage = NightSubStageNoDefense
		return g.processNightCards()
	}

	// Player has options - let them choose
	g.NightSubStage = NightSubStageChooseDefense
	return g.processNightCards()
}

// createDefenseChoiceInput creates the input request for defense stack selection
func (g *GameState) createDefenseChoiceInput() *PlayerInputNeeded {
	player := g.CurrentPlayer()
	zc := *g.CurrentZombie
	allStacks := player.Farm.FindStacksThatCanKill(zc)

	// Convert to 1-based indices for display
	allStacks1Based := make([]int, len(allStacks))
	for i, idx := range allStacks {
		allStacks1Based[i] = idx + 1
	}

	choices := append(allStacks1Based, -1)
	return &PlayerInputNeeded{
		Context:      InputContextDefense,
		RenderType:   RenderForNight,
		Message:      fmt.Sprintf("%s: choose stack to use or -1 to take life (stacks: %s)", player.Name, IntSliceChoices(allStacks1Based...)),
		ValidChoices: choices,
		ValidStacks:  allStacks,
	}
}

// provideNightInput handles input during night phase
func (g *GameState) provideNightInput(choice int) *PlayerInputNeeded {
	player := g.CurrentPlayer()

	switch g.NightSubStage {
	case NightSubStageZombieAutoKilled:
		// Now remove the night card after confirmation
		if g.CurrentNightCard != nil {
			g.DiscardNightCard(*g.CurrentNightCard)
		}
		if len(player.Farm.NightCards) > 0 {
			player.Farm.NightCards = player.Farm.NightCards[1:]
		}
		// Continue to next player
		g.NightPlayerIndex++
		if len(g.Players) > 0 {
			g.NextPlayer()
		}
		g.NightSubStage = NightSubStageProcessCards
		return g.processNightCards()

	case NightSubStageNoDefense:
		player.Lives--
		if player.Lives <= 0 {
			g.NightSubStage = NightSubStageEliminated
			return g.processNightCards()
		}
		// Remove the night card
		if len(player.Farm.NightCards) > 0 {
			player.Farm.NightCards = player.Farm.NightCards[1:]
		}
		g.NightPlayerIndex++
		if len(g.Players) > 0 {
			g.NextPlayer()
		}
		g.NightSubStage = NightSubStageProcessCards
		return g.processNightCards()

	case NightSubStageEliminated:
		playerToEliminate := player
		g.EliminatePlayer(playerToEliminate)
		g.NightPlayerIndex++
		// Don't call NextPlayer - EliminatePlayer already adjusted indices
		g.NightSubStage = NightSubStageProcessCards
		return g.processNightCards()

	case NightSubStageChooseDefense:
		if choice == -1 {
			g.NightSubStage = NightSubStageConfirmLifeLoss
			return g.processNightCards()
		}
		g.ChosenStackIdx = choice - 1

		// Check if we need shield prompt - skip if stack contains one-time-use items
		// (except Ammo, since Shotgun+Ammo stacks should still prompt for shield)
		stackUsed := player.Farm.Stacks[g.ChosenStackIdx]
		skipShieldPrompt := false
		for _, item := range stackUsed {
			if item != Ammo && item.IsOneTimeUse() {
				skipShieldPrompt = true
				break
			}
		}

		zc := *g.CurrentZombie
		if !skipShieldPrompt && zc.Traits.HasTrait(Exploding) && player.Farm.HasItemInStacks(Shield) {
			g.NightSubStage = NightSubStageChooseShield
			return g.processNightCards()
		}

		// Use defense without shield
		player.Farm.UseDefenseStack(g.ChosenStackIdx, zc, false, g)
		if g.CurrentNightCard != nil {
			g.DiscardNightCard(*g.CurrentNightCard)
		}
		if len(player.Farm.NightCards) > 0 {
			player.Farm.NightCards = player.Farm.NightCards[1:]
		}
		g.NightPlayerIndex++
		if len(g.Players) > 0 {
			g.NextPlayer()
		}
		g.NightSubStage = NightSubStageProcessCards
		return g.processNightCards()

	case NightSubStageChooseShield:
		useShield := choice == 1
		zc := *g.CurrentZombie
		player.Farm.UseDefenseStack(g.ChosenStackIdx, zc, useShield, g)
		if g.CurrentNightCard != nil {
			g.DiscardNightCard(*g.CurrentNightCard)
		}
		if len(player.Farm.NightCards) > 0 {
			player.Farm.NightCards = player.Farm.NightCards[1:]
		}
		g.NightPlayerIndex++
		if len(g.Players) > 0 {
			g.NextPlayer()
		}
		g.NightSubStage = NightSubStageProcessCards
		return g.processNightCards()

	case NightSubStageConfirmLifeLoss:
		player.Lives--
		if player.Lives <= 0 {
			g.NightSubStage = NightSubStageEliminated
			return g.processNightCards()
		}
		if len(player.Farm.NightCards) > 0 {
			player.Farm.NightCards = player.Farm.NightCards[1:]
		}
		g.NightPlayerIndex++
		if len(g.Players) > 0 {
			g.NextPlayer()
		}
		g.NightSubStage = NightSubStageProcessCards
		return g.processNightCards()

	case NightSubStageEventConfirm:
		// User confirmed - now run the event action
		player := g.CurrentPlayer()
		nightCard := player.Farm.NightCards[0]

		// Run the event action - it may return an input request (e.g., for discards)
		inputNeeded := nightCard.Event.Action(g)
		if inputNeeded != nil {
			// Action needs more input (like Tornado discards)
			// DON'T remove the card yet - it stays visible until action is complete
			return inputNeeded
		}

		// Action complete with no further input needed
		// NOW remove and discard the event card
		// (must be after action for events like Blood Moon that append to NightCards)
		g.DiscardNightCard(nightCard)
		if len(player.Farm.NightCards) > 0 {
			player.Farm.NightCards = player.Farm.NightCards[1:]
		}

		// Move to next player
		g.NightPlayerIndex++
		if len(g.Players) > 0 {
			g.NextPlayer()
		}
		g.NightSubStage = NightSubStageProcessCards
		return g.processNightCards()

	case NightSubStageEventDiscard:
		// Process the discard - calculate actual player index by cycling from start
		actualIdx := (g.EventDiscardStartIdx + g.EventDiscardPlayerIdx) % len(g.Players)
		player := g.Players[actualIdx]
		player.Farm.RemoveItemByFlatIndex(choice-1, g)
		g.EventDiscardRemaining--

		if g.EventDiscardRemaining <= 0 {
			g.EventDiscardPlayerIdx++
			g.EventDiscardRemaining = g.EventDiscardTotal
		}
		return g.createEventDiscardInput()
	}

	return nil
}

// resetNightState resets night processing state (but NOT NightCardsDealt - that's reset on transition to night)
func (g *GameState) resetNightState() {
	// NOTE: Do NOT reset NightCardsDealt here - it must persist until we actually transition
	// to the next day. Otherwise cards get re-dealt when DoDay() calls DoNightTurn() after
	// ContinueAfterInput finishes processing.
	g.NightSubStage = NightSubStageProcessCards
	g.NightPlayerIndex = 0
	g.NightPlayersToProcess = 0
	g.NightAnyCardProcessed = false
	g.CurrentNightCard = nil
	g.CurrentZombie = nil
	g.ChosenStackIdx = 0
	g.EventDiscardStartIdx = 0
	g.EventDiscardPlayerIdx = 0
	g.EventDiscardRemaining = 0
	g.EventDiscardTotal = 0
}

// DoDay runs a full day cycle.
// Returns (true, nil) if the day completed and game continues.
// Returns (false, nil) if game over.
// Returns (bool, *PlayerInputNeeded) if input is needed.
func (g *GameState) DoDay() (bool, *PlayerInputNeeded) {
	if len(g.Players) == 0 {
		return false, nil
	}

	// Process based on current turn
	switch g.Turn {
	case Morning:
		for g.PlayerTurnIndex < len(g.Players) {
			inputNeeded := g.DoPlayerDayTurn()
			if inputNeeded != nil {
				return true, inputNeeded
			}
			// Player turn completed (PlayerTurnIndex incremented in provideDayInput)
		}
		// Morning complete - move to afternoon
		g.PlayerTurnIndex = 0
		g.Turn = Afternoon
		return g.DoDay()

	case Afternoon:
		for g.PlayerTurnIndex < len(g.Players) {
			inputNeeded := g.DoPlayerDayTurn()
			if inputNeeded != nil {
				return true, inputNeeded
			}
			// Player turn completed (PlayerTurnIndex incremented in provideDayInput)
		}
		// Afternoon complete - move to night
		g.PlayerTurnIndex = 0
		g.Turn = Night
		return g.DoDay()

	case Night:
		inputNeeded := g.DoNightTurn()
		if inputNeeded != nil {
			return true, inputNeeded
		}
		// Night complete
		if len(g.Players) == 0 {
			return false, nil
		}
		// Prepare for next day - reset NightCardsDealt here so next night will deal new cards
		g.NightCardsDealt = false
		g.NightNum++
		g.Turn = Morning
		g.PlayerTurnIndex = 0
		return true, nil
	}

	return true, nil
}

// ContinueAfterInput resumes game execution after player provides input.
// Returns (gameOver bool, inputNeeded *PlayerInputNeeded)
// gameOver is true if the game has ended.
// inputNeeded is non-nil if more input is required.
func (g *GameState) ContinueAfterInput(choice int) (bool, *PlayerInputNeeded) {
	inputNeeded := g.ProvideInput(choice)
	if inputNeeded != nil {
		return false, inputNeeded
	}

	// Continue the day
	return g.DoDay()
}

func (g *GameState) HasLivingPlayers() bool {
	for _, player := range g.Players {
		if player.Lives > 0 {
			return true
		}
	}
	return false
}
