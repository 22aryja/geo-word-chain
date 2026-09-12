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
		{"!Москва", "м"}, // пропускает не-буквы
		{"Ыб", "ы"},      // ы в начале НЕ пропускается
		{"\u00a0", ""},   // регрессия: раньше паниковало
	}
	for _, tt := range tests {
		if got := FirstLetter(tt.in); got != tt.want {
			t.Errorf("FirstLetter(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestLastLetter(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Тверь", "р"},     // ь пропущен
		{"Чебоксары", "р"}, // ы пропущен
		{"Иркутск", "к"},
		{"Ростов-на-Дону", "у"},
		{"Ростов-", "в"}, // дефис в конце пропущен
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
		{"Волосвалі", false}, // украинская і
		{"Жезқазған", false}, // казахские қ, ғ
		{"Сма́ртно", false},  // комбинирующее ударение U+0301
		{"Moscow", false},    // латиница
		{"", false},
		{"---", false}, // нет букв
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

	// Messy input proves Lookup normalizes its argument.
	if display, ok := index.Lookup("  МОСКВА  "); !ok || display != "Москва" {
		t.Errorf(`Lookup("  МОСКВА  ") = %q, %v; want "Москва", true`, display, ok)
	}
	if _, ok := index.Lookup("Ленинград"); ok {
		t.Error("Ленинград should have been filtered out as a historic name")
	}

	// A letter that ends a city name must also start one, or a game can reach
	// a position with no legal move for either side.
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
		display, ok := index.Lookup(in)
		if !ok {
			t.Errorf("Lookup(%q) found nothing", in)
			continue
		}
		if display != "Нью-Йорк" {
			t.Errorf("Lookup(%q) = %q, want Нью-Йорк", in, display)
		}
	}

	want := index.Normalized("Нью-Йорк")
	for _, in := range []string{"нью йорк", "ньюйорк", "Нью–Йорк"} {
		if got := index.Normalized(in); got != want {
			t.Errorf("Normalized(%q) = %q, want %q", in, got, want)
		}
	}
}
