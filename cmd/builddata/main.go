// Command builddata turns the raw GeoNames dumps into the compact city list
// that the bot embeds.
//
// It is run by hand, not by the bot. Download the two dumps from
// https://download.geonames.org/export/dump/ and run:
//
//	go run ./cmd/builddata -alt ~/Downloads/alternateNamesV2.txt -cities ~/Downloads/cities5000.txt
//
// The generated file is committed to the repository; the bot never reads the
// dumps. Filtering and deduplication go through internal/cities so that the
// generated list and the bot's runtime lookups can never disagree.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/22aryja/geo-word-chain/internal/cities"
)

// Column indices in the GeoNames dumps, both tab-separated and documented at
// https://download.geonames.org/export/dump/readme.txt
const (
	// cities5000.txt
	cityID         = 0
	cityCountry    = 8
	cityPopulation = 14
	cityColumns    = 19

	// alternateNamesV2.txt
	altGeonameID    = 1
	altISOLanguage  = 2
	altName         = 3
	altIsPreferred  = 4
	altIsColloquial = 6
	altIsHistoric   = 7
	altColumns      = 8 // trailing "from"/"to" columns are often absent
)

type place struct {
	display string
	country string
}

func main() {
	var (
		altPath   = flag.String("alt", "", "path to alternateNamesV2.txt (required)")
		cityPath  = flag.String("cities", "", "path to cities5000.txt (required)")
		outPath   = flag.String("out", filepath.Join("internal", "cities", "data", "cities_ru.txt"), "generated city list")
		lang      = flag.String("lang", "ru", "isolanguage code to extract")
		country   = flag.String("country", "", "restrict to one ISO country code (empty means worldwide)")
		minPop    = flag.Int("min-pop", 0, "skip cities below this population")
		histogram = flag.Bool("histogram", true, "report cities per starting/ending letter")
	)
	flag.Parse()

	if *altPath == "" || *cityPath == "" {
		fmt.Fprintln(os.Stderr, "builddata: -alt and -cities are required")
		flag.Usage()
		os.Exit(2)
	}

	// Pass 1 reads the small file first so that pass 2 can discard the vast
	// majority of the 783MB dump without allocating for it: only ~70k
	// geonameids are populated places we care about.
	wanted, err := readCities(*cityPath, *country, *minPop)
	if err != nil {
		fmt.Fprintln(os.Stderr, "builddata:", err)
		os.Exit(1)
	}
	fmt.Printf("cities5000: %d places match the country/population filter\n", len(wanted))

	named, err := readNames(*altPath, *lang, wanted)
	if err != nil {
		fmt.Fprintln(os.Stderr, "builddata:", err)
		os.Exit(1)
	}
	fmt.Printf("alternate names: %d of them have a usable %q name\n", len(named), *lang)

	list, dropped := dedupe(named)
	fmt.Printf("after dedupe by normalized form: %d unique cities (%d duplicates collapsed)\n", len(list), dropped)

	if err := write(*outPath, list); err != nil {
		fmt.Fprintln(os.Stderr, "builddata:", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s\n", *outPath)

	if *histogram {
		reportLetters(list)
	}
}

// readCities returns the geonameids of populated places passing the filters.
func readCities(path, country string, minPop int) (map[string]string, error) {
	f, sc, err := open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := make(map[string]string)
	for sc.Scan() {
		c := strings.Split(sc.Text(), "\t")
		if len(c) < cityColumns {
			continue
		}
		if country != "" && c[cityCountry] != country {
			continue
		}
		if minPop > 0 {
			pop, err := strconv.Atoi(c[cityPopulation])
			if err != nil || pop < minPop {
				continue
			}
		}
		out[c[cityID]] = c[cityCountry]
	}
	return out, sc.Err()
}

// readNames streams the large dump and keeps one name per wanted geonameid.
func readNames(path, lang string, wanted map[string]string) (map[string]place, error) {
	f, sc, err := open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := make(map[string]place, len(wanted))
	preferred := make(map[string]bool, len(wanted))

	for sc.Scan() {
		c := strings.Split(sc.Text(), "\t")
		if len(c) < altColumns {
			continue
		}
		if c[altISOLanguage] != lang {
			continue
		}
		// Historic names (Ленинград, Молотов) and colloquial ones (Питер, СПб)
		// are real answers to a different question than the one the game asks.
		if c[altIsHistoric] == "1" || c[altIsColloquial] == "1" {
			continue
		}
		id := c[altGeonameID]
		cc, ok := wanted[id]
		if !ok {
			continue
		}
		// GeoNames labels some Ukrainian, Kazakh and Serbian names as "ru".
		// This is the filter that removes them, and it is the same predicate
		// the bot uses at runtime.
		name := strings.TrimSpace(c[altName])
		if !cities.IsRussianName(name) {
			continue
		}
		// A city with no chainable last letter would dead-end the game.
		if cities.LastLetter(name) == "" {
			continue
		}

		isPref := c[altIsPreferred] == "1"
		if _, seen := out[id]; !seen || (isPref && !preferred[id]) {
			out[id] = place{display: name, country: cc}
			preferred[id] = isPref
		}
	}
	return out, sc.Err()
}

// dedupe collapses places whose names normalize to the same key. Homonyms are
// common (many settlements are called Александровка) and the game only ever
// compares the string, so one entry per spelling is enough.
func dedupe(named map[string]place) ([]string, int) {
	byKey := make(map[string]string, len(named))
	for _, p := range named {
		key := cities.Normalize(p.display)
		// Map iteration is randomized, so pick deterministically rather than
		// letting whichever entry arrives last win.
		if prev, ok := byKey[key]; ok {
			if prev <= p.display {
				continue
			}
		}
		byKey[key] = p.display
	}

	list := make([]string, 0, len(byKey))
	for _, display := range byKey {
		list = append(list, display)
	}
	sort.Strings(list)
	return list, len(named) - len(byKey)
}

func write(path string, list []string) error {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, name := range list {
		if _, err := fmt.Fprintln(w, name); err != nil {
			return err
		}
	}
	return w.Flush()
}

// reportLetters prints how many cities start and end with each letter. A letter
// that ends words but starts none is a dead end: reaching it makes the next
// move impossible for either player.
func reportLetters(list []string) {
	first := make(map[string]int)
	last := make(map[string]int)
	for _, name := range list {
		first[cities.FirstLetter(name)]++
		last[cities.LastLetter(name)]++
	}

	letters := make([]string, 0, len(first))
	seen := make(map[string]bool)
	for _, m := range []map[string]int{first, last} {
		for l := range m {
			if !seen[l] {
				seen[l] = true
				letters = append(letters, l)
			}
		}
	}
	sort.Strings(letters)

	fmt.Println("\nletter  starting  ending")
	for _, l := range letters {
		note := ""
		switch {
		case first[l] == 0 && last[l] > 0:
			note = "  <-- DEAD END"
		case last[l] > 0 && first[l] < 10:
			note = "  <-- thin"
		}
		fmt.Printf("  %s   %8d  %6d%s\n", l, first[l], last[l], note)
	}
}

func open(path string) (*os.File, *bufio.Scanner, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	sc := bufio.NewScanner(f)
	// The alternatenames column in cities5000.txt runs well past the 64KB
	// default token size.
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	return f, sc, nil
}
