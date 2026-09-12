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

const helpText = "Игра «Города».\n\n" +
	"Называй город — я отвечу городом на его последнюю букву. " +
	"Буквы ь, ъ и ы пропускаются. Повторяться нельзя.\n\n" +
	"/play — начать заново\n" +
	"/stop — закончить\n" +
	"/help — эти правила"

type Handler struct {
	store storage.Store
}

func New(store storage.Store) *Handler {
	return &Handler{store: store}
}

// Register wires every command handler onto the bot. The gameplay handler is
// registered separately as the default, since it answers anything that is not
// a command.
func (h *Handler) Register(b *bot.Bot) {
	// MatchTypePrefix rather than MatchTypeCommand: in groups Telegram delivers
	// "/play@YourBot", which the command matcher would not match.
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypePrefix, h.Play)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/play", bot.MatchTypePrefix, h.Play)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/stop", bot.MatchTypePrefix, h.Stop)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypePrefix, h.Help)
}

// Play starts a new game, discarding any chain already in progress.
func (h *Handler) Play(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	h.store.Start(update.Message.Chat.ID)
	h.send(ctx, b, update.Message.Chat.ID, helpText+"\n\nТвой ход — назови любой город.")
}

// Stop ends the game in this chat.
func (h *Handler) Stop(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	chatID := update.Message.Chat.ID

	text := "Сейчас нет игры. /play — начать."
	if g, ok := h.store.Get(chatID); ok {
		text = fmt.Sprintf("Игра окончена. Твой счёт: %d.\n\n/play — начать заново.", g.Score())
		h.store.Stop(chatID)
	}
	h.send(ctx, b, chatID, text)
}

func (h *Handler) Help(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	h.send(ctx, b, update.Message.Chat.ID, helpText)
}

// Move handles every non-command message: in a running game it is the player's
// turn.
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
			h.send(ctx, b, chatID, "Сейчас нет игры. /play — начать.")
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
		return fmt.Sprintf("%s — сдаюсь, у меня нет ответа. Ты победил!\nСчёт: %d.\n\n/play — ещё раз.",
			move.PlayerCity, score)

	default:
		return "Что-то пошло не так."
	}
}

func (h *Handler) send(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
}
