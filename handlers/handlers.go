// Package handlers adapts Telegram updates to the game. It is the only package
// that knows about either Telegram or the wording shown to players.
package handlers

import (
	"context"
	"fmt"

	"github.com/22aryja/geo-word-chain/internal/game"
	"github.com/22aryja/geo-word-chain/internal/storage"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const helpText = "Игра «Города»\n\n" +
	"Называй город — я отвечу городом на его последнюю букву. " +
	"Буквы ь, ъ и ы пропускаются. Повторяться нельзя.\n"

const noGameText = "Сейчас нет игры. Нажми «" + btnPlay + "» или /play."

type Handler struct {
	store storage.Store
}

func New(store storage.Store) *Handler {
	return &Handler{store: store}
}

// Register wires the commands and the reply-keyboard buttons onto the bot. The
// gameplay handler is registered separately as the default, since it answers
// anything that is not one of these.
func (h *Handler) Register(b *bot.Bot) {
	// MatchTypePrefix rather than MatchTypeCommand: in groups Telegram delivers
	// "/play@YourBot", which the command matcher would not match.
	for pattern, handler := range map[string]bot.HandlerFunc{
		"/start": h.Play,
		"/play":  h.Play,
		"/score": h.Score,
		"/stop":  h.Stop,
		"/help":  h.Help,
	} {
		b.RegisterHandler(bot.HandlerTypeMessageText, pattern, bot.MatchTypePrefix, handler)
	}

	// Button labels arrive as plain messages, so they need exact-match handlers
	// of their own. Without these they would reach Move and be judged as city
	// guesses.
	for label, handler := range map[string]bot.HandlerFunc{
		btnPlay:  h.Play,
		btnScore: h.Score,
		btnRules: h.Help,
		btnStop:  h.Stop,
	} {
		b.RegisterHandler(bot.HandlerTypeMessageText, label, bot.MatchTypeExact, handler)
	}
}

// Play starts a new game, discarding any chain already in progress.
func (h *Handler) Play(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	h.store.Start(update.Message.Chat.ID)
	h.send(ctx, b, update.Message.Chat.ID, helpText+"\nТвой ход — назови любой город.")
}

// Stop ends the game in this chat.
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

// Score reports the state of the chain without changing it.
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
		text += fmt.Sprintf(" Тебе на «%s».", letter)
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

// Move handles every other message: in a running game it is the player's turn.
func (h *Handler) Move(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}
	chatID := update.Message.Chat.ID

	g, ok := h.store.Get(chatID)
	if !ok {
		// Staying quiet in groups matters: the bot would otherwise answer every
		// unrelated message in the room.
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

// render turns a move into the message the player sees. It is the only place
// player-facing wording lives.
func render(move game.Move, score int) string {
	switch move.Result {
	case game.Accepted:
		return fmt.Sprintf("%s → %s\n\nТебе на «%s». Счёт: %d",
			move.PlayerCity, move.BotCity, move.NextLetter, score)

	case game.UnknownCity:
		if move.NextLetter == "" {
			return "Не знаю такого города."
		}
		return fmt.Sprintf("Не знаю такого города. Нужен город на «%s».", move.NextLetter)

	case game.WrongLetter:
		return fmt.Sprintf("«%s» не подходит — нужен город на «%s».",
			move.PlayerCity, move.NextLetter)

	case game.AlreadyUsed:
		return fmt.Sprintf("«%s» уже было. Нужен другой город на «%s».",
			move.PlayerCity, move.NextLetter)

	case game.BotLost:
		return fmt.Sprintf("%s — сдаюсь, у меня нет ответа. Ты победил!\nСчёт: %d.",
			move.PlayerCity, score)

	default:
		return "Что-то пошло не так."
	}
}

// send always re-attaches the keyboard so the panel stays visible.
func (h *Handler) send(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: mainKeyboard,
	})
}
