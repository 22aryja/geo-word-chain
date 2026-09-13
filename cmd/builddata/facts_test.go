package main

import "testing"

func TestWikiTitle(t *testing.T) {
	tests := []struct{ link, want string }{
		{"https://ru.wikipedia.org/wiki/%D0%93%D0%B0%D0%BC%D0%B1%D1%83%D1%80%D0%B3", "Гамбург"},
		{"https://ru.wikipedia.org/wiki/Алматы", "Алматы"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := wikiTitle(tt.link); got != tt.want {
			t.Errorf("wikiTitle(%q) = %q, want %q", tt.link, got, tt.want)
		}
	}
}

func TestTrimSentences(t *testing.T) {
	long := "Гамбург — город на севере Германии. Вольный и ганзейский город является одной из земель. Это второй по величине город в стране."
	got := trimSentences(long, 2)
	want := "Гамбург — город на севере Германии. Вольный и ганзейский город является одной из земель."
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}

	if got := trimSentences("", 2); got != "" {
		t.Errorf("empty input produced %q", got)
	}

	short := "Аба — город в Нигерии."
	if got := trimSentences(short, 2); got != short {
		t.Errorf("single sentence changed: %q", got)
	}

	abbrev := "Село расположено в 5 км. от райцентра и известно с 1802 года. Вторая фраза здесь."
	if got := trimSentences(abbrev, 1); got == "Село расположено в 5 км." {
		t.Errorf("split on an abbreviation: %q", got)
	}
}

func TestLooksLikePlace(t *testing.T) {
	places := []string{
		"Гамбург — город на севере Германии.",
		"Абуджа — столица Нигерии с 1991 года.",
		"Аньер-сюр-Сен — коммуна во Франции.",
		"Ытык-Кюёль — село в Якутии.",
	}
	for _, p := range places {
		if !looksLikePlace(p) {
			t.Errorf("rejected a real place: %q", p)
		}
	}

	notPlaces := []string{
		"Абаза — один из абхазо-адыгских народов.",
		"Аргентина — музыкальный альбом группы.",
		"",
	}
	for _, p := range notPlaces {
		if looksLikePlace(p) {
			t.Errorf("accepted a non-place: %q", p)
		}
	}
}
