package decision

// DefaultDirectRules are always prepended before backend rules (ТЗ §6: *.ru, банки, гос).
// Ensures Russian traffic stays DIRECT even when the published rule set is sparse.
var DefaultDirectRules = []Rule{
	{Kind: KindDomain, Pattern: "*.ru", Mode: ModeDirect},
	{Kind: KindDomain, Pattern: "*.su", Mode: ModeDirect},
	{Kind: KindDomain, Pattern: "*.xn--p1ai", Mode: ModeDirect},
	{Kind: KindDomain, Pattern: "*.рф", Mode: ModeDirect},
	{Kind: KindDomain, Pattern: "*.mos.ru", Mode: ModeDirect},
	{Kind: KindDomain, Pattern: "*.gov.ru", Mode: ModeDirect},
	{Kind: KindDomain, Pattern: "*.edu.ru", Mode: ModeDirect},
	{Kind: KindDomain, Pattern: "*.mil.ru", Mode: ModeDirect},
	{Kind: KindDomain, Pattern: "vk.com", Mode: ModeDirect},
	{Kind: KindDomain, Pattern: "*.vk.com", Mode: ModeDirect},
}

// DefaultRelayRules are appended after backend rules when the published set is sparse.
// First-match wins: *.ru DIRECT above always beats these for Russian hosts.
// Covers CDN/AI domains that break when left on DefaultMode DIRECT (Instagram media, Gemini, …).
var DefaultRelayRules = []Rule{
	// Meta / Instagram (main + CDN — required for feed/video like YouTube)
	{Kind: KindDomain, Pattern: "instagram.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.instagram.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "cdninstagram.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.cdninstagram.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "facebook.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.facebook.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "fbcdn.net", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.fbcdn.net", Mode: ModeRelay},

	// YouTube / Google media
	{Kind: KindDomain, Pattern: "youtube.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.youtube.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "youtu.be", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "ytimg.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.ytimg.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "googlevideo.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.googlevideo.com", Mode: ModeRelay},

	// Google AI / Gemini and common AI services
	{Kind: KindDomain, Pattern: "gemini.google.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.gemini.google.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "ai.google.dev", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "aistudio.google.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "bard.google.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "generativelanguage.googleapis.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.googleapis.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "google.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.google.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "chatgpt.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.chatgpt.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "openai.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.openai.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "claude.ai", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.claude.ai", Mode: ModeRelay},

	// Social / jobs / messaging (domain rules — need DNS-in-TUN for hostname)
	{Kind: KindDomain, Pattern: "telegram.org", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.telegram.org", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "t.me", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.t.me", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "telegram.me", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.telegram.me", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "linkedin.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.linkedin.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "licdn.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.licdn.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "upwork.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.upwork.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "indeed.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.indeed.com", Mode: ModeRelay},

	// Social / dev
	{Kind: KindDomain, Pattern: "twitter.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.twitter.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "x.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.x.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "github.com", Mode: ModeRelay},
	{Kind: KindDomain, Pattern: "*.github.com", Mode: ModeRelay},

	// IP-only safety net when HostForIP miss (Chrome/QUIC caches). Domain rules remain
	// primary. No Cloudflare /12 (docs/07.4 §11). Google/Meta ranges = known accelerator CDNs.
	{Kind: KindCIDR, Pattern: "142.250.0.0/15", Mode: ModeRelay},
	{Kind: KindCIDR, Pattern: "172.217.0.0/16", Mode: ModeRelay},
	{Kind: KindCIDR, Pattern: "216.58.0.0/15", Mode: ModeRelay},
	{Kind: KindCIDR, Pattern: "74.125.0.0/16", Mode: ModeRelay},
	{Kind: KindCIDR, Pattern: "108.177.0.0/16", Mode: ModeRelay},
	{Kind: KindCIDR, Pattern: "142.251.0.0/16", Mode: ModeRelay},
	{Kind: KindCIDR, Pattern: "157.240.0.0/16", Mode: ModeRelay},
	{Kind: KindCIDR, Pattern: "31.13.64.0/18", Mode: ModeRelay},
	{Kind: KindCIDR, Pattern: "185.60.216.0/22", Mode: ModeRelay},
	{Kind: KindCIDR, Pattern: "57.144.0.0/16", Mode: ModeRelay},

	// Telegram DC ranges (narrow)
	{Kind: KindCIDR, Pattern: "91.108.4.0/22", Mode: ModeRelay},
	{Kind: KindCIDR, Pattern: "91.108.8.0/22", Mode: ModeRelay},
	{Kind: KindCIDR, Pattern: "91.108.56.0/22", Mode: ModeRelay},
	{Kind: KindCIDR, Pattern: "149.154.160.0/20", Mode: ModeRelay},
}

// MergeWithDefaults prepends DIRECT rules for Russian TLDs and appends built-in RELAY fallbacks.
func MergeWithDefaults(rules []Rule) []Rule {
	out := make([]Rule, 0, len(DefaultDirectRules)+len(rules)+len(DefaultRelayRules))
	out = append(out, DefaultDirectRules...)
	out = append(out, rules...)
	out = append(out, DefaultRelayRules...)
	return out
}
