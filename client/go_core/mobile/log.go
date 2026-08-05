package mobile

import (
	"sync"

	"streampass/go_core/internal/dnscache"
	"streampass/go_core/internal/tunbridge"
)

// EventLogger receives diagnostic lines from the Go core (DNS bootstrap, stop, etc.).
type EventLogger interface {
	Log(message string)
}

var (
	eventLogMu sync.RWMutex
	eventLog   EventLogger
)

// SetEventLogger installs a platform logger (ConnectLogger on Android).
func SetEventLogger(l EventLogger) {
	eventLogMu.Lock()
	eventLog = l
	eventLogMu.Unlock()
	if l == nil {
		dnscache.SetLogger(nil)
		tunbridge.SetLogger(nil)
		return
	}
	fn := func(message string) {
		l.Log(message)
	}
	dnscache.SetLogger(fn)
	tunbridge.SetLogger(fn)
}

func logEvent(message string) {
	eventLogMu.RLock()
	l := eventLog
	eventLogMu.RUnlock()
	if l != nil {
		l.Log(message)
	}
}
