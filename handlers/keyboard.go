package handlers

import "github.com/go-telegram/bot/models"

// Reply-keyboard buttons send their label as an ordinary message, so each label
// is also the pattern its handler is registered under. They live in constants
// so the two can never drift apart.
const (
	btnPlay  = "🎮 Новая игра"
	btnScore = "🏆 Счёт"
	btnRules = "📖 Правила"
	btnStop  = "🛑 Стоп"
)

// mainKeyboard is the button panel shown under the input field. The outer slice
// is rows, so this renders as a 2x2 grid.
var mainKeyboard = &models.ReplyKeyboardMarkup{
	Keyboard: [][]models.KeyboardButton{
		{{Text: btnPlay}, {Text: btnScore}},
		{{Text: btnRules}, {Text: btnStop}},
	},
	ResizeKeyboard:        true, // fit the buttons instead of filling the screen
	IsPersistent:          true, // keep the panel open between messages
	InputFieldPlaceholder: "Назови город…",
}
