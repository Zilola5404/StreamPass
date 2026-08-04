package configsvc

import (
	"context"
	"testing"
	"time"

	"streampass/backend/internal/domain/appconfig"
	"streampass/shared/logger"
)

type fakeClock struct{ t time.Time }

func (c fakeClock) Now() time.Time { return c.t }

type fakeRepo struct {
	latest *appconfig.Config
}

func (f *fakeRepo) Latest(ctx context.Context) (*appconfig.Config, error) {
	if f.latest == nil {
		return nil, appconfig.ErrNoConfig
	}
	return f.latest, nil
}

func (f *fakeRepo) Publish(ctx context.Context, c appconfig.Config, publishedAt time.Time) (*appconfig.Config, error) {
	c.Version = 1
	c.UpdatedAt = publishedAt
	f.latest = &c
	return f.latest, nil
}

func validConfig() appconfig.Config {
	return appconfig.Config{
		MinSupportedClientVer: "1.0.0",
		TelemetryEnabled:      true,
		RulePollIntervalSec:   300,
		RelayPollIntervalSec:  60,
	}
}

func TestService_Publish_RejectsMissingClientVersion(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeClock{t: time.Now()}, logger.New("test", "error"))

	cfg := validConfig()
	cfg.MinSupportedClientVer = ""
	if _, err := svc.Publish(context.Background(), cfg); err == nil {
		t.Error("expected error for missing client version, got nil")
	}
}

func TestService_Publish_RejectsNonPositivePollIntervals(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeClock{t: time.Now()}, logger.New("test", "error"))

	cfg := validConfig()
	cfg.RulePollIntervalSec = 0
	if _, err := svc.Publish(context.Background(), cfg); err == nil {
		t.Error("expected error for non-positive rule poll interval, got nil")
	}
}

func TestService_PublishThenGetLatest(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeClock{t: time.Now()}, logger.New("test", "error"))

	published, err := svc.Publish(context.Background(), validConfig())
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
	if latest.MinSupportedClientVer != "1.0.0" {
		t.Errorf("MinSupportedClientVer = %q, want 1.0.0", latest.MinSupportedClientVer)
	}
}

func TestService_Publish_RejectsNonHTTPSDownloadURL(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeClock{t: time.Now()}, logger.New("test", "error"))

	cfg := validConfig()
	cfg.ClientDownloadURL = "http://example.com/app.apk"
	if _, err := svc.Publish(context.Background(), cfg); err == nil {
		t.Error("expected error for non-https download url, got nil")
	}
}

func TestService_Publish_AcceptsHTTPSDownloadURL(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeClock{t: time.Now()}, logger.New("test", "error"))

	cfg := validConfig()
	cfg.LatestClientVersion = "0.1.2"
	cfg.ClientDownloadURL = "https://example.com/StreamPass.apk"
	published, err := svc.Publish(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if published.ClientDownloadURL != cfg.ClientDownloadURL {
		t.Fatalf("download url = %q", published.ClientDownloadURL)
	}
}

func TestService_GetLatest_NoConfigPublishedYet(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo, fakeClock{t: time.Now()}, logger.New("test", "error"))

	if _, err := svc.GetLatest(context.Background()); err == nil {
		t.Error("expected error when no config has been published, got nil")
	}
}
