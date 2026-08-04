package region_test

import (
	"testing"

	"streampass/shared/region"
)

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"NL":           "nl",
		"nl":           "nl",
		"Amsterdam":    "nl",
		"netherlands":  "nl",
		"de":           "de",
		"Frankfurt":    "de",
		"Germany":      "de",
		"pl":           "pl",
		"Warsaw":       "pl",
		"fi":           "fi",
		"Helsinki":     "fi",
		"  DE  ":       "de",
		"custom-east":  "custom-east",
		"":             "",
	}
	for in, want := range cases {
		if got := region.Normalize(in); got != want {
			t.Errorf("Normalize(%q)=%q want %q", in, got, want)
		}
	}
}

func TestIsKnown(t *testing.T) {
	if !region.IsKnown("NL") {
		t.Fatal("NL should be known")
	}
	if region.IsKnown("custom-east") {
		t.Fatal("custom-east should not be known")
	}
}

func TestCatalogHasFourRegions(t *testing.T) {
	if len(region.Catalog) != 4 {
		t.Fatalf("catalog len=%d want 4", len(region.Catalog))
	}
	want := []string{"de", "nl", "pl", "fi"}
	for i, code := range want {
		if region.Catalog[i].Code != code {
			t.Errorf("Catalog[%d].Code=%q want %q", i, region.Catalog[i].Code, code)
		}
	}
}

func TestLabelOf(t *testing.T) {
	if got := region.LabelOf("nl"); got != "Amsterdam (NL)" {
		t.Fatalf("LabelOf(nl)=%q", got)
	}
}
