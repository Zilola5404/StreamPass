package relay_test

import (
	"context"
	"testing"
	"time"

	relaysvc "streampass/backend/internal/application/relay"
	"streampass/backend/internal/domain/relay"
	"streampass/shared/logger"
)

type memRepo struct {
	items []relay.Server
}

func (m *memRepo) List(ctx context.Context) ([]relay.Server, error) {
	out := make([]relay.Server, len(m.items))
	copy(out, m.items)
	return out, nil
}
func (m *memRepo) Register(ctx context.Context, s relay.Server) (*relay.Server, error) {
	s.Region = relay.Region(string(s.Region))
	m.items = append(m.items, s)
	return &s, nil
}
func (m *memRepo) UpdateHealth(ctx context.Context, id relay.ID, healthy bool, load float64, rtt int, at time.Time) error {
	return nil
}
func (m *memRepo) Delete(ctx context.Context, id relay.ID) error { return nil }

func TestListAvailable_FiltersRegionAndRanks(t *testing.T) {
	repo := &memRepo{items: []relay.Server{
		{ID: "nl-heavy", Region: "NL", Healthy: true, LoadRatio: 0.9, RTTMillis: 10},
		{ID: "nl-light", Region: "amsterdam", Healthy: true, LoadRatio: 0.1, RTTMillis: 40},
		{ID: "pl-1", Region: "pl", Healthy: true, LoadRatio: 0.05, RTTMillis: 80},
		{ID: "nl-down", Region: "nl", Healthy: false, LoadRatio: 0, RTTMillis: 1},
	}}
	svc := relaysvc.NewService(repo, logger.New("test", "info"))

	all, err := svc.ListAvailable(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("len=%d want 3", len(all))
	}
	if all[0].ID != "pl-1" {
		t.Fatalf("first=%s want pl-1 (lowest load)", all[0].ID)
	}

	nl, err := svc.ListAvailable(context.Background(), "Netherlands")
	if err != nil {
		t.Fatal(err)
	}
	if len(nl) != 2 {
		t.Fatalf("nl len=%d want 2", len(nl))
	}
	if nl[0].ID != "nl-light" {
		t.Fatalf("best nl=%s want nl-light", nl[0].ID)
	}
	if string(nl[0].Region) != "nl" {
		t.Fatalf("normalized region=%s", nl[0].Region)
	}
}

func TestRegister_NormalizesAndRejectsUnknown(t *testing.T) {
	repo := &memRepo{}
	svc := relaysvc.NewService(repo, logger.New("test", "info"))

	srv, err := svc.Register(context.Background(), "de-1", "Frankfurt", "1.1.1.1", 443, "hysteria2://x", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if string(srv.Region) != "de" {
		t.Fatalf("region=%s want de", srv.Region)
	}

	_, err = svc.Register(context.Background(), "x-1", "mars", "1.1.1.1", 443, "hysteria2://x", time.Now().UTC())
	if err == nil {
		t.Fatal("expected unknown region error")
	}
}
