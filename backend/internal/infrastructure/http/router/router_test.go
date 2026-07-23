package router

import "testing"

func TestV1_PrefixesPathKeepsMethod(t *testing.T) {
	cases := map[string]string{
		"GET /rules":           "GET /api/v1/rules",
		"POST /servers":        "POST /api/v1/servers",
		"DELETE /servers/{id}": "DELETE /api/v1/servers/{id}",
		"GET /health":          "GET /api/v1/health",
	}
	for in, want := range cases {
		if got := v1(in); got != want {
			t.Errorf("v1(%q) = %q, want %q", in, got, want)
		}
	}
}
