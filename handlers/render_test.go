package handlers

import (
	"strings"
	"testing"

	"github.com/22aryja/geo-word-chain/internal/cities"
	"github.com/22aryja/geo-word-chain/internal/game"
)

func TestRenderAccepted(t *testing.T) {
	got := render(game.Move{
		Result:     game.Accepted,
		PlayerCity: cities.City{Name: "Москва", Country: "RU", CountryName: "Россия"},
		BotCity:    cities.City{Name: "Астана", Country: "KZ", CountryName: "Казахстан"},
		NextLetter: "а",
	}, 1)

	for _, want := range []string{"🇷🇺", "Москва", "Россия", "🇰🇿", "Астана", "Казахстан", "«А»", "Счёт: 1"} {
		if !strings.Contains(got, want) {
			t.Errorf("render output missing %q:\n%s", want, got)
		}
	}
	t.Logf("accepted:\n%s", got)
}

func TestRenderWithoutCountry(t *testing.T) {
	got := label(cities.City{Name: "Неизвестск"})
	if got != "Неизвестск" {
		t.Errorf("label() = %q, want the bare name", got)
	}
}

func TestRenderBotLost(t *testing.T) {
	got := render(game.Move{
		Result:     game.BotLost,
		PlayerCity: cities.City{Name: "Щёлково", Country: "RU", CountryName: "Россия"},
	}, 7)
	if !strings.Contains(got, "Щёлково") || !strings.Contains(got, "7") {
		t.Errorf("unexpected output:\n%s", got)
	}
	t.Logf("bot lost:\n%s", got)
}

func TestDifficultyKeyboard(t *testing.T) {
	kb := difficultyKeyboard()
	if len(kb.InlineKeyboard) != 5 {
		t.Fatalf("got %d rows, want 5", len(kb.InlineKeyboard))
	}
	seen := map[string]bool{}
	for _, row := range kb.InlineKeyboard {
		if len(row) != 1 {
			t.Fatalf("row has %d buttons, want 1", len(row))
		}
		btn := row[0]
		if !strings.HasPrefix(btn.CallbackData, difficultyPrefix) {
			t.Errorf("CallbackData %q missing prefix", btn.CallbackData)
		}
		key := strings.TrimPrefix(btn.CallbackData, difficultyPrefix)
		if _, ok := game.ParseDifficulty(key); !ok {
			t.Errorf("CallbackData %q does not parse back to a difficulty", btn.CallbackData)
		}
		if seen[btn.CallbackData] {
			t.Errorf("duplicate CallbackData %q", btn.CallbackData)
		}
		seen[btn.CallbackData] = true
		t.Logf("%-22s -> %s", btn.Text, btn.CallbackData)
	}
}

func TestCallbackDataFitsTelegramLimit(t *testing.T) {
	for _, d := range game.Difficulties() {
		if n := len(difficultyPrefix + d.Key()); n > 64 {
			t.Errorf("%s: callback data is %d bytes, Telegram allows 64", d.Key(), n)
		}
	}
}
