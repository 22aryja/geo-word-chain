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
	compact  map[string]string
	byLetter map[string][]string
}

func New() (*Index, error) {
	lines := strings.Split(raw, "\n")

	index := &Index{
		all:      make(map[string]string, len(lines)),
		compact:  make(map[string]string, len(lines)),
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

		if c := Compact(key); c != key {
			if _, seen := index.compact[c]; !seen {
				index.compact[c] = key
			}
		}
	}

	if len(index.all) < minCities {
		return nil, fmt.Errorf("cities: loaded %d names, expected at least %d", len(index.all), minCities)
	}
	return index, nil
}

func (i *Index) Lookup(word string) (display string, ok bool) {
	key := Normalize(word)
	if key == "" {
		return "", false
	}
	if display, ok = i.all[key]; ok {
		return display, true
	}
	if normalized, found := i.compact[Compact(key)]; found {
		return i.all[normalized], true
	}
	return "", false
}

func (i *Index) Normalized(word string) string {
	key := Normalize(word)
	if _, ok := i.all[key]; ok {
		return key
	}
	return i.compact[Compact(key)]
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
