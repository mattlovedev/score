package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

const maxNameLength = 40

var (
	// dominoesMaxScores matches the options on the start form.
	dominoesMaxScores = []int{100, 150, 200, 250}
	// maxIncrement is the largest button on each game's scoring screen.
	maxIncrement = map[string]int{
		gameCribbage: 15,
		gameDominoes: 30,
	}
)

// badRequestError is a validation failure; its message is safe to return to the
// client with a 400.
type badRequestError struct {
	msg string
}

func (e *badRequestError) Error() string {
	return e.msg
}

func badRequest(format string, args ...any) error {
	return &badRequestError{msg: fmt.Sprintf(format, args...)}
}

func isBadRequest(err error) bool {
	var br *badRequestError
	return errors.As(err, &br)
}

// validateNewGame checks a /start request. Names should already be trimmed.
func validateNewGame(gameType string, players []string, max int) error {
	switch gameType {
	case gameCribbage:
	case gameDominoes:
		if !slices.Contains(dominoesMaxScores, max) {
			return badRequest("max score must be one of %v", dominoesMaxScores)
		}
	default:
		return badRequest("unknown game %q", gameType)
	}

	// The game screens, Finished() and FinishedGame only handle two players.
	if len(players) != 2 {
		return badRequest("games need exactly 2 players")
	}
	seen := map[string]bool{}
	for _, name := range players {
		key := strings.ToLower(name)
		switch {
		case name == "":
			return badRequest("player names can't be empty")
		case utf8.RuneCountInString(name) > maxNameLength:
			return badRequest("player names can be at most %d characters", maxNameLength)
		case seen[key]:
			return badRequest("players need different names")
		}
		seen[key] = true
	}
	return nil
}

// validateScore checks a /score request against the game it applies to.
func validateScore(g ActiveGame, player int, incr int) error {
	if player < 0 || player >= len(g.Players) {
		return badRequest("no player %d in this game", player)
	}
	if limit := maxIncrement[g.Type]; incr == 0 || incr > limit || incr < -limit {
		return badRequest("increment must be between -%d and %d, and not 0", limit, limit)
	}
	return nil
}
