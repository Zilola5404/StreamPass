package decision_test

import (
  "testing"
  "streampass/go_core/internal/decision"
)

func TestIssue5_S7andGithub(t *testing.T) {
  e := decision.NewEngine(decision.MergeWithDefaults(nil), nil, decision.DefaultMode)
  cases := []struct{ host string; want decision.Mode }{
    {"s7.ru", decision.ModeDirect},
    {"www.s7.ru", decision.ModeDirect},
    {"ya.ru", decision.ModeDirect},
    {"2ip.ru", decision.ModeDirect},
    {"gosuslugi.ru", decision.ModeDirect},
    {"github.com", decision.ModeRelay},
    {"www.youtube.com", decision.ModeRelay},
    {"openai.com", decision.ModeRelay},
    {"example.com", decision.ModeDirect}, // unknown → DIRECT
  }
  for _, c := range cases {
    d := e.Decide(decision.Target{Host: c.host})
    if d != c.want {
      t.Fatalf("%s got %s want %s", c.host, d, c.want)
    }
  }
}
