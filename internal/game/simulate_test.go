package game

import (
	"testing"

	"github.com/22aryja/geo-word-chain/internal/cities"
)

func TestSimulateOpeningMoves(t *testing.T) {
	idx := index(t)
	for _, d := range Difficulties() {
		g := New(idx, d)
		m := g.Play("Москва")
		if m.Result != Accepted {
			t.Fatalf("%s: %v", d.Key(), m.Result)
		}
		t.Logf("%-12s first answer: %-20s pop=%-9d leaves %d options on %q",
			d.Key(), m.BotCity.Name, m.BotCity.Population,
			len(idx.ByLetter(m.NextLetter)), m.NextLetter)
	}
}

func TestSimulateSqueeze(t *testing.T) {
	idx := index(t)
	for _, d := range Difficulties() {
		g := New(idx, d)
		opts := []int{}
		name := "Москва"
		for turn := range 8 {
			m := g.Play(name)
			if m.Result != Accepted {
				t.Logf("%-12s ended at turn %d with %v", d.Key(), turn, m.Result)
				break
			}
			opts = append(opts, countUnused(g, m.NextLetter))
			next := firstUnused(g, m.NextLetter)
			if next == "" {
				t.Logf("%-12s player stuck at turn %d", d.Key(), turn)
				break
			}
			name = next
		}
		t.Logf("%-12s options left after each bot answer: %v", d.Key(), opts)
	}
}

func countUnused(g *Game, letter string) int {
	n := 0
	for _, c := range g.index.ByLetter(letter) {
		if !g.used[c] {
			n++
		}
	}
	return n
}

func firstUnused(g *Game, letter string) string {
	for _, c := range g.index.ByLetter(letter) {
		if !g.used[c] {
			return g.index.City(c).Name
		}
	}
	return ""
}

var _ = cities.Normalize

func TestAggressionCurve(t *testing.T) {
	for _, d := range Difficulties() {
		row := []float64{}
		for _, move := range []int{0, 1, 2, 3, 5, 10, 20} {
			row = append(row, d.Aggression(move))
		}
		t.Logf("%-12s aggression at moves 0,1,2,3,5,10,20: %.2f", d.Key(), row)

		if got := d.Aggression(0); got > 0.25 {
			t.Errorf("%s opens at aggression %.2f; every level should open gently", d.Key(), got)
		}
		if got := d.Aggression(1000); got > 1 || got < 0 {
			t.Errorf("%s: aggression %.2f out of range", d.Key(), got)
		}
	}

	if Simple.Aggression(100) != 0 {
		t.Error("simple must never squeeze")
	}
	if AI.Aggression(5) != 1 {
		t.Error("ai must be fully aggressive by move 5")
	}
}

func TestEasyConcedesEventually(t *testing.T) {
	if !Simple.ConcedesOutsideWindow() {
		t.Error("simple must give up rather than reach for obscure cities")
	}
	if AI.ConcedesOutsideWindow() {
		t.Error("ai must use the whole dictionary before conceding")
	}
}
