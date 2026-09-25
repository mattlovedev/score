package main

import (
	"time"
)

var (
	tz, _ = time.LoadLocation("America/New_York")
)

func timeNow() time.Time {
	return time.Now().In(tz)
}

const (
	gameCribbage = "Cribbage"
	gameDominoes = "Dominoes"
)

type WinsLosses struct {
	Wins   int
	Losses int
}

type Player struct {
	Name     string
	Cribbage WinsLosses
	Dominoes WinsLosses
}

type GamePlayer struct {
	Name  string
	Score int
}

type ActiveGame struct {
	Id         string
	Date       time.Time
	Type       string
	MaxScore   int
	NumPlayers int
	Players    []GamePlayer
}

type FinishedGame struct {
	Id       string
	Date     time.Time
	Type     string
	MaxScore int
	Winner   GamePlayer
	Loser    GamePlayer
}

func NewActiveGame(id string, game string, players []string, max int) ActiveGame {
	maxScores := map[string]int{
		gameCribbage: 121,
	}
	p := make([]GamePlayer, len(players))
	for i := range players {
		p[i] = GamePlayer{Name: players[i]}
	}
	g := ActiveGame{
		Id:         id,
		Date:       timeNow(),
		Type:       game,
		MaxScore:   max,
		NumPlayers: len(players),
		Players:    p,
	}
	if m := maxScores[g.Type]; m > 0 {
		g.MaxScore = m
	}
	return g
}

func (g ActiveGame) Finished() FinishedGame {
	f := FinishedGame{
		Id:       g.Id,
		Date:     timeNow(),
		Type:     g.Type,
		MaxScore: g.MaxScore,
	}
	if g.Players[0].Score > g.Players[1].Score {
		f.Winner = g.Players[0]
		f.Loser = g.Players[1]
	} else {
		f.Winner = g.Players[1]
		f.Loser = g.Players[0]
	}
	return f
}

// scoreButtons are the +/- amounts on each game's scoring screen.
var scoreButtons = map[string][]int{
	gameCribbage: {1, 2, 3, 4, 5, 10, 15},
	gameDominoes: {5, 10, 15, 20, 25, 30},
}

// Buttons returns the +/- amounts for the game's scoring screen.
func (g ActiveGame) Buttons() []int {
	return scoreButtons[g.Type]
}

type HomePage struct {
	Active   []ActiveGame
	Finished []FinishedGame
	Players  []Player
}

type GameOver struct {
	Game   FinishedGame
	Winner Player
	Loser  Player
}
