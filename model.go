package main

import "sort"

func loadActiveGames(s Storage) ([]ActiveGame, error) {
	var games []ActiveGame
	query := NewQuery().
		Order("Date", Asc)
	if err := s.Query(activeGamesCollection, query, &games); err != nil {
		return nil, err
	}
	for g := range games {
		games[g].Date = games[g].Date.In(tz)
	}
	return games, nil
}

func loadFinishedGames(s Storage) ([]FinishedGame, error) {
	var games []FinishedGame
	query := NewQuery().
		Order("Date", Desc).
		Limit(10)
	if err := s.Query(finishedGamesCollection, query, &games); err != nil {
		return nil, err
	}
	for g := range games {
		games[g].Date = games[g].Date.In(tz)
	}
	return games, nil
}

func createActiveGame(gameType string, gamePlayers []string, max int, s Storage) (ActiveGame, error) {
	g := NewActiveGame(GenerateRandomString(16), gameType, gamePlayers, max)
	if err := s.Add(activeGamesCollection, g.Id, g); err != nil {
		return ActiveGame{}, err
	}
	return g, nil
}

func getActiveGame(id string, s Storage) (ActiveGame, error) {
	var g ActiveGame
	if err := s.Get(activeGamesCollection, id, &g); err != nil {
		return ActiveGame{}, err
	}
	g.Id = id
	return g, nil
}

func deleteActiveGame(id string, s Storage) error {
	return s.Delete(activeGamesCollection, id)
}

// scoreActiveGame adds incr to a player's score in one transaction, so taps that
// arrive together can't overwrite each other. If the score reaches the game's max,
// the game moves to finishedGames (under the same id) in that same transaction and
// is returned as finished; a second tap on an already finished game fails because
// the active game is gone, so a win can't be counted twice.
func scoreActiveGame(id string, player int, incr int, s Storage) (ActiveGame, *FinishedGame, error) {
	var g ActiveGame
	var f *FinishedGame
	err := s.RunTransaction(func(tx Tx) error {
		g, f = ActiveGame{}, nil
		if err := tx.Get(activeGamesCollection, id, &g); err != nil {
			return err
		}
		g.Id = id
		if err := validateScore(g, player, incr); err != nil {
			return err
		}
		g.Players[player].Score += incr

		if g.Players[player].Score >= g.MaxScore {
			g.Players[player].Score = g.MaxScore
			finished := g.Finished()
			f = &finished
			if err := tx.Set(finishedGamesCollection, finished.Id, finished); err != nil {
				return err
			}
			return tx.Delete(activeGamesCollection, id)
		}

		g.Date = timeNow()
		return tx.Set(activeGamesCollection, id, g)
	})
	return g, f, err
}

// loadPlayerRecords tallies every player's wins and losses from all finished games.
func loadPlayerRecords(s Storage) ([]Player, error) {
	var games []FinishedGame
	if err := s.Query(finishedGamesCollection, NewQuery(), &games); err != nil {
		return nil, err
	}
	return playerRecords(games), nil
}

// playerRecords returns each player's wins and losses in games, sorted by name.
func playerRecords(games []FinishedGame) []Player {
	byName := map[string]*Player{}
	record := func(name string) *Player {
		if byName[name] == nil {
			byName[name] = &Player{Name: name}
		}
		return byName[name]
	}
	for _, g := range games {
		winner, loser := record(g.Winner.Name), record(g.Loser.Name)
		switch g.Type {
		case gameCribbage:
			winner.Cribbage.Wins++
			loser.Cribbage.Losses++
		case gameDominoes:
			winner.Dominoes.Wins++
			loser.Dominoes.Losses++
		}
	}
	players := make([]Player, 0, len(byName))
	for _, p := range byName {
		players = append(players, *p)
	}
	sort.Slice(players, func(i, j int) bool { return players[i].Name < players[j].Name })
	return players
}

// getPlayers returns the current records of a finished game's winner and loser.
func getPlayers(f FinishedGame, s Storage) (Player, Player, error) {
	players, err := loadPlayerRecords(s)
	if err != nil {
		return Player{}, Player{}, err
	}
	var winner, loser Player
	for _, p := range players {
		switch p.Name {
		case f.Winner.Name:
			winner = p
		case f.Loser.Name:
			loser = p
		}
	}
	return winner, loser, nil
}
