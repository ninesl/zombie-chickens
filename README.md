# Zombie Chickens

A CLI card game where players defend their farms from waves of zombie chickens.

## Requirements

- Go 1.21 or later
- No external dependencies (pure standard library)

## Run the Game

### CLI Mode

```bash
go install github.com/ninesl/zombie-chickens@latest
zombie-chickens PlayerName1 PlayerName2
```

```bash
go run . player1 player2              # CLI with player names
go run . -debug player1 player2       # Debug mode (events on top of night deck)
```

The CLI requires at least one player name (1-4 players supported).

### Web Mode

```bash
go run . -web                         # Web server at http://localhost:8080
go run . -web -debug                  # Web server with debug mode
```

The web version provides a browser-based UI with real-time updates via SSE.

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


OUTPUT:

```
Zombies Killed: 0 | Events Played: 0 | Day Cards Discarded: 2
Afternoon 1
{ Ammo*, Ammo* }
---
Lance : 5hp

Farm:
{ Shotgun, Ammo* }
{ Shield* }
Hand: { 1:Hay Bale, 2:Scarecrow, 3:Ammo*, 4:Flamethrower }
---
Play 2 cards to your farm
Lance's Afternoon
1-4 in your hand to play:4
```

```
Zombies Killed: 0 | Events Played: 0 | Day Cards Discarded: 2
Night 1
{ Shield*, Ammo* }
---
Lance : 5hp
NightCard x 0
Climber
| Climbing | Fireproof | Exploding |
Farm:
1:{ Shotgun, Ammo* }
2:{ Shield* }
3:{ Flamethrower }
Hand: { Hay Bale, Scarecrow, Ammo*, Ammo*, Ammo* }
---
Progress through the night...
Lance's Night
Lance: choose stack to use or -1 to take life (stacks: [1]):1
```

