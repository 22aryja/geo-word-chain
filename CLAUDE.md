# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Telegram bot (Go 1.26, module `github.com/22aryja/geo-word-chain`) for a geography word-chain game. Very early stage: the game logic does not exist yet, and the bot currently just echoes messages back.

## Commands

```bash
go build ./...        # build (fastest check that the tree compiles)
go run .              # run the bot (long-polls Telegram; Ctrl+C to stop)
go vet ./...          # vet
go test ./...         # tests (none exist yet)
go test -run TestName ./pkg   # single test
```

## Configuration

`TELEGRAM_API_TOKEN` is read from a git-ignored `.env` at the repo root via [config/config.go](config/config.go). `config.Config(key)` calls `godotenv.Load(".env")` on **every** call and only prints (does not fail) when the file is missing — so the bot starts with an empty token and fails inside `bot.New`. Because the path is relative, the process must be started from the repo root.

## Architecture

- [main.go](main.go) — wires `bot.Option`s and starts the long-poll loop under a `signal.NotifyContext` (SIGINT-cancellable). `handler` is the default fallback handler.
- [handlers/](handlers/) — one file per handler; handlers are `bot.HandlerFunc` (`func(ctx, *bot.Bot, *models.Update)`) that reply through `b.SendMessage`.
- [config/](config/) — env access.

Telegram integration is [github.com/go-telegram/bot](https://github.com/go-telegram/bot) (not telebot/telegram-bot-api); check that library's API when adding handlers, keyboards, or update types.

## Known broken state

[handlers/message-text-handler.go](handlers/message-text-handler.go) does not compile (`missing return`). It was written with the signature of the *option constructor* `bot.WithMessageTextHandler(pattern, matchType, handler)` rather than of a handler. `main.go:20` passes it as the single argument to `bot.WithMessageTextHandler`, which also does not match. Fix by making it a `bot.HandlerFunc` and calling `bot.WithMessageTextHandler("/start", bot.MatchTypeExact, handlers.MessageTextHandler)` (or the pattern the game needs).
