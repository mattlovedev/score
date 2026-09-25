package main

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

func updateActiveGame(g ActiveGame, s Storage) error {
	g.Date = timeNow()
	return s.Set(activeGamesCollection, g.Id, g)
}

func deleteActiveGame(id string, s Storage) error {
	return s.Delete(activeGamesCollection, id)
}

func LoadPlayers(s Storage) ([]Player, error) {
	var players []Player
	if err := s.Query(playersCollection, NewQuery(), &players); err != nil {
		return nil, err
	}
	return players, nil
}

func getPlayer(name string, s Storage) (Player, error) {
	var p Player
	if err := s.Get(playersCollection, name, &p); err != nil {
		return Player{}, err
	}
	return p, nil
}

func createPlayer(player Player, s Storage) error {
	return s.Add(playersCollection, player.Name, player)
}

func UpdatePlayer(player Player, s Storage) error {
	return s.Set(playersCollection, player.Name, player)
}

func switchActiveGameToFinished(g ActiveGame, s Storage) (FinishedGame, error) {
	f := g.Finished()
	var err error
	var winner, loser Player
	// ideally this would be a transaction obv
	if err = s.Set(finishedGamesCollection, f.Id, f); err != nil {
		return FinishedGame{}, err
	}
	if winner, err = getPlayer(f.Winner.Name, s); err != nil {
		if f.Type == gameCribbage {
			winner = Player{Name: f.Winner.Name, Cribbage: WinsLosses{Wins: 1, Losses: 0}}
		} else if f.Type == gameDominoes {
			winner = Player{Name: f.Winner.Name, Dominoes: WinsLosses{Wins: 1, Losses: 0}}
		}
		if err = createPlayer(winner, s); err != nil {
			return FinishedGame{}, err
		}
	} else {
		if f.Type == gameCribbage {
			winner.Cribbage.Wins++
		} else if f.Type == gameDominoes {
			winner.Dominoes.Wins++
		}
		if err = UpdatePlayer(winner, s); err != nil {
			return FinishedGame{}, err
		}
	}
	if loser, err = getPlayer(f.Loser.Name, s); err != nil {
		if f.Type == gameCribbage {
			loser = Player{Name: f.Loser.Name, Cribbage: WinsLosses{Wins: 0, Losses: 1}}
		} else if f.Type == gameDominoes {
			loser = Player{Name: f.Loser.Name, Dominoes: WinsLosses{Wins: 0, Losses: 1}}
		}
		if err = createPlayer(loser, s); err != nil {
			return FinishedGame{}, err
		}
	} else {
		if f.Type == gameCribbage {
			loser.Cribbage.Losses++
		} else if f.Type == gameDominoes {
			loser.Dominoes.Losses++
		}
		if err = UpdatePlayer(loser, s); err != nil {
			return FinishedGame{}, err
		}
	}
	if err = s.Delete(activeGamesCollection, g.Id); err != nil {
		return FinishedGame{}, err
	}
	return f, nil
}

func getPlayers(f FinishedGame, s Storage) (Player, Player, error) {
	var winner, loser Player
	if err := s.Get(playersCollection, f.Winner.Name, &winner); err != nil {
		return Player{}, Player{}, err
	}
	if err := s.Get(playersCollection, f.Loser.Name, &loser); err != nil {
		return Player{}, Player{}, err
	}
	return winner, loser, nil
}
