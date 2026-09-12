package handlers

import "github.com/go-telegram/bot/models"

const (
	btnPlay  = "🎮 Новая игра"
	btnScore = "🏆 Счёт"
	btnRules = "📖 Правила"
	btnStop  = "🛑 Стоп"
)

var mainKeyboard = &models.ReplyKeyboardMarkup{
	Keyboard: [][]models.KeyboardButton{
		{{Text: btnPlay}, {Text: btnScore}},
		{{Text: btnRules}, {Text: btnStop}},
	},
	ResizeKeyboard:        true,
	IsPersistent:          true,
	InputFieldPlaceholder: "Назови город…",
}
