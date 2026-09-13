package handlers

import (
	"context"
	"strings"

	"github.com/22aryja/geo-word-chain/internal/game"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const difficultyPrefix = "diff:"

const chooseDifficultyText = "Выберите сложность:"

func difficultyKeyboard() *models.InlineKeyboardMarkup {
	levels := game.Difficulties()
	rows := make([][]models.InlineKeyboardButton, 0, len(levels))
	for _, d := range levels {
		rows = append(rows, []models.InlineKeyboardButton{{
			Text:         d.String(),
			CallbackData: difficultyPrefix + d.Key(),
		}})
	}
	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func (h *Handler) AskDifficulty(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        chooseDifficultyText,
		ReplyMarkup: difficultyKeyboard(),
	})
}

func (h *Handler) ChooseDifficulty(ctx context.Context, b *bot.Bot, update *models.Update) {
	query := update.CallbackQuery
	if query == nil {
		return
	}

	difficulty, ok := game.ParseDifficulty(strings.TrimPrefix(query.Data, difficultyPrefix))

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: query.ID,
	})

	message := query.Message.Message
	if message == nil {
		return
	}
	chatID := message.Chat.ID

	if !ok {
		h.send(ctx, b, chatID, "Не понял уровень. "+noGameText)
		return
	}

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: message.ID,
		Text:      chooseDifficultyText + " " + difficulty.String(),
	})

	h.store.Start(chatID, difficulty)
	h.send(ctx, b, chatID, startText(difficulty))
}

func startText(d game.Difficulty) string {
	return d.String() + " — " + d.Description() + ".\n\n" +
		helpText + "\nТвой ход — назови любой город."
}
