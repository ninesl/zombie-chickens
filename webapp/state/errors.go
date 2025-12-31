package state

import "errors"

var (
	ErrGameAlreadyStarted = errors.New("game has already started")
	ErrGameNotStarted     = errors.New("game has not started")
	ErrGameFull           = errors.New("game is full (max 4 players)")
	ErrGameOver           = errors.New("game is over")
	ErrNotEnoughPlayers   = errors.New("not enough players to start")
	ErrPlayerNotFound     = errors.New("player not found in game")
	ErrNotYourTurn        = errors.New("not your turn")
	ErrNoInputNeeded      = errors.New("no input needed")
	ErrInvalidChoice      = errors.New("invalid choice")
	ErrSessionNotFound    = errors.New("session not found")
)
