package handlers

import (
	"context"
	"fmt"
	"html"
	"strings"

	"github.com/22aryja/geo-word-chain/internal/cities"
	"github.com/22aryja/geo-word-chain/internal/game"
	"github.com/22aryja/geo-word-chain/internal/storage"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const helpText = "Игра «Города».\n\n" +
	"Называй город — я отвечу городом на его последнюю букву. " +
	"Буквы ь, ъ и ы пропускаются. Повторяться нельзя.\n"

const noGameText = "Сейчас нет игры. Нажми «" + btnPlay + "» или /play."

type Handler struct {
	store storage.Store
}

func New(store storage.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Register(b *bot.Bot) {

	for pattern, handler := range map[string]bot.HandlerFunc{
		"/start": h.AskDifficulty,
		"/play":  h.AskDifficulty,
		"/score": h.Score,
		"/stop":  h.Stop,
		"/help":  h.Help,
	} {
		b.RegisterHandler(bot.HandlerTypeMessageText, pattern, bot.MatchTypePrefix, handler)
	}

	for label, handler := range map[string]bot.HandlerFunc{
		btnPlay:  h.AskDifficulty,
		btnScore: h.Score,
		btnRules: h.Help,
		btnStop:  h.Stop,
	} {
		b.RegisterHandler(bot.HandlerTypeMessageText, label, bot.MatchTypeExact, handler)
	}

	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, difficultyPrefix, bot.MatchTypePrefix, h.ChooseDifficulty)
}

func (h *Handler) Stop(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	chatID := update.Message.Chat.ID

	text := noGameText
	if g, ok := h.store.Get(chatID); ok {
		text = fmt.Sprintf("Игра окончена. Твой счёт: %d.", g.Score())
		h.store.Stop(chatID)
	}
	h.send(ctx, b, chatID, text)
}

func (h *Handler) Score(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	chatID := update.Message.Chat.ID

	g, ok := h.store.Get(chatID)
	if !ok {
		h.send(ctx, b, chatID, noGameText)
		return
	}

	text := fmt.Sprintf("Счёт: %d.", g.Score())
	if letter := g.Letter(); letter != "" {
		text += fmt.Sprintf(" Тебе на «%s».", strings.ToUpper(letter))
	} else {
		text += " Назови любой город."
	}
	h.send(ctx, b, chatID, text)
}

func (h *Handler) Help(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	h.send(ctx, b, update.Message.Chat.ID, helpText)
}

func (h *Handler) Move(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}
	chatID := update.Message.Chat.ID

	g, ok := h.store.Get(chatID)
	if !ok {

		if update.Message.Chat.Type == models.ChatTypePrivate {
			h.send(ctx, b, chatID, noGameText)
		}
		return
	}

	move := g.Play(update.Message.Text)
	h.send(ctx, b, chatID, render(move, g.Score()))

	if move.Result == game.BotLost {
		h.store.Stop(chatID)
	}
}

func render(move game.Move, score int) string {
	switch move.Result {
	case game.Accepted:
		return fmt.Sprintf("%s\n%s%s\n\nТебе на «%s». Счёт: %d",
			label(move.PlayerCity), label(move.BotCity), fact(move.BotCity),
			strings.ToUpper(move.NextLetter), score)

	case game.UnknownCity:
		if move.NextLetter == "" {
			return "Не знаю такого города."
		}
		return fmt.Sprintf("Не знаю такого города. Нужен город на «%s».", move.NextLetter)

	case game.WrongLetter:
		return fmt.Sprintf("«%s» не подходит — нужен город на «%s».",
			html.EscapeString(move.PlayerCity.Name), move.NextLetter)

	case game.AlreadyUsed:
		return fmt.Sprintf("«%s» уже было. Нужен другой город на «%s».",
			html.EscapeString(move.PlayerCity.Name), move.NextLetter)

	case game.BotLost:
		return fmt.Sprintf("%s\n\nСдаюсь, у меня нет ответа. Ты победил!\nСчёт: %d.",
			label(move.PlayerCity), score)

	default:
		return "Что-то пошло не так."
	}
}

func label(c cities.City) string {
	name := html.EscapeString(c.Name)
	switch {
	case c.Flag() != "" && c.CountryName != "":
		return fmt.Sprintf("%s %s %s", c.Flag(), name, " ("+html.EscapeString(c.CountryName)+")")
	case c.Flag() != "":
		return c.Flag() + " " + name
	default:
		return name
	}
}

func (h *Handler) send(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: mainKeyboard,
	})
}

func fact(c cities.City) string {
	if c.Fact == "" {
		return ""
	}
	return "\n<blockquote>" + html.EscapeString(c.Fact) + "</blockquote>"
}
