package zcgame

import (
	"testing"
)

func TestBloodMoonEvent(t *testing.T) {
	// Create a game with 1 player
	game, err := CreateNewGame("TestPlayer")
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Put Blood Moon at the top of the night deck
	game.DebugEventsOnTop()

	// Simulate going through morning and afternoon (skip input by setting up state)
	game.Turn = Night
	game.NightNum = 1
	game.NightCardsDealt = false

	player := game.CurrentPlayer()

	// Start night turn - this should deal 1 card (Blood Moon) and show confirmation
	inputNeeded := game.DoNightTurn()

	if inputNeeded == nil {
		t.Fatal("Expected input needed for event confirmation")
	}

	if inputNeeded.Context != InputContextConfirm {
		t.Errorf("Expected InputContextConfirm, got %v", inputNeeded.Context)
	}

	// Blood Moon card should still be in NightCards (not executed yet)
	if len(player.Farm.NightCards) != 1 {
		t.Errorf("Expected 1 night card before confirmation, got %d", len(player.Farm.NightCards))
	}

	t.Logf("Message: %s", inputNeeded.Message)

	// Now confirm (press 0) - this runs the action
	inputNeeded = game.ProvideInput(0)

	// After Blood Moon runs: it adds 3 cards, original Blood Moon was removed = 3 cards
	if len(player.Farm.NightCards) != 3 {
		t.Errorf("Expected 3 night cards after Blood Moon, got %d", len(player.Farm.NightCards))
	}

	t.Logf("Player night cards after confirm: %d", len(player.Farm.NightCards))
}

func TestWinterSolsticeEvent(t *testing.T) {
	// Create a game with 1 player
	game, err := CreateNewGame("TestPlayer")
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Manually set up the night deck with Winter Solstice first
	winterSolstice := NightCard{
		Event:     NightCardEvents[3], // Winter Solstice is index 3
		ZombieKey: -1,
	}
	game.NightDeck = append([]NightCard{winterSolstice}, game.NightDeck...)

	game.Turn = Night
	game.NightNum = 1
	game.NightCardsDealt = false

	player := game.CurrentPlayer()

	inputNeeded := game.DoNightTurn()

	if inputNeeded == nil {
		t.Fatal("Expected input needed for event confirmation")
	}

	// Card should still be there before confirmation
	if len(player.Farm.NightCards) != 1 {
		t.Errorf("Expected 1 night card before confirmation, got %d", len(player.Farm.NightCards))
	}

	t.Logf("Message: %s", inputNeeded.Message)

	// Confirm to run the action
	inputNeeded = game.ProvideInput(0)

	// Winter Solstice adds 2 cards, original removed = 2
	if len(player.Farm.NightCards) != 2 {
		t.Errorf("Expected 2 night cards after Winter Solstice, got %d", len(player.Farm.NightCards))
	}

	t.Logf("Player night cards after confirm: %d", len(player.Farm.NightCards))
}

func TestTornadoEvent(t *testing.T) {
	// Create a game with 1 player
	game, err := CreateNewGame("TestPlayer")
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Manually set up the night deck with Tornado first
	tornado := NightCard{
		Event:     NightCardEvents[1], // Tornado is index 1
		ZombieKey: -1,
	}
	game.NightDeck = append([]NightCard{tornado}, game.NightDeck...)

	// Give the player some farm cards to discard
	player := game.CurrentPlayer()
	player.Farm.Stacks = Stacks{
		{HayBale, HayBale, HayBale},
		{Shotgun, Ammo},
	}

	game.Turn = Night
	game.NightNum = 1
	game.NightCardsDealt = false

	inputNeeded := game.DoNightTurn()

	if inputNeeded == nil {
		t.Fatal("Expected input needed for Tornado confirmation")
	}

	// First should be confirmation
	if inputNeeded.Context != InputContextConfirm {
		t.Errorf("Expected InputContextConfirm first, got %v", inputNeeded.Context)
	}

	t.Logf("Message: %s", inputNeeded.Message)

	// Confirm to run the action
	inputNeeded = game.ProvideInput(0)

	// Now should be asking for discard
	if inputNeeded == nil {
		t.Fatal("Expected input needed for Tornado discard")
	}

	if inputNeeded.Context != InputContextEventDiscard {
		t.Errorf("Expected InputContextEventDiscard, got %v", inputNeeded.Context)
	}

	t.Logf("Discard message: %s", inputNeeded.Message)
	t.Logf("Player farm stacks: %d", len(player.Farm.Stacks))

	// Provide input to discard first card
	inputNeeded = game.ProvideInput(1)
	if inputNeeded == nil {
		t.Fatal("Expected more input needed after first discard")
	}

	t.Logf("After discard 1 - Message: %s", inputNeeded.Message)

	// Discard second card
	inputNeeded = game.ProvideInput(1)
	if inputNeeded == nil {
		t.Fatal("Expected more input needed after second discard")
	}

	t.Logf("After discard 2 - Message: %s", inputNeeded.Message)

	// Discard third card
	inputNeeded = game.ProvideInput(1)

	// After 3 discards, should continue processing
	t.Logf("After discard 3 - inputNeeded: %v", inputNeeded)
	t.Logf("Player farm stacks after tornado: %v", player.Farm.Stacks)
}

func TestNightCardsNotRedealt(t *testing.T) {
	// Create a game with 1 player
	game, err := CreateNewGame("TestPlayer")
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Set up a simple zombie as the first night card
	game.NightDeck = []NightCard{
		{ZombieKey: 0}, // First zombie type
		{ZombieKey: 0},
		{ZombieKey: 0},
	}

	game.Turn = Night
	game.NightNum = 1
	game.NightCardsDealt = false

	player := game.CurrentPlayer()

	// First call to DoNightTurn should deal cards
	inputNeeded := game.DoNightTurn()
	cardsAfterFirstCall := len(player.Farm.NightCards)

	if cardsAfterFirstCall != 1 {
		t.Errorf("Expected 1 night card dealt, got %d", cardsAfterFirstCall)
	}

	// Simulate returning from input and calling DoNightTurn again
	// This should NOT deal more cards
	if inputNeeded != nil {
		// Provide some input to continue
		game.ProvideInput(0)
	}

	// Call DoNightTurn again (simulating what happens after input)
	game.DoNightTurn()

	cardsAfterSecondCall := len(player.Farm.NightCards)

	// Cards should not have increased (card was processed, not re-dealt)
	if cardsAfterSecondCall > cardsAfterFirstCall {
		t.Errorf("Night cards were re-dealt! Had %d, now have %d", cardsAfterFirstCall, cardsAfterSecondCall)
	}

	t.Logf("Cards after first call: %d, after second call: %d", cardsAfterFirstCall, cardsAfterSecondCall)
}
