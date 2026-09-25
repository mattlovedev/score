# score

A scorekeeper for cribbage and dominoes games, live at
[score.mattlove.dev](https://score.mattlove.dev).

The home page lists games in progress, the 10 most recent finished games, and each
player's wins and losses per game. Start a game by picking the game and entering the
players; dominoes also asks for a target score (100 to 250, default 250), while
cribbage is always played to 121. Each player gets `+`/`-` buttons (1, 2, 3, 4, 5,
10 and 15 for cribbage; 5 to 30 for dominoes). The first player to reach the target
wins, the game moves to the finished list, and both players' records are updated.
Unfinished games can be continued or deleted from the home page.

It's a Go server that renders HTML templates, with [htmx](https://htmx.org) swapping
in the game screens and [Bootstrap](https://getbootstrap.com) for styling.

## Running locally

```bash
go run . -local   # then open http://localhost:8080
```

`-local` reads templates from `templates/` (reloaded on every request, so template
edits show up without a restart) and stores games as JSON files under `storage/`,
so no Google Cloud access is needed.

Without `-local`, it uses Firestore and a GCS bucket for templates:

| Variable | Default |
|---|---|
| `SCORE_PROJECT` | `mattlovedev-apps` |
| `SCORE_DATABASE` | `score` (Firestore database) |
| `SCORE_TEMPLATES_BUCKET` | `mattlovedev-apps-score-templates` |
| `PORT` | `8080` |

## Deploying

Pushes to `main` are built into a container and deployed to Cloud Run (service
`score` in `mattlovedev-apps`) by
[`.github/workflows/deploy.yml`](.github/workflows/deploy.yml), which authenticates
through Workload Identity Federation, so no keys are stored.

Templates are not part of the image; the server loads them from the bucket when it
starts. After changing `templates/`, upload them **before** pushing, so the new
revision picks them up:

```bash
gcloud storage cp templates/*.html gs://mattlovedev-apps-score-templates/
```

## Code

- `main.go`: picks local or cloud storage and templates, then starts the server.
- `api.go`: routes and handlers (`/`, `/start`, `/score`, `/continue`, `/delete`,
  `/favicon.svg`).
- `model.go`, `types.go`: games and players, and moving a game from active to finished.
- `storage_*.go`: the `Storage` interface with Firestore and local JSON implementations.
- `tmpl_*.go`: loading templates from GCS or `templates/`.

Firestore holds three collections: `activeGames`, `finishedGames` and `players`
(keyed by player name).

## Known issues

- The start form offers three or four players, but the game screens and the winner
  logic only handle two.
- `/score` doesn't check the player index, so a bad request crashes the handler.
- The cribbage template file is misspelled `cribabge.html` (templates are looked up
  by their `{{define}}` name, so it still works), and `homepage-tables.html` is unused.
