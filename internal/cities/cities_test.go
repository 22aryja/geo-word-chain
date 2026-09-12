package cities

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{"нижний регистр", "Тверь", "тверь"},
		{"ё в е", "Ёбург", "ебург"},
		{"заглавная Ё тоже", "ЁЛКИ", "елки"},
		{"обрезает края", "  Иркутск  ", "иркутск"},
		{"схлопывает пробелы", " Нижний  Новгород ", "нижний новгород"},
		{"дефис становится пробелом", "Ростов-на-Дону", "ростов на дону"},
		{"тире тоже", "Нью–Йорк", "нью йорк"},
		{"только пробелы", "\u00a0", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Normalize(tt.in); got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestFirstLetter(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Тверь", "т"},
		{"Ростов-на-Дону", "р"},
		{"!Москва", "м"},
		{"Ыб", "ы"},
		{"\u00a0", ""},
	}
	for _, tt := range tests {
		if got := FirstLetter(tt.in); got != tt.want {
			t.Errorf("FirstLetter(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestLastLetter(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Тверь", "р"},
		{"Чебоксары", "р"},
		{"Иркутск", "к"},
		{"Ростов-на-Дону", "у"},
		{"Ростов-", "в"},
		{"\u00a0", ""},
	}
	for _, tt := range tests {
		if got := LastLetter(tt.in); got != tt.want {
			t.Errorf("LastLetter(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestIsRussianName(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"Москва", true},
		{"Ростов-на-Дону", true},
		{"Нижний Новгород", true},
		{"Волосвалі", false},
		{"Жезқазған", false},
		{"Сма́ртно", false},
		{"Moscow", false},
		{"", false},
		{"---", false},
	}
	for _, tt := range tests {
		if got := IsRussianName(tt.in); got != tt.want {
			t.Errorf("IsRussianName(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestNew(t *testing.T) {
	index, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if got := index.Len(); got < 20000 {
		t.Errorf("Len() = %d, want at least 20000", got)
	}

	if city, ok := index.Lookup("  МОСКВА  "); !ok || city.Name != "Москва" {
		t.Errorf(`Lookup("  МОСКВА  ") = %q, %v; want "Москва", true`, city.Name, ok)
	}
	if _, ok := index.Lookup("Ленинград"); ok {
		t.Error("Ленинград should have been filtered out as a historic name")
	}

	for _, name := range []string{"Москва", "Тверь", "Иркутск"} {
		if last := LastLetter(name); len(index.ByLetter(last)) == 0 {
			t.Errorf("no cities start with %q (last letter of %q)", last, name)
		}
	}
	if len(index.ByLetter("Щ")) == 0 {
		t.Error(`ByLetter("Щ") is empty; uppercase input should be normalized`)
	}
}

func TestLookupSeparators(t *testing.T) {
	index, err := New()
	if err != nil {
		t.Fatal(err)
	}

	for _, in := range []string{
		"Нью-Йорк",
		"нью йорк",
		"НЬЮ ЙОРК",
		"нью-йорк",
		"ньюйорк",
		"  нью   йорк  ",
	} {
		city, ok := index.Lookup(in)
		if !ok {
			t.Errorf("Lookup(%q) found nothing", in)
			continue
		}
		if city.Name != "Нью-Йорк" {
			t.Errorf("Lookup(%q) = %q, want Нью-Йорк", in, city)
		}
	}

	want := index.Normalized("Нью-Йорк")
	for _, in := range []string{"нью йорк", "ньюйорк", "Нью–Йорк"} {
		if got := index.Normalized(in); got != want {
			t.Errorf("Normalized(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCityCountry(t *testing.T) {
	index, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if got := index.Countries(); got < 200 {
		t.Errorf("Countries() = %d, want the full ISO list", got)
	}

	tests := []struct{ name, iso, country, flag string }{
		{"Москва", "RU", "Россия", "🇷🇺"},
		{"Нью-Йорк", "US", "США", "🇺🇸"},
		{"Париж", "FR", "Франция", "🇫🇷"},
		{"Алматы", "KZ", "Казахстан", "🇰🇿"},
	}
	for _, tt := range tests {
		city, ok := index.Lookup(tt.name)
		if !ok {
			t.Errorf("Lookup(%q) found nothing", tt.name)
			continue
		}
		if city.Country != tt.iso {
			t.Errorf("%s: Country = %q, want %q", tt.name, city.Country, tt.iso)
		}
		if city.CountryName != tt.country {
			t.Errorf("%s: CountryName = %q, want %q", tt.name, city.CountryName, tt.country)
		}
		if city.Flag() != tt.flag {
			t.Errorf("%s: Flag() = %q, want %q", tt.name, city.Flag(), tt.flag)
		}
	}
}

func TestFlagFallsBack(t *testing.T) {
	for _, c := range []City{{Country: ""}, {Country: "X"}, {Country: "usa"}, {Country: "u1"}} {
		if got := c.Flag(); got != "" {
			t.Errorf("City{Country:%q}.Flag() = %q, want empty", c.Country, got)
		}
	}
}
