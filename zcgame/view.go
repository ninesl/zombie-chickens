package zcgame

// GameView wraps a gameState and provides read-only access with copies.
// Use this for API/frontend consumption to prevent accidental mutation of game state.
// GameView is a small value type (just a pointer) and should be passed by value.
type GameView struct {
	game *gameState
}

// newGameView creates a GameView wrapper around the given gameState.
// This is unexported since gameState is unexported - use CreateNewGame instead.
func NewGameView(g *gameState) GameView {
	return GameView{game: g}
}

// --- State Machine Control Methods ---

// ContinueDay advances the game through the current day cycle.
// See gameState.ContinueDay for full documentation.
func (v GameView) ContinueDay() (bool, *PlayerInputNeeded) {
	return v.game.ContinueDay()
}

// ContinueAfterInput resumes game execution after player provides input.
// See gameState.ContinueAfterInput for full documentation.
func (v GameView) ContinueAfterInput(choice int) (bool, *PlayerInputNeeded) {
	return v.game.ContinueAfterInput(choice)
}

// DebugEventsOnTop moves event cards to top of night deck for testing.
func (v GameView) DebugEventsOnTop() {
	v.game.DebugEventsOnTop()
}

// --- Read-Only Accessors (return copies, not pointers) ---

// Turn returns the current turn phase.
func (v GameView) Turn() Turn {
	return v.game.Turn
}

// NightNum returns the current night number.
func (v GameView) NightNum() int {
	return v.game.NightNum
}

// StageInTurn returns the current stage within the turn.
func (v GameView) StageInTurn() StageInTurn {
	return v.game.StageInTurn
}

// CurrentPlayerIdx returns the index of the current player.
func (v GameView) CurrentPlayerIdx() int {
	return v.game.CurrentPlayerIdx
}

// ActiveInputPlayerIdx returns the index of the player who should provide input.
// This differs from CurrentPlayerIdx during event discards, where multiple players
// take turns discarding but CurrentPlayerIdx stays on the player who drew the event.
func (v GameView) ActiveInputPlayerIdx() int {
	// During event discards, calculate the actual player who needs to discard
	if v.game.NightSubStage == NightSubStageEventDiscard {
		return (v.game.EventDiscardStartIdx + v.game.EventDiscardPlayerIdx) % len(v.game.Players)
	}
	return v.game.CurrentPlayerIdx
}

// PublicDayCards returns a copy of the public day cards.
func (v GameView) PublicDayCards() PublicDayCards {
	return v.game.PublicDayCards // Already a value type [2]FarmItemType
}

// DayDeckCount returns the number of cards remaining in the day deck.
func (v GameView) DayDeckCount() int {
	return len(v.game.DayDeck)
}

// NightDeckCount returns the number of cards remaining in the night deck.
func (v GameView) NightDeckCount() int {
	return len(v.game.NightDeck)
}

// DiscardedDayCards returns a copy of the discarded day cards map.
func (v GameView) DiscardedDayCards() map[FarmItemType]int {
	result := make(map[FarmItemType]int, len(v.game.DiscardedDayCards))
	for k, val := range v.game.DiscardedDayCards {
		result[k] = val
	}
	return result
}

// DiscardedNightCards returns a copy of the discarded night cards.
func (v GameView) DiscardedNightCards() NightCards {
	result := make(NightCards, len(v.game.DiscardedNightCards))
	copy(result, v.game.DiscardedNightCards)
	return result
}

// PlayerCount returns the number of active players.
func (v GameView) PlayerCount() int {
	return len(v.game.Players)
}

// Players returns PlayerView wrappers for all active players.
func (v GameView) Players() []PlayerView {
	result := make([]PlayerView, len(v.game.Players))
	for i, p := range v.game.Players {
		result[i] = PlayerView{player: p}
	}
	return result
}

// Player returns a PlayerView for the player at the given index.
// Returns an empty PlayerView if index is out of bounds.
func (v GameView) Player(idx int) PlayerView {
	if idx < 0 || idx >= len(v.game.Players) {
		return PlayerView{}
	}
	return PlayerView{player: v.game.Players[idx]}
}

// CurrentPlayer returns the PlayerView for the current player.
func (v GameView) CurrentPlayer() PlayerView {
	return v.Player(v.game.CurrentPlayerIdx)
}

// PlayerIdxByName returns the current index of a player by name, or -1 if not found.
// This is useful when players can be eliminated and indices shift.
func (v GameView) PlayerIdxByName(name string) int {
	for i, p := range v.game.Players {
		if p.Name == name {
			return i
		}
	}
	return -1
}

// HasLivingPlayers returns true if at least one player has lives remaining.
func (v GameView) HasLivingPlayers() bool {
	return v.game.HasLivingPlayers()
}

// --- PlayerView ---

// PlayerView wraps a Player and provides read-only access.
// All slice/map getters return fresh copies to prevent mutation of game state.
// PlayerView is a small value type (just a pointer) and should be passed by value.
type PlayerView struct {
	player *Player
}

// Name returns the player's display name.
func (pv PlayerView) Name() string {
	if pv.player == nil {
		return ""
	}
	return pv.player.Name
}

// Lives returns the player's remaining lives.
func (pv PlayerView) Lives() int {
	if pv.player == nil {
		return 0
	}
	return pv.player.Lives
}

// Hand returns a copy of the player's hand.
func (pv PlayerView) Hand() Hand {
	if pv.player == nil {
		return Hand{}
	}
	return pv.player.Hand // Already a value type [5]HandItem
}

// Stacks returns a deep copy of the player's farm stacks.
func (pv PlayerView) Stacks() Stacks {
	if pv.player == nil {
		return nil
	}
	stacks := pv.player.Farm.Stacks
	if stacks == nil {
		return nil
	}
	result := make(Stacks, len(stacks))
	for i, stack := range stacks {
		result[i] = make(Stack, len(stack))
		copy(result[i], stack)
	}
	return result
}

// NightCards returns a copy of the player's night cards.
func (pv PlayerView) NightCards() NightCards {
	if pv.player == nil {
		return nil
	}
	cards := pv.player.Farm.NightCards
	if cards == nil {
		return nil
	}
	result := make(NightCards, len(cards))
	copy(result, cards)
	return result
}
