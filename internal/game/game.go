package game

import (
	"math"
	"math/rand/v2"
	"sort"
	"sync"

	"github.com/22aryja/geo-word-chain/internal/cities"
)

const jitterFraction = 0.05

type Game struct {
	mu         sync.Mutex
	index      *cities.Index
	used       map[string]bool
	remaining  map[string]int
	letter     string
	score      int
	moves      int
	difficulty Difficulty
}

func New(index *cities.Index, difficulty Difficulty) *Game {
	return &Game{
		index:      index,
		used:       make(map[string]bool),
		remaining:  index.LetterCounts(),
		difficulty: difficulty,
	}
}

func (g *Game) Difficulty() Difficulty {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.difficulty
}

func (g *Game) Play(input string) Move {
	g.mu.Lock()
	defer g.mu.Unlock()

	display, ok := g.index.Lookup(input)
	if !ok {
		return Move{Result: UnknownCity, NextLetter: g.letter}
	}

	name := g.index.Normalized(input)
	spelled := cities.Normalize(display.Name)
	if g.letter != "" && cities.FirstLetter(spelled) != g.letter {
		return Move{Result: WrongLetter, PlayerCity: display, NextLetter: g.letter}
	}
	if g.used[name] {
		return Move{Result: AlreadyUsed, PlayerCity: display, NextLetter: g.letter}
	}

	g.consume(name)
	g.score++
	g.moves++

	answer := g.pick(cities.LastLetter(spelled))
	if answer == "" {
		return Move{Result: BotLost, PlayerCity: display}
	}
	g.consume(answer)
	g.letter = cities.LastLetter(answer)

	return Move{
		Result:     Accepted,
		PlayerCity: display,
		BotCity:    g.index.City(answer),
		NextLetter: g.letter,
	}
}

func (g *Game) consume(name string) {
	g.used[name] = true
	if letter := cities.FirstLetter(name); letter != "" {
		g.remaining[letter]--
	}
}

func (g *Game) openness(name string) int {
	last := cities.LastLetter(name)
	n := g.remaining[last]
	if cities.FirstLetter(name) == last {
		n--
	}
	return n
}

func (g *Game) pick(letter string) string {
	bucket := g.index.ByLetter(letter)
	if len(bucket) == 0 {
		return ""
	}

	window := g.difficulty.FameWindow(g.moves)
	if window > len(bucket) {
		window = len(bucket)
	}

	candidates := g.unused(bucket[:window])
	if len(candidates) == 0 {
		if g.difficulty.ConcedesOutsideWindow() {
			return ""
		}
		candidates = g.unused(bucket)
	}
	if len(candidates) == 0 {
		return ""
	}

	sort.SliceStable(candidates, func(a, b int) bool {
		oa, ob := g.openness(candidates[a]), g.openness(candidates[b])
		if oa != ob {
			return oa > ob
		}
		return candidates[a] < candidates[b]
	})

	return candidates[g.position(len(candidates))]
}

func (g *Game) unused(names []string) []string {
	out := make([]string, 0, len(names))
	for _, name := range names {
		if !g.used[name] {
			out = append(out, name)
		}
	}
	return out
}

func (g *Game) position(n int) int {
	if n <= 1 {
		return 0
	}
	last := float64(n - 1)
	target := g.difficulty.Aggression(g.moves) * last
	jitter := float64(n) * jitterFraction

	lo := int(math.Max(0, math.Round(target-jitter)))
	hi := int(math.Min(last, math.Round(target+jitter)))
	if hi <= lo {
		return lo
	}
	return lo + rand.IntN(hi-lo+1)
}

func (g *Game) Score() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.score
}

func (g *Game) Letter() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.letter
}
