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

const (
	cityID         = 0
	cityCountry    = 8
	cityPopulation = 14
	cityColumns    = 19

	altGeonameID    = 1
	altISOLanguage  = 2
	altName         = 3
	altIsPreferred  = 4
	altIsColloquial = 6
	altIsHistoric   = 7
	altColumns      = 8

	countryISO       = 0
	countryGeonameID = 16
	countryColumns   = 17
)

type place struct {
	display    string
	country    string
	population int
}

type meta struct {
	country    string
	population int
}

func main() {
	var (
		altPath     = flag.String("alt", "", "path to alternateNamesV2.txt (required)")
		cityPath    = flag.String("cities", "", "path to cities5000.txt (required)")
		countryPath = flag.String("countries", "", "path to countryInfo.txt (required)")
		outDir      = flag.String("out", filepath.Join("internal", "cities", "data"), "directory for the generated lists")
		lang        = flag.String("lang", "ru", "isolanguage code to extract")
		country     = flag.String("country", "", "restrict to one ISO country code (empty means worldwide)")
		minPop      = flag.Int("min-pop", 0, "skip cities below this population")
		histogram   = flag.Bool("histogram", true, "report cities per starting/ending letter")
	)
	flag.Parse()

	if *altPath == "" || *cityPath == "" || *countryPath == "" {
		fmt.Fprintln(os.Stderr, "builddata: -alt, -cities and -countries are required")
		flag.Usage()
		os.Exit(2)
	}

	if err := run(*altPath, *cityPath, *countryPath, *outDir, *lang, *country, *minPop, *histogram); err != nil {
		fmt.Fprintln(os.Stderr, "builddata:", err)
		os.Exit(1)
	}
}

func run(altPath, cityPath, countryPath, outDir, lang, country string, minPop int, histogram bool) error {

	wantedCities, err := readCities(cityPath, country, minPop)
	if err != nil {
		return err
	}
	fmt.Printf("cities5000: %d places match the country/population filter\n", len(wantedCities))

	wantedCountries, err := readCountries(countryPath)
	if err != nil {
		return err
	}
	fmt.Printf("countryInfo: %d countries\n", len(wantedCountries))

	named, countryNames, err := readNames(altPath, lang, wantedCities, wantedCountries)
	if err != nil {
		return err
	}
	fmt.Printf("alternate names: %d cities and %d countries have a %q name\n",
		len(named), len(countryNames), lang)

	list, dropped := dedupe(named)
	fmt.Printf("after dedupe by normalized form: %d unique cities (%d duplicates collapsed)\n", len(list), dropped)

	cityFile := filepath.Join(outDir, "cities_"+lang+".txt")
	if err := writeCities(cityFile, list); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", cityFile)

	countryFile := filepath.Join(outDir, "countries_"+lang+".txt")
	if err := writeCountries(countryFile, countryNames); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", countryFile)

	if missing := missingCountries(list, countryNames); len(missing) > 0 {
		fmt.Printf("\nwarning: %d country codes have no %q name: %v\n", len(missing), lang, missing)
	}
	if histogram {
		reportLetters(list)
	}
	return nil
}

func readCities(path, country string, minPop int) (map[string]meta, error) {
	f, sc, err := open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := make(map[string]meta)
	for sc.Scan() {
		c := strings.Split(sc.Text(), "\t")
		if len(c) < cityColumns {
			continue
		}
		if country != "" && c[cityCountry] != country {
			continue
		}

		pop, err := strconv.Atoi(c[cityPopulation])
		if err != nil {
			pop = 0
		}
		if pop < minPop {
			continue
		}
		out[c[cityID]] = meta{country: c[cityCountry], population: pop}
	}
	return out, sc.Err()
}

func readCountries(path string) (map[string]string, error) {
	f, sc, err := open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := make(map[string]string)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		c := strings.Split(line, "\t")
		if len(c) < countryColumns || c[countryGeonameID] == "" {
			continue
		}
		out[c[countryGeonameID]] = c[countryISO]
	}
	return out, sc.Err()
}

func readNames(path, lang string, cityIDs map[string]meta, countryIDs map[string]string) (map[string]place, map[string]string, error) {
	f, sc, err := open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	named := make(map[string]place, len(cityIDs))
	countryNames := make(map[string]string, len(countryIDs))
	preferredCity := make(map[string]bool, len(cityIDs))
	preferredCountry := make(map[string]bool, len(countryIDs))

	for sc.Scan() {
		c := strings.Split(sc.Text(), "\t")
		if len(c) < altColumns || c[altISOLanguage] != lang {
			continue
		}

		if c[altIsHistoric] == "1" || c[altIsColloquial] == "1" {
			continue
		}

		id := c[altGeonameID]
		name := strings.TrimSpace(c[altName])
		isPreferred := c[altIsPreferred] == "1"

		if iso, ok := countryIDs[id]; ok {

			if _, seen := countryNames[iso]; !seen || (isPreferred && !preferredCountry[iso]) {
				countryNames[iso] = name
				preferredCountry[iso] = isPreferred
			}
			continue
		}

		m, ok := cityIDs[id]
		if !ok {
			continue
		}

		if !cities.IsRussianName(name) {
			continue
		}

		if cities.LastLetter(name) == "" {
			continue
		}
		if _, seen := named[id]; !seen || (isPreferred && !preferredCity[id]) {
			named[id] = place{display: name, country: m.country, population: m.population}
			preferredCity[id] = isPreferred
		}
	}
	return named, countryNames, sc.Err()
}

func dedupe(named map[string]place) ([]place, int) {
	byKey := make(map[string]place, len(named))
	for _, p := range named {
		key := cities.Normalize(p.display)

		if prev, ok := byKey[key]; ok && !better(p, prev) {
			continue
		}
		byKey[key] = p
	}

	list := make([]place, 0, len(byKey))
	for _, p := range byKey {
		list = append(list, p)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].display < list[j].display })
	return list, len(named) - len(byKey)
}

func writeCities(path string, list []place) error {
	return writeLines(path, len(list), func(i int) string {
		return list[i].display + "\t" + list[i].country
	})
}

func writeCountries(path string, names map[string]string) error {
	isos := make([]string, 0, len(names))
	for iso := range names {
		isos = append(isos, iso)
	}
	sort.Strings(isos)

	return writeLines(path, len(isos), func(i int) string {
		return isos[i] + "\t" + names[isos[i]]
	})
}

func writeLines(path string, n int, line func(int) string) error {
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
	for i := range n {
		if _, err := fmt.Fprintln(w, line(i)); err != nil {
			return err
		}
	}
	return w.Flush()
}

func missingCountries(list []place, names map[string]string) []string {
	missing := map[string]bool{}
	for _, p := range list {
		if _, ok := names[p.country]; !ok {
			missing[p.country] = true
		}
	}
	out := make([]string, 0, len(missing))
	for iso := range missing {
		out = append(out, iso)
	}
	sort.Strings(out)
	return out
}

func reportLetters(list []place) {
	first := make(map[string]int)
	last := make(map[string]int)
	for _, p := range list {
		first[cities.FirstLetter(p.display)]++
		last[cities.LastLetter(p.display)]++
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

	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	return f, sc, nil
}

func better(a, b place) bool {
	if a.population != b.population {
		return a.population > b.population
	}
	if a.display != b.display {
		return a.display < b.display
	}
	return a.country < b.country
}
