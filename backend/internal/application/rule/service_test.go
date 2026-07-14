package rule

import (
	"context"
	"testing"
	"time"

	"streampass/backend/internal/domain/rule"
	"streampass/shared/logger"
)

type fakeClock struct{ t time.Time }

func (c fakeClock) Now() time.Time { return c.t }

type fakeRepo struct {
	latest    *rule.Set
	published []rule.Rule
}

func (f *fakeRepo) Latest(ctx context.Context) (*rule.Set, error) {
	if f.latest == nil {
		return nil, rule.ErrNoRuleSet
	}
	return f.latest, nil
}

func (f *fakeRepo) Publish(ctx context.Context, rules []rule.Rule, createdAt time.Time) (*rule.Set, error) {
	f.published = rules
	set := rule.NewSet(1, rules, createdAt)
	f.latest = set
	return set, nil
}

func TestService_Publish_RejectsEmptyRuleSet(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeClock{t: time.Now()}, logger.New("test", "error"))

	if _, err := svc.Publish(context.Background(), nil); err == nil {
		t.Error("expected error for empty rule set, got nil")
	}
}

func TestService_Publish_RejectsInvalidMode(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeClock{t: time.Now()}, logger.New("test", "error"))

	rules := []rule.Rule{{Kind: rule.KindDomain, Pattern: "*.ru", Mode: "BOGUS"}}
	if _, err := svc.Publish(context.Background(), rules); err == nil {
		t.Error("expected error for invalid mode, got nil")
	}
}

func TestService_PublishThenGetLatest(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeClock{t: time.Now()}, logger.New("test", "error"))

	rules := []rule.Rule{{Kind: rule.KindDomain, Pattern: "*.ru", Mode: rule.ModeDirect}}
	published, err := svc.Publish(context.Background(), rules)
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if published.Version != 1 {
		t.Errorf("Version = %d, want 1", published.Version)
	}

	latest, err := svc.GetLatest(context.Background())
	if err != nil {
		t.Fatalf("GetLatest() error = %v", err)
	}
	if len(latest.Rules) != 1 || latest.Rules[0].Pattern != "*.ru" {
		t.Errorf("GetLatest() = %+v, want one rule matching *.ru", latest.Rules)
	}
}
