package main

import "testing"

func TestCleanTitle(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Алеппо", "Алеппо"},
		{"Хайдарабад_(город_в_Индии)", "Хайдарабад"},
		{"Аренал (Коста-Рика)", "Аренал"},
		{"Ростов-на-Дону", "Ростов-на-Дону"},
	}
	for _, tt := range tests {
		if got := cleanTitle(tt.in); got != tt.want {
			t.Errorf("cleanTitle(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestConfirmsRealCities(t *testing.T) {
	type tc struct {
		title, fact, country string
		population           int
		want                 bool
	}
	tests := []tc{
		{title: "Алеппо", fact: "Але́ппо, также Ха́леб — крупнейший город Сирии и центр одноимённой мухафазы.", country: "Сирия", population: 2098210, want: true},
		{title: "Калькутта", fact: "Кальку́тта — город в дельте Ганга на востоке Индии, столица штата Западная Бенгалия.", country: "Индия", population: 4631392, want: true},
		{title: "Эль-Гиза", fact: "Эль-Ги́за — город в Арабской Республике Египет, в Верхнем Египте.", country: "Египет", population: 2443203, want: true},
		{title: "Гцгебеха", fact: "Гцгебе́ха — город в ЮАР, в Восточной Капской провинции.", country: "ЮАР", population: 967677, want: true},
		{title: "Манас", fact: "Мана́с — третий по величине город в Кыргызстане, административный центр Джалал-Абадской области.", country: "Кыргызстан", population: 123239, want: true},
		{title: "Ытык-Кюёль", fact: "Ытык-Кюёль — село в Якутии, административный центр Таттинского улуса России.", country: "Россия", population: 6000, want: true},

		{title: "Путешествие", fact: "Путеше́ствие — передвижение по какой-либо территории или акватории.", country: "Бельгия", population: 69554, want: false},
		{title: "Замок Химэдзи", fact: "Замок Химэдзи (яп. 姫路城), также «Замок Белой Цапли» — один из древнейших сохранившихся замков Японии.", country: "Япония", population: 530495, want: false},
		{title: "Никомедия", fact: "Никомедия, уст. Никомидия — древний город в Малой Азии, центр области Вифиния.", country: "Турция", population: 363416, want: false},
		{title: "Виргиния", fact: "Вирги́ния — штат на востоке США.", country: "США", population: 8000, want: false},
		{title: "Метрополитен", fact: "Метрополите́н, ме́тро, скоростной транзи́т, подземная железная дорога.", country: "Филиппины", population: 100000, want: false},
		{title: "Торетам", fact: "Торета́м — посёлок в Кармакшинском районе Кызылординской области Казахстана.", country: "Казахстан", population: 70000, want: false},
	}
	for _, tt := range tests {
		if got := confirms(wikiEntry{title: tt.title, fact: tt.fact}, tt.country, tt.population); got != tt.want {
			t.Errorf("%s: confirms = %v, want %v", tt.title, got, tt.want)
		}
	}
}

func TestMentionsCountryIgnoresGenericWords(t *testing.T) {
	fact := "Алма-Ата — город республиканского значения в Казахстане."
	if mentionsCountry(fact, "Республика Корея") {
		t.Error("«республиканского» must not count as a mention of Республика Корея")
	}
	if !mentionsCountry("Сеул — столица Южной Кореи.", "Республика Корея") {
		t.Error("«Кореи» should match Республика Корея")
	}
}

func TestApplyWiki(t *testing.T) {
	list := []place{
		{id: "aleppo", display: "Халеб", country: "SY"},
		{id: "kyiv", display: "Киев", country: "UA"},
		{id: "tournai", display: "Турне", country: "BE"},
		{id: "tokyo", display: "Токио", country: "JP"},
		{id: "nowiki", display: "Абаза", country: "RU"},
	}
	cache := map[string]wikiEntry{
		"aleppo":  {title: "Алеппо", fact: "Алеппо — крупнейший город Сирии."},
		"kyiv":    {title: "Киев", fact: "Киев — столица Украины."},
		"tournai": {title: "Путешествие", fact: "Путешествие — передвижение по территории."},
		"tokyo":   {title: "Tokyo", fact: "Tokyo — столица Японии."},
	}
	countries := map[string]string{"SY": "Сирия", "UA": "Украина", "BE": "Бельгия", "JP": "Япония", "RU": "Россия"}

	out, renames := applyWiki(list, cache, countries, map[string]string{"kyiv": "Киев"})
	got := map[string]place{}
	for _, p := range out {
		got[p.id] = p
	}

	if p := got["aleppo"]; p.display != "Алеппо" || p.alias != "Халеб" || !p.wiki {
		t.Errorf("aleppo: display=%q alias=%q wiki=%v", p.display, p.alias, p.wiki)
	}
	if p := got["kyiv"]; p.display != "Киев" || p.alias != "" || !p.wiki {
		t.Errorf("kyiv: display=%q alias=%q wiki=%v", p.display, p.alias, p.wiki)
	}
	if p := got["tournai"]; p.display != "Турне" || p.alias != "" || p.wiki {
		t.Errorf("tournai must keep its name when the article is wrong: display=%q wiki=%v", p.display, p.wiki)
	}
	if p := got["tokyo"]; p.display != "Токио" || p.wiki {
		t.Errorf("a Latin title must be ignored, got %q", p.display)
	}
	if p := got["nowiki"]; p.display != "Абаза" || p.wiki {
		t.Errorf("a city with no cache entry must be untouched, got %q", p.display)
	}

	var accepted, rejected int
	for _, r := range renames {
		if r.accepted {
			accepted++
		} else {
			rejected++
		}
	}
	if accepted != 1 || rejected != 1 {
		t.Errorf("renames: %d accepted, %d rejected; want 1 and 1", accepted, rejected)
	}
}

func TestResolvePrefersWikipediaOnCollision(t *testing.T) {
	list := []place{
		{id: "kozhikode", display: "Калькутта", population: 550440},
		{id: "kolkata", display: "Калькутта", alias: "Колката", population: 4631392, wiki: true},
	}
	out, collapsed := resolve(list)
	if len(out) != 1 || collapsed != 1 {
		t.Fatalf("got %d cities, %d collapsed", len(out), collapsed)
	}
	if out[0].id != "kolkata" {
		t.Errorf("kept %s; the Wikipedia-confirmed city must win", out[0].id)
	}
}

func TestResolveDropsAliasThatShadowsAnotherCity(t *testing.T) {
	list := []place{
		{id: "a", display: "Алеппо", alias: "Москва", wiki: true, population: 2},
		{id: "b", display: "Москва", population: 1},
	}
	out, _ := resolve(list)
	for _, p := range out {
		if p.id == "a" && p.alias != "" {
			t.Errorf("alias %q would hijack a real city", p.alias)
		}
	}
}

func TestOverrideWinsOverWikipedia(t *testing.T) {
	list := []place{{id: "1526384", display: "Алматы", country: "KZ"}}
	cache := map[string]wikiEntry{
		"1526384": {title: "Алма-Ата", fact: "Алма-Ата — город республиканского значения в Казахстане."},
	}
	out, _ := applyWiki(list, cache, map[string]string{"KZ": "Казахстан"}, map[string]string{"1526384": "Алматы"})
	if out[0].display != "Алматы" || out[0].alias != "" {
		t.Errorf("override ignored: display=%q alias=%q", out[0].display, out[0].alias)
	}
}
