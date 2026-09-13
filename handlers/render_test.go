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

func TestRenderIncludesFact(t *testing.T) {
	got := render(game.Move{
		Result:     game.Accepted,
		PlayerCity: cities.City{Name: "Санкт-Петербург", Country: "RU", CountryName: "Россия"},
		BotCity: cities.City{
			Name:        "Гамбург",
			Country:     "DE",
			CountryName: "Германия",
			Fact:        "Гамбург — город на севере Германии.",
		},
		NextLetter: "г",
	}, 1)

	if !strings.Contains(got, "Гамбург — город на севере Германии.") {
		t.Errorf("fact missing from output:\n%s", got)
	}
	t.Logf("with fact:\n%s", got)
}

func TestRenderWithoutFact(t *testing.T) {
	got := render(game.Move{
		Result:     game.Accepted,
		PlayerCity: cities.City{Name: "Москва", Country: "RU", CountryName: "Россия"},
		BotCity:    cities.City{Name: "Абаза", Country: "RU", CountryName: "Россия"},
		NextLetter: "а",
	}, 1)

	if strings.Contains(got, "\n\n\n") {
		t.Errorf("blank line left where the fact would be:\n%q", got)
	}
	t.Logf("without fact:\n%s", got)
}

func TestRenderFactIsBlockquote(t *testing.T) {
	got := render(game.Move{
		Result:     game.Accepted,
		PlayerCity: cities.City{Name: "Астана", Country: "KZ", CountryName: "Казахстан"},
		BotCity: cities.City{
			Name:        "Абуджа",
			Country:     "NG",
			CountryName: "Нигерия",
			Fact:        "Абуджа — столица Нигерии с 12 декабря 1991 года.",
		},
		NextLetter: "а",
	}, 3)

	if !strings.Contains(got, "<blockquote>Абуджа — столица Нигерии с 12 декабря 1991 года.</blockquote>") {
		t.Errorf("fact is not wrapped in a blockquote:\n%s", got)
	}
	if strings.Count(got, "<blockquote>") != strings.Count(got, "</blockquote>") {
		t.Errorf("unbalanced blockquote tags:\n%s", got)
	}
	t.Logf("%s", got)
}

func TestRenderEscapesHTML(t *testing.T) {
	got := render(game.Move{
		Result:     game.Accepted,
		PlayerCity: cities.City{Name: "A<b>&", Country: "RU", CountryName: "Рос&сия"},
		BotCity:    cities.City{Name: "X>Y", Country: "RU", CountryName: "Россия", Fact: "5 < 6 & 7 > 2"},
		NextLetter: "а",
	}, 1)

	for _, raw := range []string{"A<b>", "5 < 6", "7 > 2", "Рос&сия"} {
		if strings.Contains(got, raw) {
			t.Errorf("unescaped %q would break HTML parse mode:\n%s", raw, got)
		}
	}
	if !strings.Contains(got, "&lt;") || !strings.Contains(got, "&amp;") {
		t.Errorf("expected escaped entities:\n%s", got)
	}
}

func TestRenderRejectionsEscape(t *testing.T) {
	for _, result := range []game.Result{game.WrongLetter, game.AlreadyUsed} {
		got := render(game.Move{
			Result:     result,
			PlayerCity: cities.City{Name: "<script>"},
			NextLetter: "а",
		}, 0)
		if strings.Contains(got, "<script>") {
			t.Errorf("%v leaves raw HTML:\n%s", result, got)
		}
	}
}
