package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/22aryja/geo-word-chain/config"
	"github.com/22aryja/geo-word-chain/handlers"
	"github.com/22aryja/geo-word-chain/internal/cities"
	"github.com/22aryja/geo-word-chain/internal/storage"
	"github.com/go-telegram/bot"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	token := config.Config("TELEGRAM_API_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_API_TOKEN is not set: add it to .env in the repository root")
	}

	// The index is read-only once built, so one copy is shared by every game.
	index, err := cities.New()
	if err != nil {
		log.Fatalf("loading cities: %v", err)
	}
	log.Printf("loaded %d cities", index.Len())

	h := handlers.New(storage.NewMemory(index))

	b, err := bot.New(token, bot.WithDefaultHandler(h.Move))
	if err != nil {
		log.Fatalf("creating bot: %v", err)
	}
	h.Register(b)
	handlers.Describe(ctx, b, index.Len())

	log.Println("bot started; press Ctrl+C to stop")
	b.Start(ctx)
}
