# score

A scorekeeper for cribbage and dominoes games, live at
[score.mattlove.dev](https://score.mattlove.dev).

The home page lists games in progress, the 10 most recent finished games, and each
player's wins and losses per game. Start a game by picking the game and entering two
players; dominoes also asks for a target score (100 to 250, default 250), while
cribbage is always played to 121. Each player gets `+`/`-` buttons (1, 2, 3, 4, 5,
10 and 15 for cribbage; 5 to 30 for dominoes), and scores can't go below 0. The
first player to reach the target wins: a tap that would end the game asks for
confirmation first, then the game moves to the finished list. Unfinished games can
be continued or deleted (after a confirmation) from the home page.

It's a Go server that renders HTML templates, with [htmx](https://htmx.org) swapping
in the game screens and [Bootstrap](https://getbootstrap.com) for styling.

## Running locally

```bash
go run . -local   # then open http://localhost:8080
```

`-local` reads templates from `templates/` (reloaded on every request, so template
edits show up without a restart) and stores games as JSON files under `storage/`,
so no Google Cloud access is needed.

Without `-local`, it uses Firestore and the templates embedded in the binary:

| Variable | Default |
|---|---|
| `SCORE_PROJECT` | `mattlovedev-apps` |
| `SCORE_DATABASE` | `score` (Firestore database) |
| `PORT` | `8080` |

## Deploying

Pushes to `main` are built into a container and deployed to Cloud Run (service
`score` in `mattlovedev-apps`, at most 2 instances) by
[`.github/workflows/deploy.yml`](.github/workflows/deploy.yml), which authenticates
through Workload Identity Federation, so no keys are stored. Templates and the
favicon are embedded in the binary, so a push deploys everything. Pushes that only
change Markdown files don't deploy.

## Code

- `main.go`: picks local or cloud storage and templates, then starts the server.
- `api.go`: routes and handlers. `GET /` and `/favicon.svg`; `POST /start`,
  `/score`, `/continue` and `/delete`. Bad input gets a 400, a missing game a 404,
  and other errors are logged and returned as a generic 500; the page shows these in
  an alert.
- `validate.go`: input checks for starting and scoring games.
- `model.go`, `types.go`: games and players. Scoring runs in a transaction, which
  also moves a finished game from active to finished.
- `storage.go`, `storage_firestore.go`, `storage_local.go`: the `Storage` interface
  with Firestore and local JSON implementations.
- `tmpl.go`, `tmpl_embed.go`, `tmpl_local.go`: parsing the embedded templates, or
  `templates/` in local mode.
- `templates/`: the home page and its tables, the scoring screen (`game.html`) and
  the game-over screen.

Firestore holds two collections, `activeGames` and `finishedGames`. Player records
aren't stored; they're tallied from `finishedGames`.

## Future work

- Games with three or four players. The start form has the buttons, disabled for
  now; the scoring screen, the winner logic, how finished games are stored and the
  records all assume two players.
- Replace the generic, reflection-based storage layer (`Storage` with `Query`,
  `Where`, `Count`) with methods specific to what the app needs.
