//go:build !windows

package protect

func resetUnderlaySession() {}

// HasUnderlay is Windows-only; other platforms always re-bind.
func HasUnderlay() bool { return false }

// InvalidateUnderlayCache is a no-op outside Windows.
func InvalidateUnderlayCache() {}
