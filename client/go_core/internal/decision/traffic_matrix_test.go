package decision_test

import (
	"testing"

	"streampass/go_core/internal/decision"
)

// prodLikeRules mirrors a typical published /api/v1/rules payload plus client defaults.
func prodLikeEngine(t *testing.T) *decision.Engine {
	t.Helper()
	raw := `{"version":4,"rules":[
		{"kind":"DOMAIN","pattern":"youtube.com","mode":"RELAY"},
		{"kind":"DOMAIN","pattern":"*.youtube.com","mode":"RELAY"},
		{"kind":"DOMAIN","pattern":"instagram.com","mode":"RELAY"},
		{"kind":"DOMAIN","pattern":"*.instagram.com","mode":"RELAY"},
		{"kind":"DOMAIN","pattern":"facebook.com","mode":"RELAY"},
		{"kind":"DOMAIN","pattern":"twitter.com","mode":"RELAY"},
		{"kind":"DOMAIN","pattern":"x.com","mode":"RELAY"},
		{"kind":"CIDR","pattern":"142.250.0.0/15","mode":"RELAY"}
	]}`
	engine, err := decision.NewEngineFromJSON(raw, "")
	if err != nil {
		t.Fatal(err)
	}
	return engine
}

// TestTrafficMatrix documents expected Decision Engine routing for sites and services.
// Android adds separate layers: RU CIDR excludeRoute, split DNS, app bypass (docs/33).
func TestTrafficMatrix(t *testing.T) {
	engine := prodLikeEngine(t)

	cases := []struct {
		name string
		host string
		want decision.Mode
		note string
	}{
		// RU sites — DIRECT (default rules *.ru + backend)
		{"yandex", "yandex.ru", decision.ModeDirect, "RU search, split DNS Yandex"},
		{"gosuslugi", "gosuslugi.ru", decision.ModeDirect, "RU gov portal"},
		{"sberbank", "online.sberbank.ru", decision.ModeDirect, "RU bank web"},
		{"2ip", "2ip.ru", decision.ModeDirect, "geo IP check should show RU IP on device"},
		{"vk", "vk.com", decision.ModeDirect, "built-in RU social DIRECT"},
		{"mail_ru", "mail.ru", decision.ModeDirect, "RU mail"},
		{"mos_ru", "www.mos.ru", decision.ModeDirect, "built-in *.mos.ru DIRECT"},

		// Foreign acceleration — RELAY
		{"youtube", "www.youtube.com", decision.ModeRelay, "accelerated via Hysteria"},
		{"youtube_root", "youtube.com", decision.ModeRelay, "explicit relay rule"},
		{"instagram", "instagram.com", decision.ModeRelay, "accelerated via relay"},
		{"instagram_cdn", "cdninstagram.com", decision.ModeRelay, "built-in CDN relay fallback"},
		{"instagram_fbcdn", "scontent.cdninstagram.com", decision.ModeRelay, "Instagram media CDN"},
		{"gemini", "gemini.google.com", decision.ModeRelay, "Google AI via relay"},
		{"google", "google.com", decision.ModeRelay, "built-in relay fallback"},
		{"cloudflare", "cloudflare.com", decision.ModeDirect, "DefaultMode=DIRECT (FS §6 / 07.4)"},
		{"linkedin", "www.linkedin.com", decision.ModeRelay, "built-in jobs relay"},
		{"telegram", "web.telegram.org", decision.ModeRelay, "built-in messenger relay"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := engine.Decide(decision.Target{Host: tc.host})
			if got != tc.want {
				t.Fatalf("Decide(%q)=%s want %s (%s)", tc.host, got, tc.want, tc.note)
			}
		})
	}
}

// TestTrafficMatrix_userExclusions verifies user overrides win over relay rules.
func TestTrafficMatrix_userExclusions(t *testing.T) {
	raw := `{"version":1,"rules":[{"kind":"DOMAIN","pattern":"youtube.com","mode":"RELAY"}]}`
	ex := `["youtube.com"]`
	engine, err := decision.NewEngineFromJSON(raw, ex)
	if err != nil {
		t.Fatal(err)
	}
	if got := engine.Decide(decision.Target{Host: "youtube.com"}); got != decision.ModeDirect {
		t.Fatalf("user exclusion = %s want DIRECT", got)
	}
}
