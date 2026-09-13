package cities

import (
	_ "embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//go:embed data/cities_ru.txt
var raw string

//go:embed data/countries_ru.txt
var rawCountries string

//go:embed data/facts_ru.txt
var rawFacts string

const minCities = 1000

type City struct {
	Name        string
	Country     string
	CountryName string
	Population  int
	Fact        string
}

func (c City) Flag() string {
	if len(c.Country) != 2 {
		return ""
	}
	var sb strings.Builder
	for _, r := range c.Country {
		if r < 'A' || r > 'Z' {
			return ""
		}
		sb.WriteRune(0x1F1E6 + (r - 'A'))
	}
	return sb.String()
}

type Index struct {
	all       map[string]City
	compact   map[string]string
	byLetter  map[string][]string
	countries map[string]string
}

func New() (*Index, error) {
	countries := parseCountries(rawCountries)
	facts := parseFacts(rawFacts)
	lines := strings.Split(raw, "\n")

	index := &Index{
		all:       make(map[string]City, len(lines)),
		compact:   make(map[string]string, len(lines)),
		byLetter:  make(map[string][]string),
		countries: countries,
	}

	for _, line := range lines {
		fields := strings.Split(strings.TrimSpace(line), "\t")
		display := strings.TrimSpace(fields[0])
		if display == "" {
			continue
		}
		var iso string
		var population int
		if len(fields) > 1 {
			iso = strings.TrimSpace(fields[1])
		}
		if len(fields) > 2 {
			population, _ = strconv.Atoi(strings.TrimSpace(fields[2]))
		}

		key := Normalize(display)
		letter := FirstLetter(key)
		if key == "" || letter == "" {
			continue
		}
		if _, seen := index.all[key]; seen {
			continue
		}

		index.all[key] = City{
			Name:        display,
			Country:     iso,
			CountryName: countries[iso],
			Population:  population,
			Fact:        facts[key],
		}
		index.byLetter[letter] = append(index.byLetter[letter], key)

		if c := Compact(key); c != key {
			if _, seen := index.compact[c]; !seen {
				index.compact[c] = key
			}
		}
	}

	index.sortByFame()

	if len(index.all) < minCities {
		return nil, fmt.Errorf("cities: loaded %d names, expected at least %d", len(index.all), minCities)
	}
	return index, nil
}

func parseCountries(raw string) map[string]string {
	out := make(map[string]string)
	for _, line := range strings.Split(raw, "\n") {
		iso, name, ok := strings.Cut(strings.TrimSpace(line), "\t")
		if !ok {
			continue
		}
		if iso = strings.TrimSpace(iso); iso != "" {
			out[iso] = strings.TrimSpace(name)
		}
	}
	return out
}

func (i *Index) Lookup(word string) (city City, ok bool) {
	key := Normalize(word)
	if key == "" {
		return City{}, false
	}
	if city, ok = i.all[key]; ok {
		return city, true
	}
	if normalized, found := i.compact[Compact(key)]; found {
		return i.all[normalized], true
	}
	return City{}, false
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

func (i *Index) City(normalized string) City {
	return i.all[normalized]
}

func (i *Index) Len() int {
	return len(i.all)
}

func (i *Index) Countries() int {
	return len(i.countries)
}

func (i *Index) sortByFame() {
	for letter, bucket := range i.byLetter {
		sort.SliceStable(bucket, func(a, b int) bool {
			pa, pb := i.all[bucket[a]].Population, i.all[bucket[b]].Population
			if pa != pb {
				return pa > pb
			}
			return bucket[a] < bucket[b]
		})
		i.byLetter[letter] = bucket
	}
}

func (i *Index) LetterCounts() map[string]int {
	out := make(map[string]int, len(i.byLetter))
	for letter, bucket := range i.byLetter {
		out[letter] = len(bucket)
	}
	return out
}

func parseFacts(raw string) map[string]string {
	out := make(map[string]string)
	for _, line := range strings.Split(raw, "\n") {
		name, fact, ok := strings.Cut(strings.TrimSpace(line), "\t")
		if !ok {
			continue
		}
		if key := Normalize(name); key != "" {
			out[key] = strings.TrimSpace(fact)
		}
	}
	return out
}

func (i *Index) Facts() int {
	n := 0
	for _, c := range i.all {
		if c.Fact != "" {
			n++
		}
	}
	return n
}
