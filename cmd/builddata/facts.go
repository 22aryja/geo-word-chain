package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	userAgent      = "geo-word-chain-builddata/1.0 (https://github.com/22aryja/geo-word-chain)"
	factMaxRunes   = 320
	factSentences  = 2
	minSentence    = 30
	fetchWorkers   = 10
	fetchDelay     = 20 * time.Millisecond
	requestTimeout = 25 * time.Second
	maxAttempts    = 4
)

type summary struct {
	Type    string `json:"type"`
	Extract string `json:"extract"`
}

func wikiTitle(link string) string {
	parsed, err := url.Parse(link)
	if err != nil {
		return ""
	}
	title := strings.TrimPrefix(parsed.Path, "/wiki/")
	if decoded, err := url.PathUnescape(title); err == nil {
		return decoded
	}
	return title
}

func trimSentences(text string, want int) string {
	text = strings.Join(strings.Fields(text), " ")
	if text == "" {
		return ""
	}

	runes := []rune(text)
	found, start := 0, 0
	for i, r := range runes {
		if r != '.' && r != '!' && r != '?' {
			continue
		}
		if i+1 < len(runes) && runes[i+1] != ' ' {
			continue
		}
		if i-start < minSentence {
			continue
		}
		found++
		start = i + 1
		if found >= want {
			runes = runes[:i+1]
			break
		}
	}

	if len(runes) > factMaxRunes {
		return strings.TrimSpace(string(runes[:factMaxRunes])) + "…"
	}
	return strings.TrimSpace(string(runes))
}

type stats struct {
	mu      sync.Mutex
	reasons map[string]int
	samples map[string]string
}

func newStats() *stats {
	return &stats{reasons: map[string]int{}, samples: map[string]string{}}
}

func (s *stats) record(reason, detail string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reasons[reason]++
	if _, ok := s.samples[reason]; !ok {
		s.samples[reason] = detail
	}
}

func (s *stats) report() {
	s.mu.Lock()
	defer s.mu.Unlock()

	keys := make([]string, 0, len(s.reasons))
	for k := range s.reasons {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(a, b int) bool { return s.reasons[keys[a]] > s.reasons[keys[b]] })

	for _, k := range keys {
		line := fmt.Sprintf("  %-18s %6d", k, s.reasons[k])
		if sample := s.samples[k]; sample != "" && k != "ok" {
			line += "   e.g. " + sample
		}
		fmt.Println(line)
	}
}

func fetchSummary(client *http.Client, lang, title string, tally *stats) string {
	endpoint := fmt.Sprintf("https://%s.wikipedia.org/api/rest_v1/page/summary/%s",
		lang, url.PathEscape(title))

	backoff := time.Second
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			tally.record("bad request", title)
			return ""
		}
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			if attempt == maxAttempts {
				tally.record("network", title+": "+err.Error())
				return ""
			}
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			wait := backoff
			if retry := resp.Header.Get("Retry-After"); retry != "" {
				if secs, err := strconv.Atoi(retry); err == nil {
					wait = time.Duration(secs) * time.Second
				}
			}
			resp.Body.Close()
			if attempt == maxAttempts {
				tally.record("throttled "+strconv.Itoa(resp.StatusCode), title)
				return ""
			}
			time.Sleep(wait)
			backoff *= 2
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			tally.record("http "+strconv.Itoa(resp.StatusCode), title)
			return ""
		}

		var out summary
		err = json.NewDecoder(resp.Body).Decode(&out)
		resp.Body.Close()
		if err != nil {
			tally.record("bad json", title)
			return ""
		}
		if out.Type == "disambiguation" {
			tally.record("disambiguation", title)
			return ""
		}

		if !looksLikePlace(out.Extract) {
			tally.record("not a place", title)
			return ""
		}

		fact := trimSentences(out.Extract, factSentences)
		if fact == "" {
			tally.record("empty extract", title)
			return ""
		}
		tally.record("ok", "")
		return fact
	}
	return ""
}

type factJob struct {
	display string
	text    string
}

func factCandidates(list []place, notable map[string]bool, limit int) []place {
	out := make([]place, 0, len(list))
	for _, p := range list {
		if notable[p.id] {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(a, b int) bool {
		if out[a].population != out[b].population {
			return out[a].population > out[b].population
		}
		return out[a].display < out[b].display
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func fetchFacts(list []place, links map[string]string, notable map[string]bool, lang string, limit int) map[string]string {
	candidates := factCandidates(list, notable, limit)
	fmt.Printf("fetching Wikipedia summaries for the %d most populous notable cities\n", len(candidates))

	jobs := make(chan factJob)
	results := make(chan factJob)
	tally := newStats()

	client := &http.Client{Timeout: requestTimeout}

	var wg sync.WaitGroup
	for range fetchWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				if fact := fetchSummary(client, lang, job.text, tally); fact != "" {
					results <- factJob{display: job.display, text: fact}
				}
				time.Sleep(fetchDelay)
			}
		}()
	}

	go func() {
		for _, p := range candidates {
			title := wikiTitle(links[p.id])
			if title == "" {
				title = p.display
			}
			jobs <- factJob{display: p.display, text: title}
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	facts := make(map[string]string)
	for r := range results {
		facts[r.display] = r.text
		if len(facts)%500 == 0 {
			fmt.Printf("  %d facts so far\n", len(facts))
		}
	}

	fmt.Println("fetch outcomes:")
	tally.report()
	return facts
}

func writeFacts(path string, facts map[string]string, list []place) error {
	ordered := make([]string, 0, len(facts))
	for _, p := range list {
		if _, ok := facts[p.display]; ok {
			ordered = append(ordered, p.display)
		}
	}
	return writeLines(path, len(ordered), func(i int) string {
		return ordered[i] + "\t" + facts[ordered[i]]
	})
}

var placeWords = []string{
	"город", "столица", "посёлок", "поселок", "село", "деревня", "коммуна",
	"муниципалитет", "населённый пункт", "населенный пункт", "округ", "штат",
	"провинция", "префектура", "район", "порт", "курорт", "агломерация",
	"община", "графство", "уезд", "воеводство", "департамент", "кантон",
	"местечко", "аул", "станица", "хутор", "гмина", "коммуны", "тауншип",
}

func looksLikePlace(extract string) bool {
	lower := strings.ToLower(extract)
	for _, word := range placeWords {
		if strings.Contains(lower, word) {
			return true
		}
	}
	return false
}
