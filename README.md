# Zombie Chickens

A CLI card game where players defend their farms from waves of zombie chickens.

## Requirements

- Go 1.21 or later
- No external dependencies (pure standard library)

## Installation

### Option 1: Go Install (Recommended)

```bash
go install github.com/ninesl/zombie-chickens@latest
zombie-chickens
```

### Option 2: Go Run (No Installation)

```bash
go run github.com/ninesl/zombie-chickens@latest
```

### Option 3: Build from Source

```bash
git clone https://github.com/ninesl/zombie-chickens.git
cd zombie-chickens
go build .
./zombie-chickens
```

## Usage

```bash
# Start with default player names
zombie-chickens

# Start with custom player names
zombie-chickens Alice Bob

# Debug mode (events appear on top of night deck)
zombie-chickens -debug Alice Bob
```

## How to Play

Zombie Chickens follows a day/night cycle:

**Day Phase** - Draw cards and build defenses on your farm. Stack items strategically to create powerful combinations like loaded shotguns, hay walls, and booby traps.

**Night Phase** - Zombie chickens attack! Use your prepared defenses to fight them off. Each zombie has unique traits that may overcome certain defenses.

Survive the zombie onslaught longer than your opponents to win.
