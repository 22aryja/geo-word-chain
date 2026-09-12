package game

type Result int

const (
	invalidResult Result = iota
	Accepted
	UnknownCity
	WrongLetter
	AlreadyUsed
	BotLost
)

func (r Result) String() string {
	switch r {
	case Accepted:
		return "accepted"
	case UnknownCity:
		return "unknown city"
	case WrongLetter:
		return "wrong letter"
	case AlreadyUsed:
		return "already used"
	case BotLost:
		return "bot lost"
	default:
		return "invalid"
	}
}

type Move struct {
	Result     Result
	PlayerCity string
	BotCity    string
	NextLetter string
}
