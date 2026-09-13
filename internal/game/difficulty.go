package game

type Difficulty int

const (
	Simple Difficulty = iota
	Solid
	Complicated
	Sadistic
	AI
)

var difficulties = []Difficulty{Simple, Solid, Complicated, Sadistic, AI}

func Difficulties() []Difficulty {
	out := make([]Difficulty, len(difficulties))
	copy(out, difficulties)
	return out
}

func (d Difficulty) Key() string {
	switch d {
	case Simple:
		return "simple"
	case Solid:
		return "solid"
	case Complicated:
		return "complicated"
	case Sadistic:
		return "sadistic"
	case AI:
		return "ai"
	default:
		return "unknown"
	}
}

func (d Difficulty) Title() string {
	switch d {
	case Simple:
		return "Simple"
	case Solid:
		return "Solid"
	case Complicated:
		return "Complicated"
	case Sadistic:
		return "Sadistic"
	case AI:
		return "AI"
	default:
		return "Unknown"
	}
}

func (d Difficulty) Emoji() string {
	switch d {
	case Simple:
		return "🤓"
	case Solid:
		return "😎"
	case Complicated:
		return "😫"
	case Sadistic:
		return "🤕"
	case AI:
		return "🤖"
	default:
		return "🤷"
	}
}

func (d Difficulty) Description() string {
	switch d {
	case Simple:
		return "знаю только крупные города и подставляюсь"
	case Solid:
		return "знаю больше городов, хожу наугад"
	case Complicated:
		return "знаю много и стараюсь дать неудобную букву"
	case Sadistic:
		return "знаю всё и загоняю в угол"
	case AI:
		return "знаю всё и приберегаю редкие буквы напоследок"
	default:
		return ""
	}
}

func (d Difficulty) String() string {
	return d.Emoji() + " " + d.Title()
}

func ParseDifficulty(key string) (Difficulty, bool) {
	for _, d := range difficulties {
		if d.Key() == key {
			return d, true
		}
	}
	return Simple, false
}

type curve struct {
	aggressionStart float64
	aggressionRate  float64
	famePool        int
	fameGrowth      int
	concedes        bool
}

var curves = map[Difficulty]curve{
	Simple:      {aggressionStart: 0.00, aggressionRate: 0.000, famePool: 40, fameGrowth: 6, concedes: true},
	Solid:       {aggressionStart: 0.05, aggressionRate: 0.025, famePool: 90, fameGrowth: 25, concedes: true},
	Complicated: {aggressionStart: 0.10, aggressionRate: 0.070, famePool: 200, fameGrowth: 90, concedes: false},
	Sadistic:    {aggressionStart: 0.15, aggressionRate: 0.160, famePool: 120, fameGrowth: 400, concedes: false},
	AI:          {aggressionStart: 0.20, aggressionRate: 0.320, famePool: 150, fameGrowth: 1200, concedes: false},
}

func (d Difficulty) curve() curve {
	if c, ok := curves[d]; ok {
		return c
	}
	return curves[Solid]
}

func (d Difficulty) Aggression(move int) float64 {
	c := d.curve()
	a := c.aggressionStart + c.aggressionRate*float64(move)
	if a < 0 {
		return 0
	}
	if a > 1 {
		return 1
	}
	return a
}

func (d Difficulty) FameWindow(move int) int {
	c := d.curve()
	return c.famePool + c.fameGrowth*move
}

func (d Difficulty) ConcedesOutsideWindow() bool {
	return d.curve().concedes
}
