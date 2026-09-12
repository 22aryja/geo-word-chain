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
