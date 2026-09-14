package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/22aryja/geo-word-chain/internal/cities"
)

type wikiEntry struct {
	title string
	fact  string
}

func loadWikiCache(path string) (map[string]wikiEntry, error) {
	out := make(map[string]wikiEntry)
	f, sc, err := open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}
	defer f.Close()

	for sc.Scan() {
		c := strings.SplitN(sc.Text(), "\t", 3)
		if len(c) < 3 || c[0] == "" {
			continue
		}
		out[c[0]] = wikiEntry{title: c[1], fact: c[2]}
	}
	return out, sc.Err()
}

func saveWikiCache(path string, cache map[string]wikiEntry) error {
	ids := make([]string, 0, len(cache))
	for id := range cache {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	return writeLines(path, len(ids), func(i int) string {
		e := cache[ids[i]]
		return ids[i] + "\t" + e.title + "\t" + e.fact
	})
}

func cleanTitle(title string) string {
	title = strings.TrimSpace(strings.ReplaceAll(title, "_", " "))
	if i := strings.Index(title, " ("); i > 0 && strings.HasSuffix(title, ")") {
		title = title[:i]
	}
	return strings.TrimSpace(title)
}

func usableName(name string) bool {
	return cities.IsRussianName(name) && cities.LastLetter(name) != ""
}

type rename struct {
	from     string
	to       string
	accepted bool
}

const ruralPopulationLimit = 30000

var urbanStems = []string{
	"город", "столиц", "мегаполис", "агломерац", "муниципалитет", "коммун",
	"тауншип", "гмин", "населённ", "населенн",
}

var ruralStems = []string{
	"посёлок", "поселок", "село", "селени", "деревн", "аул", "станиц", "хутор", "местечк",
}

var countryStopwords = map[string]bool{
	"и": true, "республика": true, "королевство": true, "федерация": true, "штаты": true,
	"соединённые": true, "соединенные": true, "объединённые": true, "объединенные": true,
	"демократическая": true, "народная": true, "острова": true, "остров": true,
}

func hasStem(words []string, stems []string) bool {
	for _, word := range words {
		for _, stem := range stems {
			if strings.HasPrefix(word, stem) {
				return true
			}
		}
	}
	return false
}

func definesSettlement(fact string, population int) bool {
	_, definition, ok := strings.Cut(fact, " — ")
	if !ok {
		return false
	}
	words := strings.Fields(strings.ToLower(definition))
	if len(words) > 6 {
		words = words[:6]
	}
	for i := range words {
		words[i] = strings.Trim(words[i], ",.;:()«»\"")
	}

	if hasStem(words, urbanStems) {
		return true
	}
	return hasStem(words, ruralStems) && population < ruralPopulationLimit
}

func mentionsCountry(fact, country string) bool {
	lower := strings.ToLower(fact)
	for _, word := range strings.Fields(strings.ToLower(country)) {
		word = strings.Trim(word, ",.;:()«»\"")
		if countryStopwords[word] {
			continue
		}
		stem := []rune(word)
		switch {
		case len(stem) >= 6:
			stem = stem[:len(stem)-2]
		case len(stem) == 5:
			stem = stem[:len(stem)-1]
		}
		if len(stem) >= 3 && strings.Contains(lower, string(stem)) {
			return true
		}
	}
	return false
}

func confirms(entry wikiEntry, country string, population int) bool {
	return definesSettlement(entry.fact, population) && mentionsCountry(entry.fact, country)
}

func applyWiki(list []place, cache map[string]wikiEntry, countryNames, overrides map[string]string) ([]place, []rename) {
	out := make([]place, len(list))
	var renames []rename
	for i, p := range list {
		out[i] = p

		if name, ok := overrides[p.id]; ok {
			out[i].display = name
			out[i].alias = ""
			out[i].wiki = true
			continue
		}

		entry, ok := cache[p.id]
		if !ok {
			continue
		}
		title := cleanTitle(entry.title)
		if !usableName(title) {
			continue
		}

		sameName := cities.Normalize(title) == cities.Normalize(p.display)
		if !confirms(entry, countryNames[p.country], p.population) {
			if !sameName {
				renames = append(renames, rename{from: p.display, to: title})
			}
			continue
		}

		out[i].wiki = true
		if !sameName {
			out[i].alias = p.display
			renames = append(renames, rename{from: p.display, to: title, accepted: true})
		}
		out[i].display = title
	}
	return out, renames
}

func resolve(list []place) ([]place, int) {
	byKey := make(map[string]place, len(list))
	for _, p := range list {
		key := cities.Normalize(p.display)
		if prev, ok := byKey[key]; ok && !better(p, prev) {
			continue
		}
		byKey[key] = p
	}

	out := make([]place, 0, len(byKey))
	for _, p := range byKey {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return better(out[i], out[j]) })

	claimed := make(map[string]bool)
	for i := range out {
		if out[i].alias == "" {
			continue
		}
		key := cities.Normalize(out[i].alias)
		if _, canonical := byKey[key]; canonical || claimed[key] {
			out[i].alias = ""
			continue
		}
		claimed[key] = true
	}

	sort.Slice(out, func(i, j int) bool { return out[i].display < out[j].display })
	return out, len(list) - len(byKey)
}

func countWiki(list []place) (titled, aliased int) {
	for _, p := range list {
		if p.wiki {
			titled++
		}
		if p.alias != "" {
			aliased++
		}
	}
	return titled, aliased
}

func reportRenames(renames []rename) {
	var accepted, rejected []rename
	for _, r := range renames {
		if r.accepted {
			accepted = append(accepted, r)
		} else {
			rejected = append(rejected, r)
		}
	}
	byTitle := func(list []rename) {
		sort.Slice(list, func(i, j int) bool { return list[i].to < list[j].to })
	}
	byTitle(accepted)
	byTitle(rejected)

	fmt.Printf("wikipedia renames: %d confirmed (old name kept as a second name), %d rejected as the wrong article\n",
		len(accepted), len(rejected))
	for _, r := range accepted {
		fmt.Printf("  + %-28s <- %s\n", r.to, r.from)
	}
	for _, r := range rejected {
		fmt.Printf("  - %-28s <- %s\n", r.to, r.from)
	}
}

func loadOverrides(path string) (map[string]string, error) {
	out := make(map[string]string)
	f, sc, err := open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}
	defer f.Close()

	for sc.Scan() {
		id, name, ok := strings.Cut(strings.TrimSpace(sc.Text()), "\t")
		if !ok || id == "" {
			continue
		}
		if name = strings.TrimSpace(name); usableName(name) {
			out[strings.TrimSpace(id)] = name
		}
	}
	return out, sc.Err()
}
