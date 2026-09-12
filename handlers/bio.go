package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func Describe(ctx context.Context, b *bot.Bot, cityCount int) {
	if _, err := b.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{Command: "play", Description: "Начать игру"},
			{Command: "score", Description: "Текущий счёт"},
			{Command: "stop", Description: "Закончить"},
			{Command: "help", Description: "Правила"},
		},
	}); err != nil {
		log.Printf("setting command menu: %v", err)
	}

	if _, err := b.SetMyName(ctx, &bot.SetMyNameParams{
		Name: "word chain",
	}); err != nil {
		log.Printf("setting name: %v", err)
	}

	description := fmt.Sprintf("Игра «Города» 🌍\n\n"+
		"Называй город — я отвечу городом на его последнюю букву. "+
		"Буквы ь, ъ и ы пропускаются, повторяться нельзя.\n\n"+
		"%d городов мира. Жми START и ходи первым.", cityCount)

	if _, err := b.SetMyDescription(ctx, &bot.SetMyDescriptionParams{
		Description: description,
	}); err != nil {
		log.Printf("setting description: %v", err)
	}

	if _, err := b.SetMyShortDescription(ctx, &bot.SetMyShortDescriptionParams{
		ShortDescription: "cities only",
	}); err != nil {
		log.Printf("setting short description: %v", err)
	}
}
