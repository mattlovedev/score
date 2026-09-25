package main

import (
	_ "embed"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func (srv *server) indexHandler(w http.ResponseWriter, r *http.Request) {
	s := srv.storage
	t := srv.tmpl.GetTemplate()
	if active, err := loadActiveGames(s); err != nil {
		writeError(w, r, err)
	} else if finished, err := loadFinishedGames(s); err != nil {
		writeError(w, r, err)
	} else if players, err := loadPlayerRecords(s); err != nil {
		writeError(w, r, err)
	} else if err = t.ExecuteTemplate(w, "homepage.html", HomePage{Active: active, Finished: finished, Players: players}); err != nil {
		writeError(w, r, err)
	}
}

func (srv *server) startHandler(w http.ResponseWriter, r *http.Request) {
	s := srv.storage
	t := srv.tmpl.GetTemplate()

	if err := r.ParseForm(); err != nil {
		writeError(w, r, badRequest("invalid form: %v", err))
		return
	}

	gameType := r.FormValue("game")
	gamePlayers := r.Form["players"]
	for i := range gamePlayers {
		gamePlayers[i] = strings.TrimSpace(gamePlayers[i])
	}
	gameMax, _ := strconv.Atoi(r.FormValue("max")) // only used for dominoes; validated below

	if err := validateNewGame(gameType, gamePlayers, gameMax); err != nil {
		writeError(w, r, err)
		return
	}

	if g, err := createActiveGame(gameType, gamePlayers, gameMax, s); err != nil {
		writeError(w, r, err)
	} else if err = t.ExecuteTemplate(w, "game", g); err != nil {
		writeError(w, r, err)
	}
}

func (srv *server) scoreHandler(w http.ResponseWriter, r *http.Request) {
	s := srv.storage
	t := srv.tmpl.GetTemplate()

	if err := r.ParseForm(); err != nil {
		writeError(w, r, badRequest("invalid form: %v", err))
		return
	}

	gameID := r.Form.Get("gameId")
	player, perr := strconv.Atoi(r.Form.Get("player"))
	incr, ierr := strconv.Atoi(r.Form.Get("increment"))
	if gameID == "" || perr != nil || ierr != nil {
		http.Error(w, "gameId, player and increment are required", http.StatusBadRequest)
		return
	}

	g, f, err := scoreActiveGame(gameID, player, incr, s)
	if err != nil {
		writeError(w, r, err)
		return
	}

	if f != nil {
		winner, loser, err := winnerAndLoserRecords(*f, s)
		if err != nil {
			writeError(w, r, err)
			return
		}

		gameover := GameOver{
			Game:   *f,
			Winner: winner,
			Loser:  loser,
		}

		if err = t.ExecuteTemplate(w, "gameover", gameover); err != nil {
			writeError(w, r, err)
		}
		return
	}

	if err = t.ExecuteTemplate(w, "game", g); err != nil {
		writeError(w, r, err)
	}
}

func (srv *server) deleteHandler(w http.ResponseWriter, r *http.Request) {
	s := srv.storage
	t := srv.tmpl.GetTemplate()

	if err := r.ParseForm(); err != nil {
		writeError(w, r, badRequest("invalid form: %v", err))
		return
	}

	gameID := r.Form.Get("gameId")

	if err := deleteActiveGame(gameID, s); err != nil {
		writeError(w, r, err)
	} else if games, err := loadActiveGames(s); err != nil {
		writeError(w, r, err)
	} else if err = t.ExecuteTemplate(w, "active-games", games); err != nil {
		writeError(w, r, err)
	}

}

func (srv *server) continueHandler(w http.ResponseWriter, r *http.Request) {
	s := srv.storage
	t := srv.tmpl.GetTemplate()

	if err := r.ParseForm(); err != nil {
		writeError(w, r, badRequest("invalid form: %v", err))
		return
	}

	gameID := r.Form.Get("gameId")

	if g, err := getActiveGame(gameID, s); err != nil {
		writeError(w, r, err)
	} else if err = t.ExecuteTemplate(w, "game", g); err != nil {
		writeError(w, r, err)
	}
}

// writeError sends err to the client. Validation errors and missing games get
// their own status and message; anything else is logged and hidden behind a
// generic 500, so internal details don't reach the page.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case isBadRequest(err):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case isNotFound(err):
		http.Error(w, "That game no longer exists.", http.StatusNotFound)
	default:
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
		http.Error(w, "Something went wrong, try again.", http.StatusInternalServerError)
	}
}

type server struct {
	storage Storage
	tmpl    TemplateGetter
}

func newRouter(s Storage, t TemplateGetter) http.Handler {
	srv := &server{storage: s, tmpl: t}
	router := http.NewServeMux()
	router.HandleFunc("GET /{$}", srv.indexHandler)
	router.HandleFunc("POST /start", srv.startHandler)
	router.HandleFunc("POST /score", srv.scoreHandler)
	router.HandleFunc("POST /delete", srv.deleteHandler)
	router.HandleFunc("POST /continue", srv.continueHandler)
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
