package main

import (
	_ "embed"
	"net/http"
	"strconv"
)

func (srv *server) indexHandler(w http.ResponseWriter, r *http.Request) {
	s := srv.storage
	t := srv.tmpl.GetTemplate()
	if active, err := loadActiveGames(s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if finished, err := loadFinishedGames(s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if players, err := loadPlayerRecords(s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if err = t.Execute(w, HomePage{Active: active, Finished: finished, Players: players}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (srv *server) startHandler(w http.ResponseWriter, r *http.Request) {
	s := srv.storage
	t := srv.tmpl.GetTemplate()

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gameType := r.FormValue("game")
	gamePlayers := r.Form["players"]
	gameMax, _ := strconv.Atoi(r.FormValue("max"))

	if g, err := createActiveGame(gameType, gamePlayers, gameMax, s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if err = t.ExecuteTemplate(w, g.Template(), g); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (srv *server) scoreHandler(w http.ResponseWriter, r *http.Request) {
	s := srv.storage
	t := srv.tmpl.GetTemplate()

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gameId := r.Form.Get("gameId")
	player, _ := strconv.Atoi(r.Form.Get("player"))
	incr, _ := strconv.Atoi(r.Form.Get("increment"))

	g, f, err := scoreActiveGame(gameId, player, incr, s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if f != nil {
		winner, loser, err := getPlayers(*f, s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		gameover := GameOver{
			Game:   *f,
			Winner: winner,
			Loser:  loser,
		}

		if err = t.ExecuteTemplate(w, "gameover", gameover); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err = t.ExecuteTemplate(w, g.Template(), g); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (srv *server) deleteHandler(w http.ResponseWriter, r *http.Request) {
	s := srv.storage
	t := srv.tmpl.GetTemplate()

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gameId := r.Form.Get("gameId")

	if err := deleteActiveGame(gameId, s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if games, err := loadActiveGames(s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if err = t.ExecuteTemplate(w, "active-games", games); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (srv *server) continueHandler(w http.ResponseWriter, r *http.Request) {
	s := srv.storage
	t := srv.tmpl.GetTemplate()

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gameId := r.Form.Get("gameId")

	if g, err := getActiveGame(gameId, s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if err = t.ExecuteTemplate(w, g.Template(), g); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type server struct {
	storage Storage
	tmpl    TemplateGetter
}

func newRouter(s Storage, t TemplateGetter) http.Handler {
	srv := &server{storage: s, tmpl: t}
	router := http.NewServeMux()
	router.HandleFunc("/{$}", srv.indexHandler)
	router.HandleFunc("/start", srv.startHandler)
	router.HandleFunc("/score", srv.scoreHandler)
	router.HandleFunc("/delete", srv.deleteHandler)
	router.HandleFunc("/continue", srv.continueHandler)
	router.HandleFunc("GET /favicon.svg", faviconHandler)
	return router
}

//go:embed static/favicon.svg
var favicon []byte

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(favicon)
}
