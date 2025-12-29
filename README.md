# Zombie Chickens

A CLI card game where players defend their farms from waves of zombie chickens.

## Requirements

- Go 1.21 or later
- No external dependencies (pure standard library)

## Installation

### Option 1: Go Install (Recommended)

```bash
go install github.com/ninesl/zombie-chickens@latest
zombie-chickens Alice Bob
```

### Option 2: Go Run (No Installation)

```bash
go run github.com/ninesl/zombie-chickens@latest Alice Bob
```

### Option 3: Build from Source

```bash
git clone https://github.com/ninesl/zombie-chickens.git
cd zombie-chickens
go build .
./zombie-chickens Alice Bob
```

## CLI Usage

The CLI requires at least one player name (1-4 players supported).

```bash
# Single player
zombie-chickens Alice

# Multiple players
zombie-chickens Alice Bob

# Debug mode (events appear on top of night deck)
zombie-chickens -debug Alice Bob
```

## How to Play

Zombie Chickens follows a day/night cycle:

**Day Phase** - Draw cards and build defenses on your farm. Stack items strategically to create powerful combinations like loaded shotguns, hay walls, and booby traps.

**Night Phase** - Zombie chickens attack! Use your prepared defenses to fight them off. Each zombie has unique traits that may overcome certain defenses.

Survive the zombie onslaught longer than your opponents to win.

## API Usage

The `zcgame` package provides a public API for integrating the game into other applications.

```go
import "github.com/ninesl/zombie-chickens/zcgame"
```

### Creating a Game

```go
game, err := zcgame.CreateNewGame("Alice", "Bob")
if err != nil {
    log.Fatal(err)
}
```

`CreateNewGame(playerNames ...string) (GameView, error)` creates a new game with 1-4 players.

### Game Loop Pattern

```go
game, err := zcgame.CreateNewGame("Alice", "Bob")
if err != nil {
    log.Fatal(err)
}

for {
    gameContinues, inputNeeded := game.ContinueDay()
    
    for inputNeeded != nil {
        choice := getPlayerInput(inputNeeded) // Your input handling
        gameContinues, inputNeeded = game.ContinueAfterInput(choice)
    }
    
    if !gameContinues {
        break // Game over
    }
}
```

### GameView

`GameView` is a value type (pass by value). All accessors return copies to prevent mutation.

**State Machine Control:**

| Method | Description |
|--------|-------------|
| `ContinueDay() (bool, *PlayerInputNeeded)` | Advances game state; returns (gameContinues, inputNeeded) |
| `ContinueAfterInput(choice int) (bool, *PlayerInputNeeded)` | Resumes game after player provides input |
| `DebugEventsOnTop()` | Moves events to top of night deck for testing |

**Read-Only Accessors:**

| Method | Returns | Description |
|--------|---------|-------------|
| `Turn()` | `Turn` | Current turn phase (Morning, Afternoon, Night) |
| `NightNum()` | `int` | Current night number |
| `StageInTurn()` | `StageInTurn` | Current stage (OptionalDiscard, Play2Cards, Draw2Cards, Nighttime) |
| `CurrentPlayerIdx()` | `int` | Index of current player |
| `PublicDayCards()` | `PublicDayCards` | Two face-up cards available for drawing |
| `DayDeckCount()` | `int` | Cards remaining in day deck |
| `NightDeckCount()` | `int` | Cards remaining in night deck |
| `DiscardedDayCards()` | `map[FarmItemType]int` | Discarded day cards by type |
| `DiscardedNightCards()` | `NightCards` | Discarded night cards |
| `PlayerCount()` | `int` | Number of active players |
| `Players()` | `[]PlayerView` | All players as PlayerView wrappers |
| `Player(idx int)` | `PlayerView` | Single player by index |
| `CurrentPlayer()` | `PlayerView` | Current player |
| `HasLivingPlayers()` | `bool` | True if any player has lives remaining |

### PlayerView

`PlayerView` is a value type (pass by value). All accessors return copies.

| Method | Returns | Description |
|--------|---------|-------------|
| `Name()` | `string` | Player's display name |
| `Lives()` | `int` | Remaining lives |
| `Hand()` | `Hand` | Copy of player's hand (5 cards) |
| `Stacks()` | `Stacks` | Deep copy of farm stacks |
| `NightCards()` | `NightCards` | Copy of pending night cards |

### PlayerInputNeeded

Returned when the game requires player input.

| Field | Type | Description |
|-------|------|-------------|
| `Message` | `string` | Prompt to display |
| `ValidChoices` | `[]int` | Valid input options |
| `RenderType` | `RenderType` | How to render game state |

### CLI Mode Functions

These functions support terminal rendering with ANSI colors.

| Function | Description |
|----------|-------------|
| `SetCLIMode(enabled bool)` | Enable/disable ANSI color output |
| `IsCLIMode() bool` | Check if CLI mode is enabled |
| `RefreshRender(v GameView)` | Clear screen and print game state |
| `GatherCLIInput(v GameView, inputNeeded *PlayerInputNeeded) int` | Read validated input from stdin |
