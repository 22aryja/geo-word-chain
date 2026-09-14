package cities

import "testing"

func TestAliasesResolveToOneCity(t *testing.T) {
	index, err := New()
	if err != nil {
		t.Fatal(err)
	}

	pairs := []struct{ canonical, other string }{
		{"Алеппо", "Халеб"},
		{"Калькутта", "Колката"},
	}
	for _, p := range pairs {
		main, ok := index.Lookup(p.canonical)
		if !ok {
			t.Errorf("%s not found", p.canonical)
			continue
		}
		second, ok := index.Lookup(p.other)
		if !ok {
			t.Errorf("%s not accepted as a second name for %s", p.other, p.canonical)
			continue
		}
		if index.Normalized(p.canonical) != index.Normalized(p.other) {
			t.Errorf("%s and %s resolve to different cities", p.canonical, p.other)
		}
		if second.Name != p.other {
			t.Errorf("Lookup(%q).Name = %q; the typed spelling must be kept", p.other, second.Name)
		}
		if second.Country != main.Country {
			t.Errorf("%s and %s disagree on country", p.canonical, p.other)
		}
	}
}

func TestDeprecatedNamesStayRejected(t *testing.T) {
	index, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Ленинград", "Сталинград", "Свердловск", "Нур-Султан", "Питер", "Киеву", "Алма-Ата"} {
		if city, ok := index.Lookup(name); ok {
			t.Errorf("%q accepted as %q", name, city.Name)
		}
	}
}

func TestAlmatyIsAlmaty(t *testing.T) {
	index, err := New()
	if err != nil {
		t.Fatal(err)
	}
	city, ok := index.Lookup("алматы")
	if !ok || city.Name != "Алматы" || city.Country != "KZ" {
		t.Errorf("Lookup(алматы) = %q %q %v", city.Name, city.Country, ok)
	}
	if city.Fact == "" {
		t.Error("Алматы lost its fact")
	}
}
