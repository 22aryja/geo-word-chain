package cities

import (
	"strings"
)

const ruAlphabet string = "абвгдежзийклмнопрстуфхцчшщъыьэюяё"

var allowed = func() map[rune]bool {
	m := make(map[rune]bool, len([]rune(ruAlphabet)))
	for _, r := range ruAlphabet {
		m[r] = true
	}
	return m
}()

func Normalize(word string) string {
	var sanitized string = strings.ToLower(strings.TrimSpace(word))
	var runes []rune = []rune(sanitized)

	for i := 0; i < len(runes); i++ {
		switch runes[i] {
		case 'ё':
			runes[i] = 'е'
		case '-', '\u2010', '\u2011', '\u2012', '\u2013', '\u2014', '\u2212':
			runes[i] = ' '
		}
	}

	return strings.Join(strings.Fields(string(runes)), " ")
}

func Compact(normalized string) string {
	return strings.ReplaceAll(normalized, " ", "")
}

func FirstLetter(word string) string {
	runes := []rune(Normalize(word))
	for i := 0; i < len(runes); i++ {
		if !allowed[runes[i]] {
			continue
		}
		return string(runes[i])
	}
	return ""
}

func LastLetter(word string) string {
	var normalized string = Normalize(word)
	var runes []rune = []rune(normalized)
	var letter string = ""

	for i := len(runes) - 1; i >= 0; i-- {
		if !allowed[runes[i]] || runes[i] == 'ы' || runes[i] == 'ь' || runes[i] == 'ъ' {
			continue
		}
		letter = string(runes[i])
		break
	}

	return letter
}

func IsRussianName(s string) bool {
	norm := Normalize(s)
	hasLetter := false
	for _, r := range norm {
		switch {
		case allowed[r]:
			hasLetter = true
		case r == ' ':
		default:
			return false
		}
	}
	return hasLetter
}
