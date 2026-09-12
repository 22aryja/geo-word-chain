package game

import (
	"sync"
	"testing"

	"github.com/22aryja/geo-word-chain/internal/cities"
)

func index(t *testing.T) *cities.Index {
	t.Helper()
	idx, err := cities.New()
	if err != nil {
		t.Fatal(err)
	}
	return idx
}

func TestPlayFirstMoveChains(t *testing.T) {
	g := New(index(t))

	m := g.Play("  МОСКВА  ")
	if m.Result != Accepted {
		t.Fatalf("Result = %v, want accepted", m.Result)
	}
	if m.PlayerCity != "Москва" {
		t.Errorf("PlayerCity = %q, want canonical spelling", m.PlayerCity)
	}
	if got := cities.FirstLetter(m.BotCity); got != "а" {
		t.Errorf("bot answered %q starting with %q, want а", m.BotCity, got)
	}
	if m.NextLetter != cities.LastLetter(m.BotCity) {
		t.Errorf("NextLetter = %q, want last letter of %q", m.NextLetter, m.BotCity)
	}
	if g.Score() != 1 {
		t.Errorf("Score = %d, want 1", g.Score())
	}
}

func TestPlayRejections(t *testing.T) {
	idx := index(t)

	t.Run("unknown city", func(t *testing.T) {
		g := New(idx)
		if m := g.Play("Йцукенгшщз"); m.Result != UnknownCity {
			t.Errorf("Result = %v, want unknown city", m.Result)
		}
		if g.Score() != 0 {
			t.Error("a rejected move must not score")
		}
	})

	t.Run("wrong letter", func(t *testing.T) {
		g := New(idx)
		g.letter = "щ"
		m := g.Play("Москва")
		if m.Result != WrongLetter {
			t.Errorf("Result = %v, want wrong letter", m.Result)
		}
		if m.NextLetter != "щ" {
			t.Errorf("NextLetter = %q, want the letter still required", m.NextLetter)
		}
	})

	t.Run("already used", func(t *testing.T) {
		g := New(idx)
		g.used["москва"] = true
		g.letter = "м"
		if m := g.Play("Москва"); m.Result != AlreadyUsed {
			t.Errorf("Result = %v, want already used", m.Result)
		}
	})
}

func TestPlayBotLost(t *testing.T) {
	idx := index(t)
	g := New(idx)

	// Москва ends in "а"; leave the bot nothing to answer with.
	for _, c := range idx.ByLetter("а") {
		g.used[c] = true
	}

	m := g.Play("Москва")
	if m.Result != BotLost {
		t.Fatalf("Result = %v, want bot lost", m.Result)
	}
	if m.PlayerCity != "Москва" {
		t.Errorf("PlayerCity = %q, want the move to stand", m.PlayerCity)
	}
	if m.NextLetter != "" {
		t.Errorf("NextLetter = %q, want empty once the game is over", m.NextLetter)
	}
}

// The bot dispatches each update in its own goroutine, so Play must tolerate
// concurrent calls for the same chat. Only one of these can be accepted.
func TestPlayConcurrent(t *testing.T) {
	g := New(index(t))

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.Play("Москва")
		}()
	}
	wg.Wait()

	if got := g.Score(); got != 1 {
		t.Errorf("Score = %d, want 1: only the first Москва may count", got)
	}
}

// A city named without its hyphen must be the same city, including for the
// repeat check.
func TestPlayIgnoresSeparators(t *testing.T) {
	g := New(index(t))
	g.letter = "н"

	m := g.Play("нью йорк")
	if m.Result != Accepted {
		t.Fatalf("Result = %v, want accepted", m.Result)
	}
	if m.PlayerCity != "Нью-Йорк" {
		t.Errorf("PlayerCity = %q, want the canonical hyphenated spelling", m.PlayerCity)
	}

	g2 := New(index(t))
	g2.letter = "н"
	g2.Play("Нью-Йорк")
	g2.letter = "н"
	if m := g2.Play("ньюйорк"); m.Result != AlreadyUsed {
		t.Errorf("Result = %v, want already used: same city, different spelling", m.Result)
	}
}
