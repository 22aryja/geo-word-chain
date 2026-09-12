package game

import (
	"math/rand/v2"
	"sync"

	"github.com/22aryja/geo-word-chain/internal/cities"
)

type Game struct {
	mu     sync.Mutex
	index  *cities.Index
	used   map[string]bool
	letter string
	score  int
}

func New(index *cities.Index) *Game {
	return &Game{
		index: index,
		used:  make(map[string]bool),
	}
}

func (g *Game) Play(input string) Move {
	g.mu.Lock()
	defer g.mu.Unlock()

	display, ok := g.index.Lookup(input)
	if !ok {
		return Move{Result: UnknownCity, NextLetter: g.letter}
	}

	name := cities.Normalize(input)
	if g.letter != "" && cities.FirstLetter(name) != g.letter {
		return Move{Result: WrongLetter, PlayerCity: display, NextLetter: g.letter}
	}
	if g.used[name] {
		return Move{Result: AlreadyUsed, PlayerCity: display, NextLetter: g.letter}
	}

	g.used[name] = true
	g.score++

	answer := g.pick(cities.LastLetter(name))
	if answer == "" {
		return Move{Result: BotLost, PlayerCity: display}
	}
	g.used[answer] = true
	g.letter = cities.LastLetter(answer)

	return Move{
		Result:     Accepted,
		PlayerCity: display,
		BotCity:    g.index.Display(answer),
		NextLetter: g.letter,
	}
}

func (g *Game) Score() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.score
}

func (g *Game) pick(letter string) string {
	candidates := g.index.ByLetter(letter)
	if len(candidates) == 0 {
		return ""
	}

	start := rand.IntN(len(candidates))
	for i := range candidates {
		if c := candidates[(start+i)%len(candidates)]; !g.used[c] {
			return c
		}
	}
	return ""
}
