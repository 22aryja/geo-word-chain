package cities

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed data/cities_ru.txt
var raw string

const minCities = 1000

type Index struct {
	all      map[string]string
	byLetter map[string][]string
}

func New() (*Index, error) {
	lines := strings.Split(raw, "\n")

	index := &Index{
		all:      make(map[string]string, len(lines)),
		byLetter: make(map[string][]string),
	}

	for _, line := range lines {
		display := strings.TrimSpace(line)
		if display == "" {
			continue
		}

		key := Normalize(display)
		letter := FirstLetter(key)
		if key == "" || letter == "" {
			continue
		}
		if _, seen := index.all[key]; seen {
			continue
		}

		index.all[key] = display
		index.byLetter[letter] = append(index.byLetter[letter], key)
	}

	if len(index.all) < minCities {
		return nil, fmt.Errorf("cities: loaded %d names, expected at least %d", len(index.all), minCities)
	}
	return index, nil
}

func (i *Index) Lookup(word string) (display string, ok bool) {
	display, ok = i.all[Normalize(word)]
	return display, ok
}

func (i *Index) ByLetter(letter string) []string {
	return i.byLetter[Normalize(letter)]
}

func (i *Index) Display(normalized string) string {
	return i.all[normalized]
}

func (i *Index) Len() int {
	return len(i.all)
}
