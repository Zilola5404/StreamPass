package dnscache

import "strings"

// IsRussianDomain reports whether a hostname should use a local RU resolver
// (Yandex) instead of Cloudflare DoH — ТЗ: Russian resources stay local.
// PreferredDNSRoute returns the primary resolver label before network I/O.
// Russian TLDs use Yandex (77.88.8.8); foreign domains use Cloudflare DoH.
func PreferredDNSRoute(name string) string {
	if IsRussianDomain(name) {
		return "yandex"
	}
	return "doh"
}

func IsRussianDomain(name string) bool {
	host := strings.ToLower(strings.TrimSpace(name))
	host = strings.TrimSuffix(host, ".")
	if host == "" {
		return false
	}
	switch {
	case host == "ru" || strings.HasSuffix(host, ".ru"):
		return true
	case host == "su" || strings.HasSuffix(host, ".su"):
		return true
	case host == "xn--p1ai" || strings.HasSuffix(host, ".xn--p1ai"):
		return true
	case strings.HasSuffix(host, ".рф"):
		return true
	default:
		return false
	}
}
